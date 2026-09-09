const baseUrl =
  process.env.EXPO_PUBLIC_API_URL?.replace(/\/$/, "") ||
  "http://localhost:8080";

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

type ApiEnvelope<T> = {
  success: boolean;
  data?: T;
  error?: string;
};

export async function apiPost<T>(
  path: string,
  body: unknown,
  accessToken?: string,
): Promise<T> {
  const res = await fetch(`${baseUrl}${path}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
    },
    body: JSON.stringify(body),
  });
  const envelope = (await res.json()) as ApiEnvelope<T>;
  if (!res.ok || !envelope.success) {
    throw new ApiError(
      res.status,
      envelope.error ?? "Something went wrong. Try again.",
    );
  }
  return envelope.data as T;
}

export async function apiGet<T>(path: string, accessToken: string): Promise<T> {
  const res = await fetch(`${baseUrl}${path}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  const envelope = (await res.json()) as ApiEnvelope<T>;
  if (!res.ok || !envelope.success) {
    throw new ApiError(
      res.status,
      envelope.error ?? "Something went wrong. Try again.",
    );
  }
  return envelope.data as T;
}

type AuthHooks = {
  getAccessToken: () => string | null;
  /**
   * Renews the pair. Resolves null on definitive rejection (caller drops the
   * session); rejects on transient failure (caller keeps the session).
   */
  performRefresh: () => Promise<string | null>;
  onAuthFailed: () => void;
};

let hooks: AuthHooks | null = null;
// Single in-flight refresh: every renewal path funnels through this gate,
// so concurrent calls queue behind one renewal.
let refreshPromise: Promise<string | null> | null = null;

/** Wired once by the session store; keeps this module free of store imports. */
export function configureAuth(next: AuthHooks): void {
  hooks = next;
  refreshPromise = null;
}

export function singleFlightRefresh(): Promise<string | null> {
  if (!hooks) return Promise.resolve(null);
  if (!refreshPromise) {
    refreshPromise = hooks.performRefresh().finally(() => {
      refreshPromise = null;
    });
  }
  return refreshPromise;
}

/**
 * Authenticated request. Attaches the access token, and on a 401 renews it
 * once through the single-flight refresh before retrying. If renewal fails,
 * the session is dropped to guest.
 */
export async function apiAuth<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  if (!hooks) throw new ApiError(401, "You are not signed in.");
  const accessToken = hooks.getAccessToken();
  const res = await fetch(`${baseUrl}${path}`, {
    ...init,
    headers: {
      ...(init?.headers ?? {}),
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
    },
  });
  if (res.status !== 401) {
    const envelope = (await res.json()) as ApiEnvelope<T>;
    if (!res.ok || !envelope.success) {
      throw new ApiError(
        res.status,
        envelope.error ?? "Something went wrong. Try again.",
      );
    }
    return envelope.data as T;
  }
  const renewed = await singleFlightRefresh();
  if (!renewed) {
    hooks.onAuthFailed();
    throw new ApiError(401, "Your session expired. Sign in again.");
  }
  const retry = await fetch(`${baseUrl}${path}`, {
    ...init,
    headers: {
      ...(init?.headers ?? {}),
      Authorization: `Bearer ${renewed}`,
    },
  });
  const envelope = (await retry.json()) as ApiEnvelope<T>;
  if (!retry.ok || !envelope.success) {
    if (retry.status === 401) hooks.onAuthFailed();
    throw new ApiError(
      retry.status,
      envelope.error ?? "Something went wrong. Try again.",
    );
  }
  return envelope.data as T;
}
