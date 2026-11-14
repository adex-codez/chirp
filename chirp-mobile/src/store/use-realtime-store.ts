import { create } from "zustand";

import {
  isConnectionReady,
  type EventEnvelope,
  type RealtimeStatus,
} from "@/lib/realtime/types";

type RealtimeState = {
  status: RealtimeStatus;
  connectionId: string | null;
  lastEvent: EventEnvelope | null;
  setStatus: (status: RealtimeStatus) => void;
  receive: (event: EventEnvelope) => void;
  reset: () => void;
};

export const useRealtimeStore = create<RealtimeState>((set) => ({
  status: "offline",
  connectionId: null,
  lastEvent: null,
  setStatus: (status) => set({ status }),
  receive: (event) =>
    set((state) => ({
      lastEvent: event,
      connectionId: isConnectionReady(event)
        ? event.data.connection_id
        : state.connectionId,
    })),
  reset: () => set({ status: "offline", connectionId: null, lastEvent: null }),
}));
