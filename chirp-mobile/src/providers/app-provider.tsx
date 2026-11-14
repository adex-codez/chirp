import { QueryClientProvider } from "@tanstack/react-query";
import { HeroUINativeProvider } from "heroui-native/provider";
import { GestureHandlerRootView } from "react-native-gesture-handler";

import { queryClient } from "@/lib/query-client";
import { RealtimeProvider } from "@/providers/realtime-provider";

type AppProviderProps = {
  children: React.ReactNode;
};

export function AppProvider({ children }: AppProviderProps) {
  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <QueryClientProvider client={queryClient}>
        <RealtimeProvider>
          <HeroUINativeProvider>{children}</HeroUINativeProvider>
        </RealtimeProvider>
      </QueryClientProvider>
    </GestureHandlerRootView>
  );
}
