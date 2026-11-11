import * as SecureStore from "expo-secure-store";

import { parsePersistedSessionUser, type SessionUser } from "@/lib/session-user";

const SESSION_KEY = "chirp.session";
let sessionWrite = Promise.resolve();

export type PersistedSession = {
  user: SessionUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  pendingEmail: string | null;
  pendingToken: string | null;
};

function writeSession(write: () => Promise<void>): Promise<void> {
  const pending = sessionWrite.then(write);
  sessionWrite = pending.catch(() => {});
  return pending;
}

/**
 * Platform secure storage for session data. Tokens never touch
 * unencrypted storage, logs, or analytics.
 */
export async function saveSession(session: PersistedSession): Promise<void> {
  const raw = JSON.stringify(session);
  await writeSession(() => SecureStore.setItemAsync(SESSION_KEY, raw));
}

export async function loadSession(): Promise<PersistedSession | null> {
  const raw = await SecureStore.getItemAsync(SESSION_KEY);
  if (!raw) return null;
  try {
    const value: unknown = JSON.parse(raw);
    if (typeof value !== "object" || value === null || Array.isArray(value)) {
      throw new Error("Invalid saved session.");
    }
    const user = "user" in value ? value.user : null;
    return {
      user: user == null ? null : parsePersistedSessionUser(user),
      accessToken: storedString("accessToken" in value ? value.accessToken : null),
      refreshToken: storedString("refreshToken" in value ? value.refreshToken : null),
      pendingEmail: storedString("pendingEmail" in value ? value.pendingEmail : null),
      pendingToken: storedString("pendingToken" in value ? value.pendingToken : null),
    };
  } catch {
    await clearSession();
    return null;
  }
}

function storedString(value: unknown): string | null {
  if (value == null) return null;
  if (typeof value !== "string") throw new Error("Invalid saved session.");
  return value;
}

export async function clearSession(): Promise<void> {
  await writeSession(() => SecureStore.deleteItemAsync(SESSION_KEY));
}
