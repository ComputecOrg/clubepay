export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("clubepay_token");
}

export function setToken(token: string): void {
  localStorage.setItem("clubepay_token", token);
}

export function clearToken(): void {
  localStorage.removeItem("clubepay_token");
}
