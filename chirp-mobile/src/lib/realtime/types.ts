// Wire format for server-pushed events. Mirrors server/internal/event.
export const EVENT_VERSION = 1;

export type RealtimeStatus = "offline" | "connecting" | "online" | "backoff";

export type EventEnvelope<Data = unknown> = {
  v: number;
  id: string;
  type: string;
  topic: string;
  ts: string;
  data: Data;
};

export type ConnectionReady = {
  connection_id: string;
  username: string;
  heartbeat_seconds: number;
  access_expires_at: string;
};

export type SessionEndedReason =
  | "expired"
  | "revoked"
  | "signed_out"
  | "server_shutdown";

export type SessionEnded = {
  reason: SessionEndedReason;
};

const SESSION_ENDED_REASONS: ReadonlySet<string> = new Set([
  "expired",
  "revoked",
  "signed_out",
  "server_shutdown",
]);

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

// Returns null for malformed frames and unknown envelope versions, so a newer
// server cannot make this client misread an old shape.
export function parseServerEvent(raw: string): EventEnvelope | null {
  let value: unknown;
  try {
    value = JSON.parse(raw);
  } catch {
    return null;
  }
  if (!isRecord(value)) return null;
  if (value.v !== EVENT_VERSION) return null;
  if (typeof value.id !== "string" || value.id === "") return null;
  if (typeof value.type !== "string" || value.type === "") return null;
  if (typeof value.ts !== "string") return null;
  return {
    v: EVENT_VERSION,
    id: value.id,
    type: value.type,
    topic: typeof value.topic === "string" ? value.topic : "",
    ts: value.ts,
    data: value.data,
  };
}

export function isConnectionReady(
  event: EventEnvelope,
): event is EventEnvelope<ConnectionReady> {
  if (event.type !== "connection.ready") return false;
  const data = event.data;
  return (
    isRecord(data) &&
    typeof data.connection_id === "string" &&
    typeof data.username === "string" &&
    typeof data.heartbeat_seconds === "number" &&
    typeof data.access_expires_at === "string"
  );
}

export function isSessionEnded(
  event: EventEnvelope,
): event is EventEnvelope<SessionEnded> {
  if (event.type !== "session.ended") return false;
  const data = event.data;
  return (
    isRecord(data) &&
    typeof data.reason === "string" &&
    SESSION_ENDED_REASONS.has(data.reason)
  );
}
