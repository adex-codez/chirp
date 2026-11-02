import * as AppleAuthentication from "expo-apple-authentication";
import {
  GoogleSignin,
  isSuccessResponse,
} from "@react-native-google-signin/google-signin";
import { Button } from "@/components/button";
import { useEffect, useState } from "react";
import { View } from "react-native";

import { googleConfigured, newAppleNonce } from "@/lib/social";
import { useSessionStore } from "@/store/use-session-store";
import { GoogleMark } from "@/components/provider-marks";

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
    if (!googleConfigured()) {
      useSessionStore.setState({
        error: "Google sign-in is not configured yet.",
      });
      return;
    }
    try {
      await GoogleSignin.hasPlayServices();
      const response = await GoogleSignin.signIn();
      if (isSuccessResponse(response)) {
        // Google hands back a user without a token only when no web client
        // ID was configured, which googleConfigured() already prevents.
        if (!response.data.idToken) {
          throw new Error("Google sign-in returned no identity token.");
        }
        await socialSignIn("google", response.data.idToken);
      }
    } catch {
      // Cancellation resolves as a result; every rejection is a real failure.
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
