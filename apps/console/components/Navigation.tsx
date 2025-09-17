import { services } from "@/lib/services";
import { WorkOSUser } from "@/lib/validation-schemas";
import NavigationClient from "./NavigationClient";

export default async function Navigation() {
  let user: WorkOSUser | null = null;
  let canManageOrganization = false;

  try {
    const session = await services.workos.withAuth();
    user = session?.user as WorkOSUser;
    // Check if user has organization management permission
    canManageOrganization = session?.permissions?.includes("organization:manage") || false;
  } catch {
    // User is not authenticated - that's ok, we'll show Sign In button
  }

  return <NavigationClient user={user} canManageOrganization={canManageOrganization} />;
}
