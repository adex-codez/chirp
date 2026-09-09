import "../global.css";

import { StatusBar } from "expo-status-bar";
import { SplashScreen, Stack, router, useSegments } from "expo-router";
import { useEffect } from "react";

import { AppProvider } from "@/providers/app-provider";
import { useSessionStore } from "@/store/use-session-store";

void SplashScreen.preventAutoHideAsync();

// Leaf names of the public sign-in group. (Expo strips the `(auth)` group
// from segments, so the guard matches leaves. Extend when the group gains
// screens such as the social Username picker.)
const AUTH_SEGMENTS = new Set(["sign-in", "sign-up", "verify"]);

function RouteGuard() {
  const status = useSessionStore((state) => state.status);
  const restore = useSessionStore((state) => state.restore);
  const segments = useSegments();

  useEffect(() => {
    void restore();
  }, [restore]);

  useEffect(() => {
    if (status === "restoring") return;
    const first = segments[0];
    const inAuthGroup = first !== undefined && AUTH_SEGMENTS.has(first);
    if (status === "authenticated" && inAuthGroup) {
      router.replace("/");
    } else if (status === "pending-verification" && !inAuthGroup) {
      router.replace("/verify");
    } else if (status === "guest" && !inAuthGroup) {
      router.replace("/sign-in");
    }
  }, [status, segments]);

  useEffect(() => {
    if (status !== "restoring") {
      void SplashScreen.hideAsync();
    }
  }, [status]);

  return <Stack />;
}

export default function RootLayout() {
  return (
    <AppProvider>
      <StatusBar style="dark" />
      <RouteGuard />
    </AppProvider>
  );
}
