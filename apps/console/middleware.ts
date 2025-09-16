import { authkitMiddleware } from "@workos-inc/authkit-nextjs";
import { NextRequest, NextResponse } from "next/server";

// Create middleware function that checks MOCK_AUTH at runtime
const middleware = (request: NextRequest) => {
  // Check if we're in mock mode at runtime
  if (process.env.MOCK_AUTH === "true") {
    return NextResponse.next();
  }

  // Otherwise use the real auth middleware
  return authkitMiddleware()(request, {} as never);
};

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
