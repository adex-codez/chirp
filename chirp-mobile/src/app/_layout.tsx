import "../global.css";

import { StatusBar } from "expo-status-bar";
import { SplashScreen, Stack, router, useSegments } from "expo-router";
import { useEffect } from "react";

import { AppProvider } from "@/providers/app-provider";
import { useSessionStore } from "@/store/use-session-store";

void SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const status = useSessionStore((state) => state.status);
  const restore = useSessionStore((state) => state.restore);
  const segments = useSegments();

  useEffect(() => {
    void restore();
  }, [restore]);

  // Pending Users have unfinished setup, so they belong on their setup
  // screen rather than anywhere else in the sign-in group.
  useEffect(() => {
    if (status === "pending-verification" && segments[0] !== "verify") {
      router.replace("/verify");
    } else if (status === "needs-username" && segments[0] !== "choose-username") {
      router.replace("/choose-username");
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
      <Stack>
        <Stack.Protected guard={status !== "authenticated"}>
          <Stack.Screen name="(auth)" />
        </Stack.Protected>
        <Stack.Protected guard={status === "authenticated"}>
          <Stack.Screen name="(app)" />
        </Stack.Protected>
      </Stack>
    </AppProvider>
  );
}
