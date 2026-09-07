import "../global.css";

import { Stack } from "expo-router";

import { AppProvider } from "@/providers/app-provider";

export default function RootLayout() {
  return (
    <AppProvider>
      <Stack />
    </AppProvider>
  );
}
