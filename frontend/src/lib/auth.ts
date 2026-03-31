// assinapix_session: indicador de sessão não-HttpOnly (valor "1").
// O JWT real está em assinapix_token (HttpOnly, gerenciado pelo backend via Set-Cookie).
const SESSION_COOKIE = "assinapix_session";

// assinapix_role: papel do usuário (ex: "owner", "subscriber").
// Não é sensível — apenas para decisões de roteamento client-side.
const ROLE_COOKIE = "assinapix_role";

const MAX_AGE = 60 * 60 * 24; // 24 horas

/** Retorna "1" se há sessão ativa, null caso contrário. */
export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  const match = document.cookie.match(/(?:^|;\s*)assinapix_session=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

/**
 * Marca a sessão como ativa. O parâmetro token é ignorado —
 * o JWT real é gerenciado pelo backend via cookie HttpOnly.
 */
export function setToken(_token?: string): void {
  document.cookie = `${SESSION_COOKIE}=1; path=/; max-age=${MAX_AGE}; SameSite=Lax`;
}

/** Armazena o papel do usuário em cookie legível por JS (não sensível). */
export function setRole(role: string): void {
  document.cookie = `${ROLE_COOKIE}=${encodeURIComponent(role)}; path=/; max-age=${MAX_AGE}; SameSite=Lax`;
}

/** Retorna o papel do usuário ("owner" | "subscriber" | null). */
export function getRole(): string | null {
  if (typeof window === "undefined") return null;
  const match = document.cookie.match(/(?:^|;\s*)assinapix_role=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

/** Remove os cookies de sessão e papel. O HttpOnly cookie é removido pelo endpoint de logout. */
export function clearToken(): void {
  document.cookie = `${SESSION_COOKIE}=; path=/; max-age=0`;
  document.cookie = `${ROLE_COOKIE}=; path=/; max-age=0`;
}
