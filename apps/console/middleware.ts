import { authkitMiddleware } from "@workos-inc/authkit-nextjs";

// Middleware always runs normally
// MOCK_AUTH is handled in the services layer
export default authkitMiddleware();

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
