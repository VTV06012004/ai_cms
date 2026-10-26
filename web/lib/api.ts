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
