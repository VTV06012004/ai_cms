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
  refreshAccessToken?: () => Promise<string>;
  setToken?: (token: string) => void;
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
  let refreshPromise: Promise<string> | null = null;
  let sessionExpired = false;
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
    async (error: unknown) => {
      if (!axios.isAxiosError(error)) {
        return Promise.reject(error);
      }
      const config = error.config;
      if (
        error.response?.status === 401 &&
        config &&
        isApiRequest(config.url) &&
        !isAuthEndpoint(config.url) &&
        hasRealBearerToken(config.headers.get("Authorization"))
      ) {
        const retryConfig = config as typeof config & { _sessionRetry?: boolean };
        if (retryConfig._sessionRetry || !options.refreshAccessToken) {
          options.onSessionExpired();
          return Promise.reject(error);
        }

        retryConfig._sessionRetry = true;
        refreshPromise ??= options.refreshAccessToken().finally(() => {
          refreshPromise = null;
        });
        try {
          const token = await refreshPromise;
          sessionExpired = false;
          options.setToken?.(token);
          retryConfig.headers.set("Authorization", `Bearer ${token}`);
          return client.request(retryConfig);
        } catch {
          if (!sessionExpired) {
            sessionExpired = true;
            options.onSessionExpired();
          }
          return Promise.reject(error);
        }
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
export const apiClient = axios.create({ withCredentials: true });

const refreshClient = axios.create({ withCredentials: true });

async function refreshAccessToken(): Promise<string> {
  const response = await refreshClient.post(apiUrl('auth/refresh'));
  const token: unknown = response.data?.token;
  if (typeof token !== 'string' || token === '') {
    throw new Error('Refresh endpoint returned no access token');
  }
  const user: unknown = response.data?.user;
  if (typeof user === 'object' && user !== null) {
    const refreshedUser = user as Record<string, unknown>;
    if (typeof refreshedUser.role === 'string') {
      localStorage.setItem('userRole', refreshedUser.role);
    }
    if (typeof refreshedUser.name === 'string') {
      localStorage.setItem('userName', refreshedUser.name);
    }
    if (typeof refreshedUser.email === 'string') {
      localStorage.setItem('userEmail', refreshedUser.email);
    }
    if (typeof refreshedUser.id === 'number') {
      localStorage.setItem('userId', String(refreshedUser.id));
    }
  }
  return token;
}

export async function logout(): Promise<void> {
  try {
    await apiClient.post(apiUrl('auth/logout'));
  } catch (error) {
    console.error('Không thể thu hồi phiên đăng nhập trên máy chủ:', error);
  } finally {
    clearSession();
  }
}

installSessionHandling(apiClient, {
  getToken,
  onSessionExpired: signOutAndRedirect,
  refreshAccessToken,
  setToken: (token) => localStorage.setItem('token', token),
});
