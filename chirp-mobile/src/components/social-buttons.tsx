import * as AppleAuthentication from "expo-apple-authentication";
import {
  GoogleSignin,
  isSuccessResponse,
} from "@react-native-google-signin/google-signin";
import { Button } from "@/components/button";
import { useEffect, useState } from "react";
import { Platform, View } from "react-native";

import { ApiError } from "@/lib/api";
import { googleClientIds, googleConfigured, newAppleNonce } from "@/lib/social";
import { useSessionStore } from "@/store/use-session-store";
import { GoogleMark } from "@/components/provider-marks";

type GoogleAsyncStage = "play-services" | "native-sign-in" | "session";
type GoogleFailureStage = GoogleAsyncStage | "identity-token";
type ObservedSessionStatus = ReturnType<typeof useSessionStore.getState>["status"];

type GoogleErrorSummary = {
  name: string | null;
  message: string | null;
  code: string | number | null;
} & (
  | { kind: "api"; status: ApiError["status"]; requestId: string | null }
  | { kind: "exception" }
);

type GoogleDiagnostic =
  | {
      stage: "configuration";
      outcome: "ready" | "blocked";
      iosClientIdPresent: boolean;
      webClientIdPresent: boolean;
    }
  | { stage: GoogleAsyncStage; outcome: "started" }
  | { stage: "play-services"; outcome: "resolved"; available: boolean }
  | {
      stage: "native-sign-in";
      outcome: "success";
      responseType: "success";
      hasIdToken: boolean;
    }
  | {
      stage: "native-sign-in";
      outcome: "cancelled";
      responseType: "cancelled";
      hasIdToken: false;
    }
  | {
      stage: "session";
      outcome: "completed";
      observedSessionStatus: ObservedSessionStatus;
    }
  | {
      stage: GoogleFailureStage;
      outcome: "failed";
      error: GoogleErrorSummary;
      observedSessionStatus: ObservedSessionStatus;
    };

function redactGoogleDiagnostic(text: string): string {
  let redacted = text
    .replace(
      /(\b(?:[a-z_-]*tokens?|[a-z_-]*credentials?|password|client[_ -]?secret|nonce|authorization|(?:server[_ -]?)?(?:auth|authorization)[_ -]?code)["']?\s*[:=]\s*)(?:"[^"]*"|'[^']*'|(?:Bearer|Basic)\s+[^\s,;}\]]+|[^\s,;}\]]+)/gi,
      "$1[redacted]",
    )
    .replace(/[A-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[A-Z0-9.-]+/gi, "[redacted-email]")
    .replace(/\b(Bearer|Basic)\s+[A-Za-z0-9._~+/-]+=*/gi, "$1 [redacted]")
    .replace(
      /\beyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b/g,
      "[redacted-token]",
    )
    .replace(
      /\b(?:[A-Za-z0-9_-]+\.apps\.googleusercontent\.com|com\.googleusercontent\.apps\.[A-Za-z0-9_-]+)\b/g,
      "[redacted-client-id]",
    );
  for (const clientId of Object.values(googleClientIds)) {
    if (clientId !== "") {
      redacted = redacted.split(clientId).join("[redacted-client-id]");
    }
  }
  return redacted;
}

function summarizeGoogleError(error: unknown): GoogleErrorSummary {
  const fields = typeof error === "object" && error !== null ? error : {};
  const name = "name" in fields ? fields.name : null;
  const message =
    typeof error === "string"
      ? error
      : "message" in fields
        ? fields.message
        : null;
  const code = "code" in fields ? fields.code : null;
  const summary = {
    name: typeof name === "string" ? redactGoogleDiagnostic(name) : null,
    message: typeof message === "string" ? redactGoogleDiagnostic(message) : null,
    code:
      typeof code === "string"
        ? redactGoogleDiagnostic(code)
        : typeof code === "number"
          ? code
          : null,
  };
  return error instanceof ApiError
    ? {
        ...summary,
        kind: "api",
        status: error.status,
        requestId:
          error.requestId === undefined
            ? null
            : redactGoogleDiagnostic(error.requestId),
      }
    : { ...summary, kind: "exception" };
}

/**
 * Social sign-in buttons, shared by the sign-in and sign-up screens.
 * Apple renders as the platform control, and only where the platform
 * offers it. Google uses a standard button with its brand mark that
 * runs the native sign-in sheet. Store errors render on the screen,
 * which owns the single error slot.
 */
export function SocialButtons() {
  const socialSignIn = useSessionStore((state) => state.socialSignIn);
  const isBusy = useSessionStore((state) => state.isBusy);
  const [appleAvailable, setAppleAvailable] = useState(false);

  useEffect(() => {
    void AppleAuthentication.isAvailableAsync().then(setAppleAvailable);
  }, []);

  const signInWithApple = async () => {
    try {
      const nonce = await newAppleNonce();
      const credential = await AppleAuthentication.signInAsync({
        requestedScopes: [
          AppleAuthentication.AppleAuthenticationScope.EMAIL,
        ],
        nonce: nonce.hashed,
      });
      if (credential.identityToken) {
        await socialSignIn("apple", credential.identityToken, nonce.raw);
      }
    } catch (e: unknown) {
      // Cancellation stays silent; real failures surface through the store.
      const code =
        typeof e === "object" && e !== null && "code" in e
          ? String((e as { code: unknown }).code)
          : "";
      if (code !== "ERR_REQUEST_CANCELED") {
        useSessionStore.setState({
          error: "Apple sign-in failed. Try again.",
        });
      }
    }
  };

  const signInWithGoogle = async () => {
    const startedAt = Date.now();
    const attemptId = `${startedAt.toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
    const log = (event: GoogleDiagnostic) => {
      if (__DEV__) {
        console.info("[auth/google]", {
          attemptId,
          platform: Platform.OS,
          elapsedMs: Date.now() - startedAt,
          ...event,
        });
      }
    };
    const configured = googleConfigured();
    log({
      stage: "configuration",
      outcome: configured ? "ready" : "blocked",
      iosClientIdPresent: googleClientIds.ios !== "",
      webClientIdPresent: googleClientIds.web !== "",
    });
    if (!configured) {
      useSessionStore.setState({
        error: "Google sign-in is not configured yet.",
      });
      return;
    }
    let stage: GoogleFailureStage = "play-services";
    try {
      log({ stage, outcome: "started" });
      const available = await GoogleSignin.hasPlayServices();
      log({ stage, outcome: "resolved", available });

      stage = "native-sign-in";
      log({ stage, outcome: "started" });
      const response = await GoogleSignin.signIn();
      if (isSuccessResponse(response)) {
        log({
          stage,
          outcome: "success",
          responseType: response.type,
          hasIdToken: Boolean(response.data.idToken),
        });
        stage = "identity-token";
        if (!response.data.idToken) {
          throw new Error("Google sign-in returned no identity token.");
        }

        stage = "session";
        log({ stage, outcome: "started" });
        await socialSignIn("google", response.data.idToken);
        log({
          stage,
          outcome: "completed",
          observedSessionStatus: useSessionStore.getState().status,
        });
      } else {
        log({
          stage,
          outcome: "cancelled",
          responseType: response.type,
          hasIdToken: false,
        });
      }
    } catch (error: unknown) {
      if (__DEV__) {
        log({
          stage,
          outcome: "failed",
          error: summarizeGoogleError(error),
          observedSessionStatus: useSessionStore.getState().status,
        });
      }
      useSessionStore.setState({
        error: "Google sign-in failed. Try again.",
      });
    }
  };

  return (
    <View className="flex-row gap-3">
      {appleAvailable ? (
        <View className="flex-1">
          <AppleAuthentication.AppleAuthenticationButton
            buttonType={
              AppleAuthentication.AppleAuthenticationButtonType.SIGN_IN
            }
            buttonStyle={AppleAuthentication.AppleAuthenticationButtonStyle.BLACK}
            cornerRadius={8}
            // One-off: 44pt is Apple's minimum touch target; the button is
            // Apple's own control, so its geometry stays in platform terms.
            style={{ height: 44 }}
            onPress={() => void signInWithApple()}
          />
        </View>
      ) : null}
      <View className="flex-1">
        <Button
          variant="outline"
          onPress={() => void signInWithGoogle()}
          isDisabled={isBusy}
          accessibilityLabel="Continue with Google"
        >
          <View className="flex-row items-center justify-center gap-2">
            <GoogleMark />
            <Button.Label>Google</Button.Label>
          </View>
        </Button>
      </View>
    </View>
  );
}
