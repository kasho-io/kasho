import { services } from "@/lib/services";
import { redirect } from "next/navigation";
import { OrganizationManagement } from "./OrganizationManagement";
import { WorkOS } from "@workos-inc/node";

const workos = new WorkOS(process.env.WORKOS_API_KEY!);

export default async function OrganizationPage() {
  const session = await services.workos.withAuth();

  if (!session?.user) {
    redirect("/login");
  }

  // Check if user has permission to manage organization
  const canManageOrganization = session.permissions?.includes("organization:manage") || false;

  if (!canManageOrganization) {
    redirect("/account/profile");
  }

  let authToken = "";
  const organizationId = session.organizationId;
  let organizationName = "";
  // Get permissions directly from the session
  const userPermissions = session.permissions || [];

  // Fetch all organization memberships for the user
  let userOrganizations: Array<{ id: string; name: string }> = [];
  try {
    const memberships = await services.workos.listOrganizationMemberships({
      userId: session.user.id,
    });

    // Filter to only organizations where user has organization:manage permission
    // For now, we'll include all organizations and filter based on roles later if needed
    userOrganizations = memberships.data.map((membership) => ({
      id: membership.organization.id,
      name: membership.organization.name,
    }));
  } catch (error) {
    console.error("Failed to fetch user organizations:", error);
  }

  // Fetch organization details
  if (organizationId) {
    try {
      // Get organization details
      const organization = await workos.organizations.getOrganization(organizationId);
      organizationName = organization.name;

      // Fetch widget token for organization management
      authToken = await services.workos.getWidgetToken({
        userId: session.user.id,
        organizationId: organizationId,
        scopes: ["widgets:users-table:manage"],
      });
    } catch (error) {
      console.error("Failed to get organization details:", error);
    }
  }

  return (
    <div className="min-h-screen bg-base-200">
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-6xl mx-auto">
          <h1 className="text-3xl font-bold mb-8">Organization Settings</h1>
          <p className="text-base-content/70 mb-4">
            Manage your organization, invite team members, and configure roles.
          </p>

          <div className="space-y-6">
            <div className="card bg-base-100 shadow-xl">
              <div className="card-body">
                <OrganizationManagement
                  authToken={authToken}
                  organizationId={organizationId}
                  currentOrganizationName={organizationName}
                  userPermissions={userPermissions}
                  userOrganizations={userOrganizations}
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
