import { services } from "@/lib/services";
import { WorkOSUser } from "@/lib/validation-schemas";

/**
 * Check if a user has no organization memberships
 */
export async function hasNoOrganization(userId: string): Promise<boolean> {
  try {
    const memberships = await services.workos.listOrganizationMemberships({ userId });
    return memberships.data.length === 0;
  } catch (error) {
    console.error("Error checking organization memberships:", error);
    // In case of error, assume they have an org to avoid creating duplicates
    return false;
  }
}

/**
 * Create a default organization for a user
 * Organization name format: "<Full Name>'s Organization" or "<email>'s Organization"
 */
export async function createDefaultOrganization(user: WorkOSUser): Promise<string | null> {
  try {
    // Construct organization name
    let orgName: string;
    if (user.firstName && user.lastName) {
      orgName = `${user.firstName} ${user.lastName}'s Organization`;
    } else if (user.firstName) {
      orgName = `${user.firstName}'s Organization`;
    } else {
      orgName = `${user.email}'s Organization`;
    }

    const org = await services.workos.createOrganization({ name: orgName });
    return org.id;
  } catch (error) {
    console.error("Error creating organization:", error);
    return null;
  }
}

/**
 * Assign a user as admin to an organization
 */
export async function assignUserAsAdmin(userId: string, organizationId: string): Promise<boolean> {
  try {
    await services.workos.createOrganizationMembership({
      userId,
      organizationId,
      roleSlug: "admin", // Assuming 'admin' is the role slug for administrators
    });
    return true;
  } catch (error) {
    console.error("Error creating organization membership:", error);
    return false;
  }
}

/**
 * Handle first-time user organization setup
 * Creates an organization and assigns the user as admin if they don't have any organizations
 */
export async function handleFirstTimeUserOrganization(user: WorkOSUser): Promise<string | null> {
  // Check if user has no organizations
  const needsOrganization = await hasNoOrganization(user.id);

  if (!needsOrganization) {
    return null; // User already has an organization
  }

  // Create default organization
  const organizationId = await createDefaultOrganization(user);

  if (!organizationId) {
    console.error("Failed to create organization for user:", user.id);
    return null;
  }

  // Assign user as admin
  const success = await assignUserAsAdmin(user.id, organizationId);

  if (!success) {
    console.error("Failed to assign user as admin to organization:", organizationId);
    // Organization was created but membership failed - this is still a partial success
  }

  return organizationId;
}
