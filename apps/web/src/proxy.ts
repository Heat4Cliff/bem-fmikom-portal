import { NextResponse, type NextRequest } from "next/server";

const securityHeaders: Record<string, string> = {
  "X-Content-Type-Options": "nosniff",
  "X-Frame-Options": "DENY",
  "Referrer-Policy": "strict-origin-when-cross-origin",
  "Permissions-Policy": "camera=(), microphone=(), geolocation=()",
  "Cross-Origin-Opener-Policy": "same-origin",
};

function contentSecurityPolicy(nonce: string) {
  const apiOrigin = process.env.NEXT_PUBLIC_API_BASE_URL
    ? new URL(process.env.NEXT_PUBLIC_API_BASE_URL).origin
    : "'self'";
  const connectSources = apiOrigin === "'self'" ? "'self'" : `'self' ${apiOrigin}`;
  const imageSources =
    apiOrigin === "'self'" ? "'self' data: blob: https:" : `'self' data: blob: https: ${apiOrigin}`;

  return [
    "default-src 'self'",
    `script-src 'self' 'nonce-${nonce}' 'strict-dynamic'`,
    `style-src 'self' 'nonce-${nonce}'`,
    `img-src ${imageSources}`,
    "font-src 'self' data:",
    `connect-src ${connectSources}`,
    "object-src 'none'",
    "base-uri 'self'",
    "frame-ancestors 'none'",
    "form-action 'self'",
    ...(process.env.NODE_ENV === "production" ? ["upgrade-insecure-requests"] : []),
  ].join("; ");
}

export function proxy(request: NextRequest) {
  const headers = new Headers(request.headers);
  const nonce = btoa(crypto.randomUUID());
  const isProduction = process.env.NODE_ENV === "production";
  const csp = contentSecurityPolicy(nonce);

  if (isProduction) {
    headers.set("Content-Security-Policy", csp);
    headers.set("x-nonce", nonce);
  }

  const response = NextResponse.next({ request: { headers } });
  for (const [name, value] of Object.entries(securityHeaders)) {
    response.headers.set(name, value);
  }
  if (isProduction) {
    response.headers.set("Content-Security-Policy", csp);
    if (process.env.NEXT_PUBLIC_SITE_URL?.startsWith("https://")) {
      response.headers.set("Strict-Transport-Security", "max-age=63072000; includeSubDomains");
    }
  }
  if (request.nextUrl.pathname.startsWith("/admin")) {
    response.headers.set("Cache-Control", "private, no-store");
  }
  return response;
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml).*)"],
};
