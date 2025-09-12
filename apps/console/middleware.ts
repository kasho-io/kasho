import { authkitMiddleware } from "@workos-inc/authkit-nextjs";

export default authkitMiddleware();

// Match all routes except static files
export const config = {
  unstable_allowDynamic: [
    // Allow WorkOS AuthKit to use dynamic code evaluation in Edge Runtime
    "**/node_modules/@workos-inc/**",
  ],
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
