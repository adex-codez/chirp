import { create } from "zustand";

import { apiPost } from "@/lib/api";

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

type SessionState = {
  status: "guest" | "pending-verification" | "authenticated";
  user: SessionUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  pendingEmail: string | null;
  error: string | null;
  isBusy: boolean;
  signUp: (username: string, email: string, password: string) => Promise<void>;
  verify: (code: string) => Promise<void>;
  resendCode: () => Promise<void>;
  signIn: (email: string, password: string) => Promise<void>;
  signOut: () => void;
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
      "isBusy" | "error" | "status" | "user" | "accessToken" | "refreshToken" | "pendingEmail"
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

/**
 * In-memory session for now. Tokens move to platform secure storage with the
 * refresh flow in the sessions ticket; nothing here persists across restarts.
 */
export const useSessionStore = create<SessionState>((set, get) => ({
  status: "guest",
  user: null,
  accessToken: null,
  refreshToken: null,
  pendingEmail: null,
  error: null,
  isBusy: false,

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
    set({
      status: "authenticated",
      user: data.user,
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      pendingEmail: null,
      isBusy: false,
    });
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
      }
      throw error;
    }
    set({
      status: "authenticated",
      user: data.user,
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      pendingEmail: null,
      isBusy: false,
    });
  },

  signOut: () =>
    set({
      status: "guest",
      user: null,
      accessToken: null,
      refreshToken: null,
      pendingEmail: null,
      error: null,
    }),

  clearError: () => set({ error: null }),
}));
