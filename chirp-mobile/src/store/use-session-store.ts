import { create } from "zustand";

import { apiGet, apiPost, configureAuth } from "@/lib/api";
import {
  clearSession,
  loadSession,
  saveSession,
} from "@/lib/session-storage";

export type SessionUser = {
  id: string;
  username: string;
  email: string;
  verified: boolean;
};

type AuthTokens = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
};

type AuthResponse = {
  user: SessionUser;
} & AuthTokens;

type SignUpResponse = {
  user: SessionUser;
  verification_sent: boolean;
  dev_verification_code?: string;
};

type SessionStatus =
  | "restoring"
  | "guest"
  | "pending-verification"
  | "authenticated";

type SessionState = {
  status: SessionStatus;
  user: SessionUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  pendingEmail: string | null;
  error: string | null;
  isBusy: boolean;
  restore: () => Promise<void>;
  signUp: (username: string, email: string, password: string) => Promise<void>;
  verify: (code: string) => Promise<void>;
  resendCode: () => Promise<void>;
  signIn: (email: string, password: string) => Promise<void>;
  refreshTokens: () => Promise<boolean>;
  signOut: () => Promise<void>;
  signOutEverywhere: () => Promise<void>;
  dropToGuest: () => Promise<void>;
  clearError: () => void;
};

function messageOf(error: unknown): string {
  if (error instanceof Error) return error.message;
  return "Something went wrong. Try again.";
}

type StoreSetter = (
  partial: Partial<
    Pick<
      SessionState,
      | "isBusy"
      | "error"
      | "status"
      | "user"
      | "accessToken"
      | "refreshToken"
      | "pendingEmail"
    >
  >,
) => void;

async function runRequest<T>(
  set: StoreSetter,
  path: string,
  body: unknown,
): Promise<T> {
  set({ isBusy: true, error: null });
  try {
    return await apiPost<T>(path, body);
  } catch (error) {
    set({ error: messageOf(error), isBusy: false });
    throw error;
  }
}

function decodeBase64Url(input: string): string {
  const alphabet =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_";
  const values = new Map(
    [...alphabet].map((char, index) => [char, index] as const),
  );
  const bytes: number[] = [];
  let buffer = 0;
  let bits = 0;
  for (const char of input) {
    const value = values.get(char);
    if (value === undefined) throw new Error("bad encoding");
    buffer = (buffer << 6) | value;
    bits += 6;
    if (bits >= 8) {
      bits -= 8;
      bytes.push((buffer >> bits) & 0xff);
    }
  }
  return String.fromCharCode(...bytes);
}

/**
 * Client-side expiry peek. The server stays the source of truth; this only
 * decides whether restore should renew proactively.
 */
function accessExpiringSoon(token: string, marginMs = 60_000): boolean {
  try {
    const payload = JSON.parse(decodeBase64Url(token.split(".")[1])) as {
      exp?: number;
    };
    if (typeof payload.exp !== "number") return true;
    return payload.exp * 1000 - Date.now() < marginMs;
  } catch {
    return true;
  }
}

export const useSessionStore = create<SessionState>((set, get) => {
  const persist = () => {
    const { user, accessToken, refreshToken, pendingEmail } = get();
    return saveSession({ user, accessToken, refreshToken, pendingEmail });
  };

  const applySession = (data: AuthResponse) => {
    set({
      status: "authenticated",
      user: data.user,
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      pendingEmail: null,
      isBusy: false,
    });
    return persist();
  };

  return {
    status: "restoring",
    user: null,
    accessToken: null,
    refreshToken: null,
    pendingEmail: null,
    error: null,
    isBusy: false,

    restore: async () => {
      const saved = await loadSession();
      if (!saved?.refreshToken || !saved.accessToken) {
        set({
          status: saved?.pendingEmail ? "pending-verification" : "guest",
          user: saved?.user ?? null,
          pendingEmail: saved?.pendingEmail ?? null,
        });
        return;
      }
      // Cold start: confirm with the server so revoked sessions (including
      // sign-out-everywhere from another device) drop to guest immediately.
      try {
        if (accessExpiringSoon(saved.accessToken)) {
          throw new Error("expired");
        }
        const user = await apiGet<SessionUser>("/auth/me", saved.accessToken);
        set({
          status: "authenticated",
          user,
          accessToken: saved.accessToken,
          refreshToken: saved.refreshToken,
          pendingEmail: null,
        });
      } catch {
        const renewed = await get().refreshTokens();
        if (!renewed) {
          await get().dropToGuest();
        }
      }
    },

    signUp: async (username, email, password) => {
      const data = await runRequest<SignUpResponse>(set, "/auth/signup", {
        username,
        email,
        password,
      });
      set({
        status: "pending-verification",
        user: data.user,
        pendingEmail: data.user.email,
        isBusy: false,
      });
      await persist();
    },

    verify: async (code) => {
      const { pendingEmail } = get();
      if (!pendingEmail) {
        set({ error: "Start by signing up first." });
        return;
      }
      const data = await runRequest<AuthResponse>(set, "/auth/verify", {
        email: pendingEmail,
        code,
      });
      await applySession(data);
    },

    resendCode: async () => {
      const { pendingEmail } = get();
      if (!pendingEmail) {
        set({ error: "Start by signing up first." });
        return;
      }
      await runRequest(set, "/auth/verify/resend", { email: pendingEmail });
      set({ isBusy: false });
    },

    signIn: async (email, password) => {
      let data: AuthResponse;
      try {
        data = await runRequest<AuthResponse>(set, "/auth/login", {
          email,
          password,
        });
      } catch (error) {
        const message = messageOf(error);
        set({ error: message, isBusy: false });
        // An unverified sign-in parks the person at verification.
        if (message.toLowerCase().includes("not verified")) {
          set({ status: "pending-verification", pendingEmail: email });
          await persist();
        }
        throw error;
      }
      await applySession(data);
    },

    refreshTokens: async () => {
      const { refreshToken } = get();
      if (!refreshToken) return false;
      try {
        const data = await apiPost<AuthResponse>("/auth/refresh", {
          refresh_token: refreshToken,
        });
        await applySession(data);
        return true;
      } catch {
        await get().dropToGuest();
        return false;
      }
    },

    signOut: async () => {
      const { refreshToken } = get();
      if (refreshToken) {
        try {
          await apiPost("/auth/logout", { refresh_token: refreshToken });
        } catch {
          // Best effort: the local session is dropped regardless.
        }
      }
      await get().dropToGuest();
    },

    signOutEverywhere: async () => {
      const { accessToken } = get();
      if (accessToken) {
        try {
          await apiPost("/auth/logout-all", {}, accessToken);
        } catch {
          // Best effort: the local session is dropped regardless.
        }
      }
      await get().dropToGuest();
    },

    dropToGuest: async () => {
      await clearSession();
      set({
        status: "guest",
        user: null,
        accessToken: null,
        refreshToken: null,
        pendingEmail: null,
        error: null,
        isBusy: false,
      });
    },

    clearError: () => set({ error: null }),
  };
});

configureAuth({
  getAccessToken: () => useSessionStore.getState().accessToken,
  getRefreshToken: () => useSessionStore.getState().refreshToken,
  refreshAccessToken: async () => {
    const renewed = await useSessionStore.getState().refreshTokens();
    return renewed ? useSessionStore.getState().accessToken : null;
  },
  onAuthFailed: () => {
    void useSessionStore.getState().dropToGuest();
  },
});
