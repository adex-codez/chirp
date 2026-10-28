import "../global.css";

import { SplashScreen, Stack, router, useSegments } from "expo-router";
import { useEffect } from "react";

import { AppProvider } from "@/providers/app-provider";
import { useSessionStore } from "@/store/use-session-store";

void SplashScreen.preventAutoHideAsync();

// Public screens inside the sign-in group. Route groups stay in `segments`,
// so the leaf is the last entry. Extend when the group gains screens.
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
  // screen rather than anywhere else in the sign-in group. Guests that land
  // outside the public group would render an empty shell, so they go to
  // sign-in instead.
  useEffect(() => {
    const screen = segments[segments.length - 1];
    if (status === "pending-verification" && screen !== "verify") {
      router.replace("/verify");
    } else if (status === "needs-username" && screen !== "choose-username") {
      router.replace("/choose-username");
    } else if (
      status === "guest" &&
      (screen === undefined || !AUTH_LEAVES.has(screen))
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
