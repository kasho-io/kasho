import { NextResponse } from "next/server";
import { services } from "@/lib/services";
import { handleFirstTimeUserOrganization } from "@/lib/auth-helpers";
import { WorkOSUser } from "@/lib/validation-schemas";

export async function POST() {
  try {
    // Get the current user session
    const session = await services.workos.withAuth();

    if (!session.user) {
      return NextResponse.json({ error: "Not authenticated" }, { status: 401 });
    }

    const user = session.user as WorkOSUser;

    // Check and create organization if needed
    const newOrganizationId = await handleFirstTimeUserOrganization(user);

    if (newOrganizationId) {
      // Switch to the new organization context
      await services.workos.switchToOrganization(newOrganizationId);

      return NextResponse.json({
        success: true,
        organizationId: newOrganizationId,
        requiresRefresh: true,
      });
    }

    return NextResponse.json({
      success: true,
      organizationId: session.organizationId,
      requiresRefresh: false,
    });
  } catch (error) {
    console.error("Error setting up organization:", error);
    return NextResponse.json({ error: "Failed to setup organization" }, { status: 500 });
  }
}
