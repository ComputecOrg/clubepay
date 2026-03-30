export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("clubepay_token");
}

export function setToken(token: string): void {
  localStorage.setItem("clubepay_token", token);
  document.cookie = `clubepay_token=${token}; path=/; max-age=${60 * 60 * 24 * 7}; SameSite=Lax`;
}

export function clearToken(): void {
  localStorage.removeItem("clubepay_token");
  document.cookie = "clubepay_token=; path=/; max-age=0; path=/";
}
