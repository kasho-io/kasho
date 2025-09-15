import { WorkOS } from "@workos-inc/node";

// Mock user data that matches what auth-wrapper returns
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

// Create a comprehensive mock client for testing
const createMockClient = () => {
  return {
    userManagement: {
      getUser: async () => mockUser,
      updateUser: async (_userId: string, data: { metadata?: Record<string, unknown> }) => ({
        ...mockUser,
        metadata: {
          ...mockUser.metadata,
          ...data.metadata,
        },
      }),
      authenticateWithCode: async () => ({
        user: mockUser,
        organizationId: "test-org-123",
      }),
      createOrganizationMembership: async () => ({
        id: "membership-123",
        userId: mockUser.id,
        organizationId: "test-org-123",
        roleSlug: "admin",
      }),
    },
    widgets: {
      getToken: async (_opts: { userId: string; organizationId: string; scopes?: string[] }) => "mock-widget-token",
    },
    portal: {
      generateLink: async (_opts: { organization: string; intent: string; returnUrl?: string }) => ({
        link: "https://example.com/mock-admin-portal",
      }),
    },
    organizations: {
      createOrganization: async (data: { name: string }) => ({
        id: "test-org-123",
        name: data.name,
        allowProfilesOutsideOrganization: false,
      }),
    },
  } as unknown as WorkOS;
};

// Check for MOCK_AUTH or missing API key (for test environments)
const shouldUseMock = process.env.MOCK_AUTH === "true" || !process.env.WORKOS_API_KEY;

// Use mock client in test mode, real client otherwise
export const workosClient = shouldUseMock ? createMockClient() : new WorkOS(process.env.WORKOS_API_KEY!);
