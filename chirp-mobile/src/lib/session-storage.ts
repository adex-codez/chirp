import * as SecureStore from "expo-secure-store";

import type { SessionUser } from "@/store/use-session-store";

const SESSION_KEY = "chirp.session";

export type PersistedSession = {
  user: SessionUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  pendingEmail: string | null;
  pendingToken: string | null;
};

/**
 * Platform secure storage for session data. Tokens never touch
 * unencrypted storage, logs, or analytics.
 */
export async function saveSession(session: PersistedSession): Promise<void> {
  await SecureStore.setItemAsync(SESSION_KEY, JSON.stringify(session));
}

export async function loadSession(): Promise<PersistedSession | null> {
  const raw = await SecureStore.getItemAsync(SESSION_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as PersistedSession;
  } catch {
    await SecureStore.deleteItemAsync(SESSION_KEY);
    return null;
  }
}

export async function clearSession(): Promise<void> {
  await SecureStore.deleteItemAsync(SESSION_KEY);
}
