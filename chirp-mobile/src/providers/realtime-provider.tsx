import { useEffect, type ReactNode } from "react";

import { RealtimeClient } from "@/lib/realtime/client";
import { useRealtimeStore } from "@/store/use-realtime-store";
import { useSessionStore } from "@/store/use-session-store";

export function RealtimeProvider({ children }: { children: ReactNode }) {
  const status = useSessionStore((state) => state.status);

  useEffect(() => {
    if (status !== "authenticated") {
      useRealtimeStore.getState().reset();
      return;
    }

    const client = new RealtimeClient({
      getAccessToken: () => useSessionStore.getState().accessToken,
      refreshAccessToken: async () => {
        try {
          const refreshed = await useSessionStore.getState().refreshTokens();
          return refreshed ? useSessionStore.getState().accessToken : null;
        } catch {
          return null;
        }
      },
      onEvent: (event) => useRealtimeStore.getState().receive(event),
      onStatus: (next) => useRealtimeStore.getState().setStatus(next),
      onSessionEnded: () => {
        void useSessionStore.getState().dropToGuest();
      },
    });
    client.start();

    return () => {
      client.stop();
      useRealtimeStore.getState().reset();
    };
  }, [status]);

  return <>{children}</>;
}
