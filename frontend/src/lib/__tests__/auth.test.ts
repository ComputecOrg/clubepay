import { describe, it, expect, beforeEach } from "vitest";
import { getToken, setToken, clearToken } from "@/lib/auth";

beforeEach(() => {
  localStorage.clear();
  document.cookie = "clubepay_token=; path=/; max-age=0";
});

describe("auth", () => {
  it("setToken stores token in localStorage", () => {
    setToken("abc123");
    expect(localStorage.getItem("clubepay_token")).toBe("abc123");
  });

  it("setToken sets cookie for middleware", () => {
    setToken("abc123");
    expect(document.cookie).toContain("clubepay_token=abc123");
  });

  it("getToken retrieves token from localStorage", () => {
    localStorage.setItem("clubepay_token", "xyz789");
    expect(getToken()).toBe("xyz789");
  });

  it("getToken returns null when no token", () => {
    expect(getToken()).toBeNull();
  });

  it("clearToken removes token from localStorage", () => {
    setToken("token");
    clearToken();
    expect(getToken()).toBeNull();
  });

  it("clearToken removes cookie", () => {
    setToken("token");
    clearToken();
    expect(document.cookie).not.toContain("clubepay_token=token");
  });

  it("setToken overwrites existing token", () => {
    setToken("first");
    setToken("second");
    expect(getToken()).toBe("second");
  });
});
