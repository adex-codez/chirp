import "../global.css";

import { StatusBar } from "expo-status-bar";
import { SplashScreen, Stack, router, useSegments } from "expo-router";
import { useEffect } from "react";

import { AppProvider } from "@/providers/app-provider";
import { useSessionStore } from "@/store/use-session-store";

void SplashScreen.preventAutoHideAsync();

// Leaf names of the public sign-in group (group name itself is stripped
// from segments). Extend when the group gains screens.
const AUTH_LEAVES = new Set([
  "sign-in",
  "sign-up",
  "verify",
  "choose-username",
  "forgot-password",
  "reset-password",
]);

export default function RootLayout() {
  const status = useSessionStore((state) => state.status);
  const restore = useSessionStore((state) => state.restore);
  const segments = useSegments();

  useEffect(() => {
    void restore();
  }, [restore]);

  // Pending Users have unfinished setup, so they belong on their setup
  // screen rather than anywhere else in the sign-in group. useSegments()
  // strips `(auth)`, so segments[0] is already the leaf (e.g. "sign-in").
  // Guests landing on `/` (no index route) match no child screen, which
  // renders as a blank group shell — send them to sign-in explicitly.
  useEffect(() => {
    const leaf = segments[0];
    if (status === "pending-verification" && leaf !== "verify") {
      router.replace("/verify");
    } else if (status === "needs-username" && leaf !== "choose-username") {
      router.replace("/choose-username");
    } else if (
      status === "guest" &&
      (leaf === undefined || !AUTH_LEAVES.has(leaf))
    ) {
      router.replace("/sign-in");
    }
  }, [status, segments]);

  useEffect(() => {
    if (status !== "restoring") {
      void SplashScreen.hideAsync();
    }
  }, [status]);

  // Hold the splash until the session is restored: with no group
  // accessible yet, there is nothing to render.
  if (status === "restoring") {
    return null;
  }

  return (
    <AppProvider>
      <StatusBar style="dark" />
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Screen name="index" options={{ headerShown: false }} />
        <Stack.Protected guard={status !== "authenticated"}>
          <Stack.Screen name="(auth)" options={{ headerShown: false }} />
        </Stack.Protected>
        <Stack.Protected guard={status === "authenticated"}>
          <Stack.Screen name="(app)" options={{ headerShown: false }} />
        </Stack.Protected>
      </Stack>
    </AppProvider>
  );
}
