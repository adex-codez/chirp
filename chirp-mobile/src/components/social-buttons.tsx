import * as AppleAuthentication from "expo-apple-authentication";
import {
  ResponseType,
  makeRedirectUri,
  useAuthRequest,
} from "expo-auth-session";
import { Button } from "heroui-native/button";
import { Typography } from "heroui-native/text";
import { useEffect, useState } from "react";
import { View } from "react-native";

import {
  googleClientId,
  googleConfigured,
  googleDiscovery,
  newAppleNonce,
} from "@/lib/social";
import { useSessionStore } from "@/store/use-session-store";
import { AppleMark, GoogleMark } from "@/components/provider-marks";

/**
 * Social sign-in buttons, shared by the sign-in and sign-up screens.
 * Apple uses the platform control where available (Apple requires it on
 * iOS); elsewhere it falls back to a standard button that explains
 * Apple sign-in is not available on that device. Google uses a standard
 * button with its brand mark that opens the provider in a browser.
 */
export function SocialButtons() {
  const socialSignIn = useSessionStore((state) => state.socialSignIn);
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const [appleAvailable, setAppleAvailable] = useState(false);

  useEffect(() => {
    void AppleAuthentication.isAvailableAsync().then(setAppleAvailable);
  }, []);

  const redirectUri = makeRedirectUri({ scheme: "chirpmobile" });
  const [googleRequest, googleResponse, promptGoogle] = useAuthRequest(
    {
      clientId: googleClientId(),
      redirectUri,
      scopes: ["openid", "profile", "email"],
      responseType: ResponseType.IdToken,
    },
    googleDiscovery,
  );

  useEffect(() => {
    if (googleResponse?.type === "success") {
      const idToken = googleResponse.params.id_token;
      if (typeof idToken === "string" && idToken !== "") {
        void socialSignIn("google", idToken);
      }
    } else if (googleResponse?.type === "error") {
      useSessionStore.setState({
        error: "Google sign-in failed. Try again.",
        isBusy: false,
      });
    }
  }, [googleResponse, socialSignIn]);

  const signInWithApple = async () => {
    if (!appleAvailable) {
      useSessionStore.setState({
        error: "Apple sign-in is not available on this device. Use Google or your Email instead.",
      });
      return;
    }
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
    if (!googleConfigured()) {
      useSessionStore.setState({
        error: "Google sign-in is not configured yet.",
      });
      return;
    }
    try {
      await promptGoogle();
    } catch {
      // Failures surface through the response effect or store error.
    }
  };

  return (
    <View className="gap-3">
      <View className="flex-row gap-3">
        <View className="flex-1">
          {appleAvailable ? (
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
          ) : (
            <Button
              variant="primary"
              onPress={() => void signInWithApple()}
              isDisabled={isBusy}
              accessibilityLabel="Continue with Apple"
            >
              <View className="flex-row items-center justify-center gap-2">
                <AppleMark />
                <Button.Label>Apple</Button.Label>
              </View>
            </Button>
          )}
        </View>
        <View className="flex-1">
          <Button
            variant="outline"
            onPress={() => void signInWithGoogle()}
            isDisabled={isBusy || !googleRequest}
            accessibilityLabel="Continue with Google"
          >
            <View className="flex-row items-center justify-center gap-2">
              <GoogleMark />
              <Button.Label>Google</Button.Label>
            </View>
          </Button>
        </View>
      </View>
      {error ? (
        <Typography.Paragraph className="text-danger">
          {error}
        </Typography.Paragraph>
      ) : null}
    </View>
  );
}
