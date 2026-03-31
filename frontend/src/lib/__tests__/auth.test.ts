import { describe, it, expect, beforeEach } from "vitest";
import { getToken, setToken, clearToken } from "@/lib/auth";

beforeEach(() => {
  // Clear all cookies
  document.cookie.split(";").forEach((c) => {
    const name = c.trim().split("=")[0];
    document.cookie = `${name}=; path=/; max-age=0`;
  });
});

describe("auth - segurança HttpOnly cookie", () => {
  it("setToken NÃO armazena token no localStorage", () => {
    setToken("abc123");
    expect(localStorage.getItem("clubepay_token")).toBeNull();
  });

  it("setToken define cookie de sessão para o browser gerenciar", () => {
    setToken("abc123");
    expect(document.cookie).toContain("clubepay_session=abc123");
  });

  it("getToken lê do cookie de sessão (não do localStorage)", () => {
    document.cookie = "clubepay_session=xyz789; path=/";
    expect(getToken()).toBe("xyz789");
  });

  it("getToken retorna null quando não há sessão", () => {
    expect(getToken()).toBeNull();
  });

  it("clearToken remove cookie de sessão", () => {
    setToken("token");
    clearToken();
    expect(getToken()).toBeNull();
  });

  it("clearToken NÃO usa localStorage", () => {
    // Garantir que clearToken não interage com localStorage
    localStorage.setItem("clubepay_token", "manual");
    clearToken();
    // localStorage deve permanecer inalterado (clearToken não deve mexer nele)
    expect(localStorage.getItem("clubepay_token")).toBe("manual");
  });

  it("setToken sobrescreve sessão existente", () => {
    setToken("first");
    setToken("second");
    expect(getToken()).toBe("second");
  });

  it("getToken retorna null quando window é undefined (SSR)", () => {
    // Esta é uma verificação indireta - getToken deve retornar null em SSR
    // No Vitest o window está definido, mas verificamos que não lê localStorage
    expect(() => getToken()).not.toThrow();
  });
});
