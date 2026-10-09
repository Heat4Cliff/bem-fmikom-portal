import type { AuthUser } from "@/types/auth";

export type Envelope<T> = {
  success: boolean;
  message: string;
  data: T;
  code?: string;
  meta?: { page: number; limit: number; total: number; total_pages: number };
};

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

let accessToken: string | null = null;
let refreshInFlight: Promise<string> | null = null;
let onAuthenticationExpired: (() => void) | null = null;

export const getMemoryAccessToken = () => accessToken;
export const setMemoryAccessToken = (value: string | null) => {
  accessToken = value;
};
export const onAuthExpired = (listener: (() => void) | null) => {
  onAuthenticationExpired = listener;
};

const baseURL = (process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1").replace(/\/$/, "");

async function readEnvelope<T>(response: Response): Promise<Envelope<T>> {
  if (response.status === 204) {
    return { success: true, message: "", data: null as T };
  }
  let envelope: Envelope<T>;
  try {
    envelope = (await response.json()) as Envelope<T>;
  } catch {
    throw new ApiError("Respons server tidak dapat dibaca", response.status);
  }
  if (!response.ok || !envelope.success) {
    throw new ApiError(envelope.message || "Permintaan gagal", response.status, envelope.code);
  }
  return envelope;
}

export function refreshAccessToken(): Promise<string> {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = (async () => {
    const response = await fetch(`${baseURL}/auth/refresh`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
    });
    const envelope = await readEnvelope<{ access_token: string }>(response);
    if (!envelope.data?.access_token) throw new ApiError("Token akses tidak tersedia", response.status);
    setMemoryAccessToken(envelope.data.access_token);
    return envelope.data.access_token;
  })()
    .catch((error: unknown) => {
      setMemoryAccessToken(null);
      onAuthenticationExpired?.();
      throw error;
    })
    .finally(() => {
      refreshInFlight = null;
    });
  return refreshInFlight;
}

export interface RequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown;
  retryOnUnauthorized?: boolean;
}

export async function apiRequestEnvelope<T>(
  path: string,
  options: RequestOptions = {},
): Promise<Envelope<T>> {
  const { retryOnUnauthorized = true, body, headers: providedHeaders, ...init } = options;
  const headers = new Headers(providedHeaders);
  if (body !== undefined && !(body instanceof FormData)) headers.set("Content-Type", "application/json");

  const perform = async (retried: boolean): Promise<Envelope<T>> => {
    const token = getMemoryAccessToken();
    const requestHeaders = new Headers(headers);
    if (token) requestHeaders.set("Authorization", `Bearer ${token}`);
    const response = await fetch(`${baseURL}${path}`, {
      ...init,
      body: body === undefined ? undefined : body instanceof FormData ? body : JSON.stringify(body),
      credentials: "include",
      headers: requestHeaders,
    });
    if (response.status === 401 && retryOnUnauthorized && token && !retried) {
      try {
        await refreshAccessToken();
      } catch {
        throw new ApiError("Sesi telah berakhir", 401, "UNAUTHENTICATED");
      }
      return perform(true);
    }
    return readEnvelope<T>(response);
  };

  return perform(false);
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const envelope = await apiRequestEnvelope<T>(path, options);
  return envelope.data;
}

export interface AuthPayload {
  access_token: string;
  expires_in: number;
  user: AuthUser;
}

export async function loginRequest(email: string, password: string): Promise<AuthPayload> {
  return apiRequest<AuthPayload>("/auth/login", {
    method: "POST",
    body: { email, password },
    retryOnUnauthorized: false,
  });
}

export async function getCurrentUser(): Promise<AuthUser> {
  return apiRequest<AuthUser>("/auth/me");
}

export async function logoutRequest(): Promise<void> {
  await apiRequest<null>("/auth/logout", { method: "POST", retryOnUnauthorized: false });
}
