import { create } from "zustand";

import {
  isConnectionReady,
  type EventEnvelope,
  type RealtimeStatus,
} from "@/lib/realtime/types";

const listeners = new Set<(event: EventEnvelope) => void>();

type RealtimeState = {
  status: RealtimeStatus;
  connectionId: string | null;
  setStatus: (status: RealtimeStatus) => void;
  receive: (event: EventEnvelope) => void;
  subscribe: (handler: (event: EventEnvelope) => void) => () => void;
  reset: () => void;
};

export const useRealtimeStore = create<RealtimeState>((set) => ({
  status: "offline",
  connectionId: null,
  setStatus: (status) => set({ status }),
  receive: (event) => {
    if (isConnectionReady(event)) {
      set({ connectionId: event.data.connection_id });
    }
    for (const listener of listeners) listener(event);
  },
  subscribe: (handler) => {
    listeners.add(handler);
    return () => {
      listeners.delete(handler);
    };
  },
  reset: () => set({ status: "offline", connectionId: null }),
}));
