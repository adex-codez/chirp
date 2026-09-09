import * as Crypto from "expo-crypto";
import { Platform } from "react-native";

/**
 * Social sign-in configuration. The IDs below come from outside the repo:
 * Google OAuth client IDs (Cloud console) and the Apple audience (the
 * app's bundle ID, matching the backend's APPLE_AUDIENCE). Until they are
 * set, the buttons report misconfiguration instead of failing obscurely.
 */
export const googleClientIds = {
  ios: process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID ?? "",
  android: process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID ?? "",
  web: process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID ?? "",
};

export function googleClientId(): string {
  return (
    Platform.select({
      ios: googleClientIds.ios,
      android: googleClientIds.android,
      default: googleClientIds.web,
    }) ?? ""
  );
}

export const googleDiscovery = {
  authorizationEndpoint: "https://accounts.google.com/o/oauth2/v2/auth",
};

export function googleConfigured(): boolean {
  return googleClientId() !== "";
}

/** Raw Apple nonce plus its sha256: the hash travels in the sign-in request,
 * the raw value goes to the backend, which compares hashes. */
export async function newAppleNonce(): Promise<{ raw: string; hashed: string }> {
  const bytes = await Crypto.getRandomBytesAsync(32);
  const raw = [...bytes]
    .map((byte) => byte.toString(16).padStart(2, "0"))
    .join("");
  const hashed = await Crypto.digestStringAsync(
    Crypto.CryptoDigestAlgorithm.SHA256,
    raw,
  );
  return { raw, hashed };
}
