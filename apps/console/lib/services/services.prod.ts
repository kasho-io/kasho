import { WorkOS, GeneratePortalLinkIntent } from "@workos-inc/node";
import { put } from "@vercel/blob";
import { withAuth, refreshSession } from "@workos-inc/authkit-nextjs";
import type { WorkOSService, VercelBlobService, WorkOSSession, WorkOSUser } from "./types";

export class ProductionWorkOSService implements WorkOSService {
  private workos: WorkOS;

  constructor() {
    this.workos = new WorkOS(process.env.WORKOS_API_KEY || "");
  }

  async withAuth(): Promise<WorkOSSession> {
    const session = await withAuth();
    // Map the actual WorkOS session to our interface
    if (!session.user) {
      throw new Error("No user in session");
    }
    return {
      user: {
        id: session.user.id,
        email: session.user.email,
        emailVerified: session.user.emailVerified,
        firstName: session.user.firstName ?? undefined,
        lastName: session.user.lastName ?? undefined,
        profilePictureUrl: session.user.profilePictureUrl ?? undefined,
        metadata: session.user.metadata,
      },
      sessionId: session.sessionId || "unknown",
      organizationId: session.organizationId ?? undefined,
      role: session.role ?? undefined,
      permissions: session.permissions,
      impersonator: session.impersonator,
    };
  }

  async refreshSession() {
    return refreshSession();
  }

  async getUser(userId: string): Promise<WorkOSUser> {
    const user = await this.workos.userManagement.getUser(userId);
    return {
      id: user.id,
      email: user.email,
      emailVerified: user.emailVerified,
      firstName: user.firstName ?? undefined,
      lastName: user.lastName ?? undefined,
      profilePictureUrl: user.profilePictureUrl ?? undefined,
      metadata: user.metadata,
    };
  }

  async updateUser(params: {
    userId: string;
    email?: string;
    metadata?: Record<string, unknown>;
  }): Promise<WorkOSUser> {
    // Convert metadata to the format WorkOS expects
    const updateParams = {
      userId: params.userId,
      email: params.email,
      metadata: params.metadata
        ? Object.fromEntries(Object.entries(params.metadata).map(([k, v]) => [k, v === undefined ? null : String(v)]))
        : undefined,
    };
    const user = await this.workos.userManagement.updateUser(updateParams);
    return {
      id: user.id,
      email: user.email,
      emailVerified: user.emailVerified,
      firstName: user.firstName ?? undefined,
      lastName: user.lastName ?? undefined,
      profilePictureUrl: user.profilePictureUrl ?? undefined,
      metadata: user.metadata,
    };
  }

  async getWidgetToken(params: { userId: string; organizationId: string; scopes?: string[] }) {
    // Pass parameters without scopes if not provided
    const tokenParams = params.scopes
      ? { userId: params.userId, organizationId: params.organizationId }
      : { userId: params.userId, organizationId: params.organizationId };
    const result = await this.workos.widgets.getToken(tokenParams);
    return result;
  }

  async generatePortalLink(params: { organization: string; intent: string; returnUrl?: string }) {
    return this.workos.portal.generateLink({
      ...params,
      intent: params.intent as GeneratePortalLinkIntent,
    });
  }

  async createOrganization(params: { name: string }) {
    return this.workos.organizations.createOrganization(params);
  }

  async createOrganizationMembership(params: { userId: string; organizationId: string; roleSlug?: string }) {
    return this.workos.userManagement.createOrganizationMembership(params);
  }
}

export class ProductionVercelBlobService implements VercelBlobService {
  async upload(
    pathname: string,
    body: Blob | ArrayBuffer | ArrayBufferView | string | ReadableStream,
    options?: {
      access?: "public" | "private";
      addRandomSuffix?: boolean;
      contentType?: string;
    },
  ) {
    // Cast body to a type that @vercel/blob accepts
    // Note: Vercel blob only supports "public" access currently
    const result = await put(pathname, body as Blob, {
      access: "public",
      addRandomSuffix: options?.addRandomSuffix,
      contentType: options?.contentType,
    });
    return {
      url: result.url,
      downloadUrl: result.downloadUrl ?? result.url,
      pathname: result.pathname,
    };
  }
}
