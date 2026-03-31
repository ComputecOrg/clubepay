import { describe, it, expect } from "vitest";
import { NextRequest } from "next/server";
import { middleware, config } from "../middleware";

function createRequest(path: string, cookieToken?: string): NextRequest {
  const url = new URL(path, "http://localhost:3000");
  const req = new NextRequest(url);
  if (cookieToken) {
    req.cookies.set("assinapix_token", cookieToken);
  }
  return req;
}

describe("middleware", () => {
  it("redirects to /login when accessing protected route without token", () => {
    const req = createRequest("/dashboard");
    const res = middleware(req);
    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toContain("/login");
  });

  it("allows access to protected route with token", () => {
    const req = createRequest("/dashboard", "valid-token");
    const res = middleware(req);
    expect(res.status).toBe(200);
  });

  it("redirects all protected paths without token", () => {
    const protectedPaths = ["/dashboard", "/planos", "/perfil", "/meu-plano"];
    for (const path of protectedPaths) {
      const req = createRequest(path);
      const res = middleware(req);
      expect(res.status).toBe(307);
      expect(res.headers.get("location")).toContain("/login");
    }
  });

  it("redirects logged-in users from /login to /dashboard", () => {
    const req = createRequest("/login", "valid-token");
    const res = middleware(req);
    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toContain("/dashboard");
  });

  it("redirects logged-in users from /register to /dashboard", () => {
    const req = createRequest("/register", "valid-token");
    const res = middleware(req);
    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toContain("/dashboard");
  });

  it("allows access to /login without token", () => {
    const req = createRequest("/login");
    const res = middleware(req);
    expect(res.status).toBe(200);
  });

  it("allows access to public routes without token", () => {
    const req = createRequest("/");
    const res = middleware(req);
    expect(res.status).toBe(200);
  });

  it("exports matcher config for protected and auth routes", () => {
    expect(config.matcher).toBeDefined();
    expect(config.matcher.length).toBeGreaterThan(0);
  });
});
