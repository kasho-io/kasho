import { authkitMiddleware } from "@workos-inc/authkit-nextjs";
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// Middleware checks NODE_ENV for test bypassing
const middleware =
  process.env.NODE_ENV === "test" ? (_request: NextRequest) => NextResponse.next() : authkitMiddleware();

export default middleware;

// Match all routes except static files
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    "/((?!_next/static|_next/image|favicon.ico).*)",
  ],
};
