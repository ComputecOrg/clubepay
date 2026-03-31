const SESSION_COOKIE = "assinapix_session";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  const match = document.cookie.match(/(?:^|;\s*)assinapix_session=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

export function setToken(token: string): void {
  document.cookie = `${SESSION_COOKIE}=${encodeURIComponent(token)}; path=/; max-age=${60 * 60 * 24}; SameSite=Lax`;
}

export function clearToken(): void {
  document.cookie = `${SESSION_COOKIE}=; path=/; max-age=0`;
}
