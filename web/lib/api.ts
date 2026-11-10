import axios, { type AxiosInstance } from "axios";

const defaultApiBaseUrl = "http://localhost:8081/api";

export const API_BASE_URL = (
  process.env.NEXT_PUBLIC_API_URL || defaultApiBaseUrl
).replace(/\/+$/, "");

export function apiUrl(path: string): string {
  return `${API_BASE_URL}/${path.replace(/^\/+/, "")}`;
}

export function mediaUrl(path: string): string {
  if (/^https?:\/\//i.test(path)) {
    return path;
  }

  const apiOrigin = new URL(API_BASE_URL).origin;
  return new URL(path, `${apiOrigin}/`).toString();
}

const SESSION_KEYS = ["token", "userRole", "userId", "userName", "userEmail"];
const LOGIN_PATH = "/login";

export function getToken(): string | null {
  try {
    return localStorage.getItem("token");
  } catch {
    return null;
  }
}

export function clearSession(): void {
  try {
    SESSION_KEYS.forEach((key) => localStorage.removeItem(key));
  } catch {
    // Storage can be unavailable (private mode); nothing to clear then.
  }
}

function isApiRequest(url?: string): boolean {
  return Boolean(url) && url!.startsWith(`${API_BASE_URL}/`);
}

// Login and register must never carry a token and must not trigger sign-out
// when the credentials are wrong.
function isAuthEndpoint(url?: string): boolean {
  return Boolean(url) && url!.startsWith(`${API_BASE_URL}/auth/`);
}

function hasRealBearerToken(value: unknown): boolean {
  if (typeof value !== "string") return false;
  const token = value.replace(/^Bearer\s+/i, "").trim();
  return token !== "" && token !== "null" && token !== "undefined";
}

type SessionOptions = {
  getToken: () => string | null;
  onSessionExpired: () => void;
};

/**
 * Attaches the stored token to every API request and signs the user out
 * when the API rejects a token that was sent (expired or invalid).
 * Exported separately so it can be tested without a browser.
 */
export function installSessionHandling(
  client: AxiosInstance,
  options: SessionOptions,
): () => void {
  const requestId = client.interceptors.request.use((config) => {
    if (!isApiRequest(config.url) || isAuthEndpoint(config.url)) {
      return config;
    }
    if (!hasRealBearerToken(config.headers.get("Authorization"))) {
      const token = options.getToken();
      if (token) {
        config.headers.set("Authorization", `Bearer ${token}`);
      }
    }
    return config;
  });

  const responseId = client.interceptors.response.use(
    (response) => response,
    (error) => {
      const config = error?.config;
      if (
        error?.response?.status === 401 &&
        config &&
        isApiRequest(config.url) &&
        !isAuthEndpoint(config.url) &&
        hasRealBearerToken(config.headers?.get?.("Authorization"))
      ) {
        options.onSessionExpired();
      }
      return Promise.reject(error);
    },
  );

  return () => {
    client.interceptors.request.eject(requestId);
    client.interceptors.response.eject(responseId);
  };
}

function signOutAndRedirect(): void {
  clearSession();
  if (typeof window !== "undefined" && window.location.pathname !== LOGIN_PATH) {
    window.location.assign(LOGIN_PATH);
  }
}

// One axios instance for the whole app. Token attachment and sign-out on 401
// are installed on this instance only, so pages must call apiClient and not
// the global axios.
export const apiClient = axios.create();

installSessionHandling(apiClient, {
  getToken,
  onSessionExpired: signOutAndRedirect,
});
