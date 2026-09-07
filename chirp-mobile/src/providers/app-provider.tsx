import { QueryClientProvider } from "@tanstack/react-query";
import { HeroUINativeProvider } from "heroui-native/provider";
import { GestureHandlerRootView } from "react-native-gesture-handler";

import { queryClient } from "@/lib/query-client";

type AppProviderProps = {
  children: React.ReactNode;
};

export function AppProvider({ children }: AppProviderProps) {
  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <QueryClientProvider client={queryClient}>
        <HeroUINativeProvider>{children}</HeroUINativeProvider>
      </QueryClientProvider>
    </GestureHandlerRootView>
  );
}
