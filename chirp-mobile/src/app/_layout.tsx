import "../global.css";

import { StatusBar } from "expo-status-bar";
import { Stack } from "expo-router";

import { AppProvider } from "@/providers/app-provider";

export default function RootLayout() {
  return (
    <AppProvider>
      <StatusBar style="dark" />
      <Stack />
    </AppProvider>
  );
}
