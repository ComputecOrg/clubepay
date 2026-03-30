import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const protectedPaths = ["/dashboard", "/planos", "/perfil", "/meu-plano"];
const authPaths = ["/login", "/register"];

export function middleware(request: NextRequest) {
  const token = request.cookies.get("clubepay_token")?.value;
  const { pathname } = request.nextUrl;

  const isProtected = protectedPaths.some((path) => pathname.startsWith(path));
  const isAuthPage = authPaths.some((path) => pathname.startsWith(path));

  if (isProtected && !token) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  if (isAuthPage && token) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/dashboard/:path*",
    "/planos/:path*",
    "/perfil/:path*",
    "/meu-plano/:path*",
    "/login",
    "/register",
  ],
};
