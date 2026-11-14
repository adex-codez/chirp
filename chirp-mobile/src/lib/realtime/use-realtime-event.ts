import { useEffect } from "react";

import type { EventEnvelope } from "@/lib/realtime/types";
import { useRealtimeStore } from "@/store/use-realtime-store";

// Runs the handler for each new server event. Screens opt in; the provider
// owns the socket.
export function useRealtimeEvent(
  handler: (event: EventEnvelope) => void,
): void {
  useEffect(
    () =>
      useRealtimeStore.subscribe((state, previous) => {
        if (state.lastEvent && state.lastEvent !== previous.lastEvent) {
          handler(state.lastEvent);
        }
      }),
    [handler],
  );
}
