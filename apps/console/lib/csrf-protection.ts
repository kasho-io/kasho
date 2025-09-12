import { NextRequest } from "next/server";

/**
 * Validates that the request is coming from a trusted origin
 * This provides CSRF protection for API routes
 */
export function validateRequestOrigin(request: NextRequest): boolean {
  const origin = request.headers.get("origin");
  const referer = request.headers.get("referer");
  const host = request.headers.get("host");

  // In development, allow localhost origins
  if (process.env.NODE_ENV === "development") {
    const allowedOrigins = [
      "http://localhost:3000",
      "http://localhost:3001",
      "http://localhost:3002",
      "http://127.0.0.1:3000",
    ];

    if (origin && allowedOrigins.includes(origin)) {
      return true;
    }

    if (referer) {
      if (allowedOrigins.some((allowed) => referer.startsWith(allowed))) {
        return true;
      }
    }
  }

  // In production, validate against the actual host
  if (!origin && !referer) {
    // No origin or referer header - could be a direct API call
    // Check for API key or other authentication method if needed
    return false;
  }

  // Validate origin matches host
  if (origin) {
    try {
      const originUrl = new URL(origin);
      const expectedHost = host || process.env.NEXT_PUBLIC_APP_URL;

      if (expectedHost && originUrl.host === expectedHost) {
        return true;
      }
    } catch {
      return false;
    }
  }

  // Validate referer matches host
  if (referer) {
    try {
      const refererUrl = new URL(referer);
      const expectedHost = host || process.env.NEXT_PUBLIC_APP_URL;

      if (expectedHost && refererUrl.host === expectedHost) {
        return true;
      }
    } catch {
      return false;
    }
  }

  return false;
}

/**
 * Validates that the request method is allowed
 */
export function validateRequestMethod(request: NextRequest, allowedMethods: string[]): boolean {
  return allowedMethods.includes(request.method);
}

/**
 * Comprehensive CSRF protection check
 */
export function isRequestValid(
  request: NextRequest,
  options: {
    requireAuth?: boolean;
    allowedMethods?: string[];
    skipOriginCheck?: boolean;
  } = {},
): { valid: boolean; error?: string } {
  const { allowedMethods = ["POST", "PUT", "DELETE", "PATCH"], skipOriginCheck = false } = options;
  // Note: requireAuth parameter reserved for future use

  // Check request method
  if (allowedMethods.length > 0 && !validateRequestMethod(request, allowedMethods)) {
    return {
      valid: false,
      error: `Method ${request.method} not allowed`,
    };
  }

  // Skip origin check for GET requests or if explicitly disabled
  if (!skipOriginCheck && request.method !== "GET") {
    if (!validateRequestOrigin(request)) {
      return {
        valid: false,
        error: "Invalid request origin",
      };
    }
  }

  // Additional security headers check
  const contentType = request.headers.get("content-type");
  if (
    request.method === "POST" &&
    contentType &&
    !contentType.includes("application/json") &&
    !contentType.includes("multipart/form-data")
  ) {
    return {
      valid: false,
      error: "Invalid content type",
    };
  }

  return { valid: true };
}
