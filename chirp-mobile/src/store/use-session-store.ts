import { create } from "zustand";

import { ApiError, apiFetch, configureAuth, singleFlightRefresh } from "@/lib/api";
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
  hasPassword: boolean;
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

type SocialResponse = {
  user: SessionUser;
  pending: boolean;
  pending_token?: string;
  access_token?: string;
  refresh_token?: string;
  expires_in?: number;
};

type SessionStatus =
  | "restoring"
  | "guest"
  | "pending-verification"
  | "needs-username"
  | "authenticated";

type SessionState = {
  status: SessionStatus;
  user: SessionUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  pendingEmail: string | null;
  pendingToken: string | null;
  error: string | null;
  isBusy: boolean;
  restore: () => Promise<void>;
  signUp: (username: string, email: string, password: string) => Promise<void>;
  verify: (code: string) => Promise<void>;
  resendCode: () => Promise<void>;
  signIn: (email: string, password: string) => Promise<void>;
  socialSignIn: (
    provider: "apple" | "google",
    idToken: string,
    nonce?: string,
  ) => Promise<void>;
  setUsername: (username: string) => Promise<void>;
  cancelUsernamePick: () => Promise<void>;
  forgotPassword: (email: string) => Promise<void>;
  resetPassword: (
    email: string,
    code: string,
    newPassword: string,
  ) => Promise<void>;
  changeUsername: (username: string) => Promise<void>;
  addPassword: (newPassword: string) => Promise<void>;
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
      | "pendingToken"
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
    return await apiFetch<T>(path, { method: "POST", body });
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
    const { user, accessToken, refreshToken, pendingEmail, pendingToken } =
      get();
    return saveSession({
      user,
      accessToken,
      refreshToken,
      pendingEmail,
      pendingToken,
    });
  };

  const applySession = (data: AuthResponse) => {
    set({
      status: "authenticated",
      user: data.user,
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      pendingEmail: null,
      pendingToken: null,
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
    pendingToken: null,
    error: null,
    isBusy: false,

    restore: async () => {
      const saved = await loadSession();
      // Sessions persisted before hasPassword existed normalize to true;
      // the server confirms the real value on the next /auth/me call.
      const savedUser = saved?.user
        ? { ...saved.user, hasPassword: saved.user.hasPassword ?? true }
        : null;
      if (!saved?.refreshToken || !saved.accessToken) {
        // A pending Username grant survives restarts while it is unexpired.
        if (
          saved?.pendingToken &&
          savedUser &&
          !accessExpiringSoon(saved.pendingToken)
        ) {
          set({
            status: "needs-username",
            user: savedUser,
            pendingToken: saved.pendingToken,
          });
          return;
        }
        set({
          status: saved?.pendingEmail ? "pending-verification" : "guest",
          user: savedUser,
          pendingEmail: saved?.pendingEmail ?? null,
          pendingToken: null,
        });
        return;
      }
      // Cold start: confirm with the server so revoked sessions (including
      // sign-out-everywhere from another device) drop to guest immediately.
      try {
        if (accessExpiringSoon(saved.accessToken)) {
          throw new Error("expired");
        }
        const user = await apiFetch<SessionUser>("/auth/me", {
          accessToken: saved.accessToken,
        });
        set({
          status: "authenticated",
          user,
          accessToken: saved.accessToken,
          refreshToken: saved.refreshToken,
          pendingEmail: null,
        });
      } catch {
        try {
          if (await get().refreshTokens()) return;
          return; // Definitive rejection already dropped to guest.
        } catch {
          // Transient failure (offline): trust the saved tokens only while
          // access is unexpired and a User is present; the interceptor
          // retries once connectivity returns.
          if (savedUser && !accessExpiringSoon(saved.accessToken)) {
            set({
              status: "authenticated",
              user: savedUser,
              accessToken: saved.accessToken,
              refreshToken: saved.refreshToken,
              pendingEmail: null,
            });
            return;
          }
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
        pendingToken: null,
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

    socialSignIn: async (provider, idToken, nonce) => {
      let data: SocialResponse;
      try {
        data = await runRequest<SocialResponse>(set, "/auth/social", {
          provider,
          id_token: idToken,
          ...(nonce ? { nonce } : {}),
        });
      } catch (error) {
        const message = messageOf(error);
        set({ error: message, isBusy: false });
        throw error;
      }
      if (data.pending && data.pending_token) {
        set({
          status: "needs-username",
          user: data.user,
          accessToken: null,
          refreshToken: null,
          pendingEmail: null,
          pendingToken: data.pending_token,
          isBusy: false,
        });
        await persist();
        return;
      }
      await applySession(data as AuthResponse);
    },

    setUsername: async (username) => {
      const { pendingToken } = get();
      if (!pendingToken) {
        set({ error: "Start with Apple or Google sign-in first." });
        return;
      }
      set({ isBusy: true, error: null });
      try {
        const data = await apiFetch<AuthResponse>("/auth/username", {
          method: "POST",
          body: { username },
          accessToken: pendingToken,
        });
        await applySession(data);
      } catch (error) {
        set({ error: messageOf(error), isBusy: false });
        throw error;
      }
    },

    cancelUsernamePick: async () => {
      // The pending grant is simply discarded; the server row stays
      // Username-less until its owner returns through Social sign-in.
      await get().dropToGuest();
    },

    forgotPassword: async (email) => {
      // Neutral by design: success reveals nothing about whether the
      // address exists. Transport failures (rate limits) still surface.
      await runRequest(set, "/auth/password/forgot", { email });
      set({ isBusy: false });
    },

    resetPassword: async (email, code, newPassword) => {
      await runRequest(set, "/auth/password/reset", {
        email,
        code,
        new_password: newPassword,
      });
      set({ isBusy: false });
      // A reset revokes every session, including this device's: sign in
      // again with the new password.
      await get().dropToGuest();
    },

    changeUsername: async (username) => {
      const { accessToken } = get();
      if (!accessToken) {
        set({ error: "You are not signed in." });
        throw new Error("You are not signed in.");
      }
      set({ isBusy: true, error: null });
      try {
        const user = await apiFetch<SessionUser>("/auth/username/change", {
          method: "POST",
          body: { username },
          accessToken,
        });
        set({ user, isBusy: false });
        await persist();
        // Renew so the access claims carry the new Username immediately.
        await get()
          .refreshTokens()
          .catch(() => false);
      } catch (error) {
        set({ error: messageOf(error), isBusy: false });
        throw error;
      }
    },

    addPassword: async (newPassword) => {
      const { accessToken } = get();
      if (!accessToken) {
        set({ error: "You are not signed in." });
        throw new Error("You are not signed in.");
      }
      set({ isBusy: true, error: null });
      try {
        await apiFetch("/auth/password", {
          method: "POST",
          body: { new_password: newPassword },
          accessToken,
        });
        const user = await apiFetch<SessionUser>("/auth/me", {
          accessToken,
        });
        set({ user, isBusy: false });
        await persist();
      } catch (error) {
        set({ error: messageOf(error), isBusy: false });
        throw error;
      }
    },

    refreshTokens: async () => {
      // Every renewal funnels through the single-flight gate, so restore()
      // and concurrent apiAuth retries never mint two refreshes at once.
      // Definitive rejection drops to guest here; transient failures throw
      // and the session is kept.
      const access = await singleFlightRefresh();
      if (!access) {
        await get().dropToGuest();
        return false;
      }
      return true;
    },

    signOut: async () => {
      const { refreshToken } = get();
      if (refreshToken) {
        try {
          await apiFetch("/auth/logout", {
            method: "POST",
            body: { refresh_token: refreshToken },
          });
        } catch {
          // Best effort: the local session is dropped regardless.
        }
      }
      await get().dropToGuest();
    },

    signOutEverywhere: async () => {
      try {
        // Renew first so an expired access token does not silently void the
        // revocation while the local session is dropped.
        let token = get().accessToken;
        if (token && accessExpiringSoon(token)) {
          await get()
            .refreshTokens()
            .catch(() => false);
          token = get().accessToken;
        }
        if (token) {
          await apiFetch("/auth/logout-all", {
            method: "POST",
            body: {},
            accessToken: token,
          });
        }
      } catch {
        // Best effort: the local session is dropped regardless.
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
        pendingToken: null,
        error: null,
        isBusy: false,
      });
    },

    clearError: () => set({ error: null }),
  };
});

configureAuth({
  getAccessToken: () => useSessionStore.getState().accessToken,
  performRefresh: async () => {
    const { refreshToken } = useSessionStore.getState();
    if (!refreshToken) return null;
    try {
      const data = await apiFetch<AuthResponse>("/auth/refresh", {
        method: "POST",
        body: { refresh_token: refreshToken },
      });
      useSessionStore.setState({
        status: "authenticated",
        user: data.user,
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
        pendingEmail: null,
        isBusy: false,
      });
      const applied = useSessionStore.getState();
      await saveSession({
        user: applied.user,
        accessToken: applied.accessToken,
        refreshToken: applied.refreshToken,
        pendingEmail: applied.pendingEmail,
        pendingToken: applied.pendingToken,
      });
      return data.access_token;
    } catch (error) {
      // Definitive rejection ends the session; transient failure keeps it.
      if (
        error instanceof ApiError &&
        (error.status === 401 || error.status === 404)
      ) {
        return null;
      }
      throw error;
    }
  },
  onAuthFailed: () => {
    void useSessionStore.getState().dropToGuest();
  },
});
