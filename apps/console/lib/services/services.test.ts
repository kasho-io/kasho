import type { WorkOSService, VercelBlobService } from "./types";

// Mock user data for testing
const mockUser = {
  id: "test-user-123",
  email: "test@example.com",
  emailVerified: true,
  firstName: "Test",
  lastName: "User",
  profilePictureUrl: "",
  metadata: {
    first_name: "Test",
    last_name: "User",
    profile_picture_url: "",
  },
};

// Mock session for testing
const mockSession = {
  user: mockUser,
  sessionId: "test-session-123",
  organizationId: "test-org-123",
  role: "admin",
  permissions: ["organization:manage"],
  impersonator: null,
};

export class TestWorkOSService implements WorkOSService {
  async withAuth() {
    return mockSession;
  }

  async refreshSession() {
    // No-op in test mode
    return Promise.resolve();
  }

  async getUser(_userId: string) {
    return mockUser;
  }

  async updateUser(params: { userId: string; email?: string; metadata?: Record<string, unknown> }) {
    return {
      ...mockUser,
      email: params.email || mockUser.email,
      metadata: {
        ...mockUser.metadata,
        ...params.metadata,
      },
    };
  }

  async getWidgetToken(_params: { userId: string; organizationId: string; scopes?: string[] }) {
    return "mock-widget-token";
  }

  async generatePortalLink(_params: { organization: string; intent: string; returnUrl?: string }) {
    return {
      link: "https://example.com/mock-admin-portal",
    };
  }

  async createOrganization(params: { name: string }) {
    return {
      id: "test-org-123",
      name: params.name,
      allowProfilesOutsideOrganization: false,
    };
  }

  async createOrganizationMembership(params: { userId: string; organizationId: string; roleSlug?: string }) {
    return {
      id: "membership-123",
      userId: params.userId,
      organizationId: params.organizationId,
      roleSlug: params.roleSlug || "member",
    };
  }
}

export class TestVercelBlobService implements VercelBlobService {
  async upload(
    pathname: string,
    _body: Blob | ArrayBuffer | ArrayBufferView | string | ReadableStream,
    _options?: {
      access?: "public" | "private";
      addRandomSuffix?: boolean;
      contentType?: string;
    },
  ) {
    return {
      url: `https://example.com/mock-upload/${pathname}`,
      downloadUrl: `https://example.com/mock-upload/${pathname}`,
      pathname,
    };
  }
}
