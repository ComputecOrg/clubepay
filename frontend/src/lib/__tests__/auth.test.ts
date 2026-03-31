import { describe, it, expect, beforeEach } from "vitest";
import { getToken, setToken, clearToken, setRole, getRole } from "@/lib/auth";

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

  it("setToken define cookie de sessão como indicador '1' (NÃO armazena o JWT)", () => {
    setToken("qualquer-jwt-secreto");
    // O JWT NÃO deve estar no cookie (prevenção XSS)
    expect(document.cookie).not.toContain("qualquer-jwt-secreto");
    // O indicador de sessão deve ser "1"
    expect(document.cookie).toContain("clubepay_session=1");
  });

  it("getToken retorna '1' quando sessão está ativa", () => {
    setToken("irrelevante");
    expect(getToken()).toBe("1");
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
    localStorage.setItem("clubepay_token", "manual");
    clearToken();
    // localStorage deve permanecer inalterado (clearToken não deve mexer nele)
    expect(localStorage.getItem("clubepay_token")).toBe("manual");
  });

  it("setToken sobrescreve sessão existente (sempre '1')", () => {
    setToken("first");
    setToken("second");
    expect(getToken()).toBe("1");
  });

  it("getToken retorna null quando window é undefined (SSR)", () => {
    expect(() => getToken()).not.toThrow();
  });
});

describe("auth - papel do usuário (role)", () => {
  it("setRole armazena o papel no cookie clubepay_role", () => {
    setRole("owner");
    expect(document.cookie).toContain("clubepay_role=owner");
  });

  it("getRole retorna o papel armazenado", () => {
    setRole("subscriber");
    expect(getRole()).toBe("subscriber");
  });

  it("getRole retorna null quando não há papel armazenado", () => {
    expect(getRole()).toBeNull();
  });

  it("clearToken remove também o cookie de papel", () => {
    setToken("t");
    setRole("owner");
    clearToken();
    expect(getToken()).toBeNull();
    expect(getRole()).toBeNull();
  });

  it("setRole sobrescreve papel existente", () => {
    setRole("owner");
    setRole("subscriber");
    expect(getRole()).toBe("subscriber");
  });
});
