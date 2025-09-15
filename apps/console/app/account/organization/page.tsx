import { withAuth } from "@/lib/auth-wrapper";
import { redirect } from "next/navigation";
import { OrganizationManagement } from "./OrganizationManagement";
import { workosClient } from "@/lib/workos-client";

export default async function OrganizationPage() {
  const session = await withAuth();

  if (!session?.user) {
    redirect("/login");
  }

  let authToken = "";
  const organizationId = session.organizationId;

  // Fetch widget token for organization management
  if (organizationId) {
    try {
      authToken = await workosClient.widgets.getToken({
        userId: session.user.id,
        organizationId: organizationId,
        scopes: ["widgets:users-table:manage"],
      });
    } catch (error) {
      console.error("Failed to get widget token:", error);
    }
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-3xl font-bold mb-8">Organization Management</h1>

        <div className="space-y-6">
          <div className="card bg-base-100 shadow-xl">
            <div className="card-body">
              <h2 className="card-title">Organization Settings</h2>
              <p className="text-base-content/70 mb-4">
                Manage your organization, invite team members, and configure roles.
              </p>

              <OrganizationManagement authToken={authToken} organizationId={organizationId} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
