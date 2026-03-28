import { describe, it, expect, beforeEach } from "vitest";
import { getToken, setToken, clearToken } from "@/lib/auth";

beforeEach(() => {
  localStorage.clear();
});

describe("auth", () => {
  it("setToken stores token in localStorage", () => {
    setToken("abc123");
    expect(localStorage.getItem("clubepay_token")).toBe("abc123");
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

  it("setToken overwrites existing token", () => {
    setToken("first");
    setToken("second");
    expect(getToken()).toBe("second");
  });
});
