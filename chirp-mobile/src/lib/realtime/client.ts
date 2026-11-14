import { Platform } from "react-native";

import { apiBaseUrl } from "@/lib/api";

import {
  parseServerEvent,
  type EventEnvelope,
  type RealtimeStatus,
} from "./types";

type RealtimeHandlers = {
  getAccessToken: () => string | null;
  refreshAccessToken: () => Promise<string | null>;
  onEvent: (event: EventEnvelope) => void;
  onStatus: (status: RealtimeStatus) => void;
  onSessionEnded: () => void;
};

const MAX_BACKOFF_MS = 30_000;
const HANDSHAKE_PROTOCOL = "chirp.v1";
const TOKEN_PROTOCOL_PREFIX = "bearer.";

type NativeWebSocket = new (
  url: string,
  protocols?: string | string[],
  options?: { headers?: Record<string, string> },
) => WebSocket;

function websocketUrl(): string {
  const explicit = process.env.EXPO_PUBLIC_WS_URL?.trim();
  if (explicit) return explicit.replace(/\/$/, "");
  return `${apiBaseUrl.replace(/^http/, "ws")}/ws`;
}

// One socket per signed-in app session. Reconnects with jittered exponential
// backoff; an expired access token is refreshed once through the existing
// HTTPS flow before the next attempt.
export class RealtimeClient {
  private socket: WebSocket | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private attempts = 0;
  private stopped = true;
  private generation = 0;

  constructor(private readonly handlers: RealtimeHandlers) {}

  start(): void {
    if (!this.stopped) return;
    this.stopped = false;
    this.attempts = 0;
    this.connect();
  }

  stop(): void {
    this.stopped = true;
    this.generation += 1;
    this.clearTimer();
    this.closeSocket(1000, "client stopped");
    this.handlers.onStatus("offline");
  }

  private connect(): void {
    if (this.stopped) return;
    const token = this.handlers.getAccessToken();
    if (!token) {
      this.handlers.onStatus("offline");
      return;
    }
    const generation = this.generation;
    this.handlers.onStatus("connecting");

    let socket: WebSocket;
    if (Platform.OS === "web") {
      // Browsers cannot set handshake headers, so the token rides as a
      // subprotocol entry. It never goes in the URL.
      socket = new WebSocket(websocketUrl(), [
        HANDSHAKE_PROTOCOL,
        `${TOKEN_PROTOCOL_PREFIX}${token}`,
      ]);
    } else {
      const NativeSocket = WebSocket as unknown as NativeWebSocket;
      socket = new NativeSocket(websocketUrl(), [HANDSHAKE_PROTOCOL], {
        headers: { Authorization: `Bearer ${token}` },
      });
    }
    this.socket = socket;

    socket.onopen = () => {
      if (generation !== this.generation) return;
      this.attempts = 0;
      this.handlers.onStatus("online");
    };
    socket.onmessage = (message) => {
      if (generation !== this.generation) return;
      if (typeof message.data !== "string") return;
      const event = parseServerEvent(message.data);
      if (event) this.handlers.onEvent(event);
    };
    socket.onerror = () => undefined;
    socket.onclose = (event) => {
      if (this.socket === socket) this.socket = null;
      if (generation !== this.generation) return;
      void this.handleClose(event.code);
    };
  }

  private async handleClose(code: number): Promise<void> {
    if (this.stopped) return;
    if (code === 4003) {
      this.stopped = true;
      this.handlers.onStatus("offline");
      this.handlers.onSessionEnded();
      return;
    }
    if (code === 4001) {
      this.handlers.onStatus("backoff");
      const refreshed = await this.handlers
        .refreshAccessToken()
        .catch(() => null);
      if (this.stopped) return;
      if (!refreshed) {
        this.stopped = true;
        this.handlers.onStatus("offline");
        return;
      }
      this.attempts = 0;
      this.connect();
      return;
    }
    if (code === 1000) {
      this.stopped = true;
      this.handlers.onStatus("offline");
      return;
    }
    this.scheduleReconnect();
  }

  private scheduleReconnect(): void {
    if (this.stopped) return;
    this.handlers.onStatus("backoff");
    const ceiling = Math.min(MAX_BACKOFF_MS, 1000 * 2 ** this.attempts);
    this.attempts += 1;
    const delay = Math.floor(ceiling * (0.5 + Math.random() * 0.5));
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  private clearTimer(): void {
    if (!this.reconnectTimer) return;
    clearTimeout(this.reconnectTimer);
    this.reconnectTimer = null;
  }

  private closeSocket(code: number, reason: string): void {
    const socket = this.socket;
    this.socket = null;
    if (!socket) return;
    socket.onopen = null;
    socket.onmessage = null;
    socket.onerror = null;
    socket.onclose = null;
    try {
      socket.close(code, reason);
    } catch {
      // The transport is already closing.
    }
  }
}
