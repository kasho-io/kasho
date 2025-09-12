import { withAuth as workosWithAuth } from "@workos-inc/authkit-nextjs";

// Mock user for testing
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

/**
 * Wrapper around WorkOS withAuth that can be mocked in test environments
 */
export async function withAuth() {
  // In test mode, return mock user
  // Check MOCK_AUTH first since it's explicitly set for tests
  if (process.env.MOCK_AUTH === "true" || process.env.NODE_ENV === "test") {
    return {
      user: mockUser,
      sessionId: "test-session-123",
      organizationId: null,
      role: null,
      permissions: null,
      impersonator: null,
    };
  }

  // In production, use real WorkOS auth
  return workosWithAuth();
}

/**
 * Get mock user for testing (can be customized per test)
 */
export function getMockUser(overrides?: Partial<typeof mockUser>) {
  return {
    ...mockUser,
    ...overrides,
    metadata: {
      ...mockUser.metadata,
      ...(overrides?.metadata || {}),
    },
  };
}
