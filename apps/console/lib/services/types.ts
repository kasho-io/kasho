// Service interfaces for dependency injection

// WorkOS types
export interface WorkOSUser {
  id: string;
  email: string;
  emailVerified: boolean;
  firstName?: string;
  lastName?: string;
  profilePictureUrl?: string;
  metadata?: Record<string, unknown>;
}

export interface WorkOSSession {
  user: WorkOSUser;
  sessionId: string;
  organizationId?: string;
  role?: string;
  permissions?: string[] | null;
  impersonator?: unknown;
}

export interface WorkOSService {
  // Authentication
  withAuth(): Promise<WorkOSSession>;
  refreshSession(): Promise<unknown>;

  // User Management
  getUser(userId: string): Promise<WorkOSUser>;
  updateUser(params: { userId: string; email?: string; metadata?: Record<string, unknown> }): Promise<WorkOSUser>;

  // Widgets
  getWidgetToken(params: { userId: string; organizationId: string; scopes?: string[] }): Promise<string>;

  // Portal
  generatePortalLink(params: { organization: string; intent: string; returnUrl?: string }): Promise<{ link: string }>;

  // Organizations
  createOrganization(params: { name: string }): Promise<{
    id: string;
    name: string;
    allowProfilesOutsideOrganization?: boolean;
  }>;
  createOrganizationMembership(params: { userId: string; organizationId: string; roleSlug?: string }): Promise<{
    id: string;
    userId: string;
    organizationId: string;
    roleSlug?: string;
  }>;
}

export interface VercelBlobService {
  upload(
    pathname: string,
    body: Blob | ArrayBuffer | ArrayBufferView | string | ReadableStream,
    options?: {
      access?: "public" | "private";
      addRandomSuffix?: boolean;
      contentType?: string;
    },
  ): Promise<{
    url: string;
    downloadUrl?: string;
    pathname: string;
  }>;
}

export interface Services {
  workos: WorkOSService;
  vercelBlob: VercelBlobService;
}
