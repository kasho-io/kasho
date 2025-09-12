import { validateRequestOrigin, validateRequestMethod, isRequestValid } from "../csrf-protection";

// Mock NextRequest for testing
class MockNextRequest {
  public method: string;
  private _headers: Map<string, string>;
  public url: string;

  constructor(url: string, options?: { method?: string; headers?: Record<string, string> }) {
    this.url = url;
    this.method = options?.method || "GET";
    this._headers = new Map(Object.entries(options?.headers || {}));
  }

  // Mock the headers.get method
  get headers() {
    return {
      get: (name: string): string | null => {
        return this._headers.get(name.toLowerCase()) || null;
      },
    };
  }
}

// Cast to NextRequest type for TypeScript
import { NextRequest as NextRequestType } from "next/server";
const NextRequest = MockNextRequest as unknown as typeof NextRequestType;

describe("csrf-protection", () => {
  const originalEnv = process.env.NODE_ENV;

  afterEach(() => {
    process.env.NODE_ENV = originalEnv;
  });

  describe("validateRequestOrigin", () => {
    describe("development environment", () => {
      beforeEach(() => {
        process.env.NODE_ENV = "development";
      });

      it("should accept localhost origins in development", () => {
        const request = new NextRequest("http://localhost:3000/api/test", {
          headers: {
            origin: "http://localhost:3000",
          },
        });
        expect(validateRequestOrigin(request)).toBe(true);
      });

      it("should accept 127.0.0.1 origins in development", () => {
        const request = new NextRequest("http://localhost:3000/api/test", {
          headers: {
            origin: "http://127.0.0.1:3000",
          },
        });
        expect(validateRequestOrigin(request)).toBe(true);
      });

      it("should accept localhost:3001 and :3002 in development", () => {
        const request3001 = new NextRequest("http://localhost:3000/api/test", {
          headers: {
            origin: "http://localhost:3001",
          },
        });
        expect(validateRequestOrigin(request3001)).toBe(true);

        const request3002 = new NextRequest("http://localhost:3000/api/test", {
          headers: {
            origin: "http://localhost:3002",
          },
        });
        expect(validateRequestOrigin(request3002)).toBe(true);
      });

      it("should accept requests with matching referer in development", () => {
        const request = new NextRequest("http://localhost:3000/api/test", {
          headers: {
            referer: "http://localhost:3000/page",
          },
        });
        expect(validateRequestOrigin(request)).toBe(true);
      });

      it("should reject non-localhost origins in development", () => {
        const request = new NextRequest("http://localhost:3000/api/test", {
          headers: {
            origin: "http://evil.com",
          },
        });
        expect(validateRequestOrigin(request)).toBe(false);
      });

      it("should reject requests without origin or referer", () => {
        const request = new NextRequest("http://localhost:3000/api/test");
        expect(validateRequestOrigin(request)).toBe(false);
      });
    });

    describe("production environment", () => {
      beforeEach(() => {
        process.env.NODE_ENV = "production";
      });

      it("should validate origin matches host", () => {
        const request = new NextRequest("https://app.example.com/api/test", {
          headers: {
            origin: "https://app.example.com",
            host: "app.example.com",
          },
        });
        expect(validateRequestOrigin(request)).toBe(true);
      });

      it("should validate referer matches host", () => {
        const request = new NextRequest("https://app.example.com/api/test", {
          headers: {
            referer: "https://app.example.com/page",
            host: "app.example.com",
          },
        });
        expect(validateRequestOrigin(request)).toBe(true);
      });

      it("should reject mismatched origin", () => {
        const request = new NextRequest("https://app.example.com/api/test", {
          headers: {
            origin: "https://evil.com",
            host: "app.example.com",
          },
        });
        expect(validateRequestOrigin(request)).toBe(false);
      });

      it("should reject mismatched referer", () => {
        const request = new NextRequest("https://app.example.com/api/test", {
          headers: {
            referer: "https://evil.com/page",
            host: "app.example.com",
          },
        });
        expect(validateRequestOrigin(request)).toBe(false);
      });

      it("should reject requests without origin or referer", () => {
        const request = new NextRequest("https://app.example.com/api/test", {
          headers: {
            host: "app.example.com",
          },
        });
        expect(validateRequestOrigin(request)).toBe(false);
      });

      it("should handle invalid URL in origin", () => {
        const request = new NextRequest("https://app.example.com/api/test", {
          headers: {
            origin: "not-a-valid-url",
            host: "app.example.com",
          },
        });
        expect(validateRequestOrigin(request)).toBe(false);
      });

      it("should handle invalid URL in referer", () => {
        const request = new NextRequest("https://app.example.com/api/test", {
          headers: {
            referer: "not-a-valid-url",
            host: "app.example.com",
          },
        });
        expect(validateRequestOrigin(request)).toBe(false);
      });

      it("should use NEXT_PUBLIC_APP_URL if host is not provided", () => {
        const originalUrl = process.env.NEXT_PUBLIC_APP_URL;
        process.env.NEXT_PUBLIC_APP_URL = "app.example.com";

        const request = new NextRequest("https://app.example.com/api/test", {
          headers: {
            origin: "https://app.example.com",
          },
        });
        expect(validateRequestOrigin(request)).toBe(true);

        process.env.NEXT_PUBLIC_APP_URL = originalUrl;
      });
    });
  });

  describe("validateRequestMethod", () => {
    it("should accept allowed methods", () => {
      const getRequest = new NextRequest("http://localhost:3000/api/test", { method: "GET" });
      expect(validateRequestMethod(getRequest, ["GET", "POST"])).toBe(true);

      const postRequest = new NextRequest("http://localhost:3000/api/test", { method: "POST" });
      expect(validateRequestMethod(postRequest, ["GET", "POST"])).toBe(true);
    });

    it("should reject disallowed methods", () => {
      const deleteRequest = new NextRequest("http://localhost:3000/api/test", { method: "DELETE" });
      expect(validateRequestMethod(deleteRequest, ["GET", "POST"])).toBe(false);

      const putRequest = new NextRequest("http://localhost:3000/api/test", { method: "PUT" });
      expect(validateRequestMethod(putRequest, ["GET", "POST"])).toBe(false);
    });

    it("should handle empty allowed methods array", () => {
      const request = new NextRequest("http://localhost:3000/api/test", { method: "GET" });
      expect(validateRequestMethod(request, [])).toBe(false);
    });

    it("should be case-sensitive", () => {
      const request = new NextRequest("http://localhost:3000/api/test", { method: "GET" });
      expect(validateRequestMethod(request, ["get"])).toBe(false);
    });
  });

  describe("isRequestValid", () => {
    beforeEach(() => {
      process.env.NODE_ENV = "development";
    });

    it("should validate request with valid origin and method", () => {
      const request = new NextRequest("http://localhost:3000/api/test", {
        method: "POST",
        headers: {
          origin: "http://localhost:3000",
          "content-type": "application/json",
        },
      });

      const result = isRequestValid(request);
      expect(result.valid).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it("should reject invalid method", () => {
      const request = new NextRequest("http://localhost:3000/api/test", {
        method: "GET",
        headers: {
          origin: "http://localhost:3000",
        },
      });

      const result = isRequestValid(request, { allowedMethods: ["POST"] });
      expect(result.valid).toBe(false);
      expect(result.error).toBe("Method GET not allowed");
    });

    it("should reject invalid origin for non-GET requests", () => {
      const request = new NextRequest("http://localhost:3000/api/test", {
        method: "POST",
        headers: {
          origin: "http://evil.com",
          "content-type": "application/json",
        },
      });

      const result = isRequestValid(request);
      expect(result.valid).toBe(false);
      expect(result.error).toBe("Invalid request origin");
    });

    it("should skip origin check for GET requests", () => {
      const request = new NextRequest("http://localhost:3000/api/test", {
        method: "GET",
        headers: {
          origin: "http://evil.com",
        },
      });

      const result = isRequestValid(request, { allowedMethods: ["GET"] });
      expect(result.valid).toBe(true);
    });

    it("should skip origin check when skipOriginCheck is true", () => {
      const request = new NextRequest("http://localhost:3000/api/test", {
        method: "POST",
        headers: {
          origin: "http://evil.com",
          "content-type": "application/json",
        },
      });

      const result = isRequestValid(request, { skipOriginCheck: true });
      expect(result.valid).toBe(true);
    });

    it("should reject POST requests with invalid content type", () => {
      const request = new NextRequest("http://localhost:3000/api/test", {
        method: "POST",
        headers: {
          origin: "http://localhost:3000",
          "content-type": "text/plain",
        },
      });

      const result = isRequestValid(request);
      expect(result.valid).toBe(false);
      expect(result.error).toBe("Invalid content type");
    });

    it("should accept POST requests with application/json", () => {
      const request = new NextRequest("http://localhost:3000/api/test", {
        method: "POST",
        headers: {
          origin: "http://localhost:3000",
          "content-type": "application/json",
        },
      });

      const result = isRequestValid(request);
      expect(result.valid).toBe(true);
    });

    it("should accept POST requests with multipart/form-data", () => {
      const request = new NextRequest("http://localhost:3000/api/test", {
        method: "POST",
        headers: {
          origin: "http://localhost:3000",
          "content-type": "multipart/form-data; boundary=----WebKitFormBoundary",
        },
      });

      const result = isRequestValid(request);
      expect(result.valid).toBe(true);
    });

    it("should accept POST requests without content-type header", () => {
      const request = new NextRequest("http://localhost:3000/api/test", {
        method: "POST",
        headers: {
          origin: "http://localhost:3000",
        },
      });

      const result = isRequestValid(request);
      expect(result.valid).toBe(true);
    });

    it("should handle all options together", () => {
      const request = new NextRequest("http://localhost:3000/api/test", {
        method: "PUT",
        headers: {
          origin: "http://localhost:3000",
          "content-type": "application/json",
        },
      });

      const result = isRequestValid(request, {
        allowedMethods: ["PUT", "PATCH"],
        skipOriginCheck: false,
      });
      expect(result.valid).toBe(true);
    });

    it("should use default allowed methods when not specified", () => {
      const deleteRequest = new NextRequest("http://localhost:3000/api/test", {
        method: "DELETE",
        headers: {
          origin: "http://localhost:3000",
        },
      });

      const result = isRequestValid(deleteRequest);
      expect(result.valid).toBe(true); // DELETE is in default allowed methods

      const getRequest = new NextRequest("http://localhost:3000/api/test", {
        method: "GET",
        headers: {
          origin: "http://localhost:3000",
        },
      });

      const getResult = isRequestValid(getRequest);
      expect(getResult.valid).toBe(false); // GET is not in default allowed methods
    });
  });
});
