import "@testing-library/jest-dom";

// Polyfill for Next.js Request/Response in tests
if (!globalThis.Request) {
  globalThis.Request = class Request {
    public url: string;
    public method: string;
    public headers: Headers;

    constructor(url: string, init?: RequestInit) {
      this.url = url;
      this.method = init?.method || "GET";
      this.headers = new Headers(init?.headers);
    }
  } as unknown as typeof Request;
}

if (!globalThis.Response) {
  globalThis.Response = class Response {
    constructor(
      public body?: BodyInit | null,
      public init?: ResponseInit,
    ) {}
  } as unknown as typeof Response;
}

// Headers polyfill if needed
if (!globalThis.Headers) {
  globalThis.Headers = class Headers {
    private headers: Map<string, string> = new Map();

    constructor(init?: HeadersInit) {
      if (init) {
        if (init instanceof Headers) {
          init.forEach((value, key) => this.headers.set(key.toLowerCase(), value));
        } else if (Array.isArray(init)) {
          init.forEach(([key, value]) => this.headers.set(key.toLowerCase(), value));
        } else {
          Object.entries(init).forEach(([key, value]) => this.headers.set(key.toLowerCase(), value));
        }
      }
    }

    get(name: string): string | null {
      return this.headers.get(name.toLowerCase()) || null;
    }

    set(name: string, value: string): void {
      this.headers.set(name.toLowerCase(), value);
    }

    forEach(callback: (value: string, key: string) => void): void {
      this.headers.forEach(callback);
    }
  } as unknown as typeof Headers;
}
