import { NextResponse } from "next/server";
import { services } from "@/lib/services";
import { WorkOS } from "@workos-inc/node";

const workos = new WorkOS(process.env.WORKOS_API_KEY!);

export async function PUT(request: Request) {
  try {
    const session = await services.workos.withAuth();
    if (!session?.user || !session?.organizationId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // Check if user has the organization:manage permission from the session
    const hasPermission = session.permissions?.includes('organization:manage') || false;

    if (!hasPermission) {
      return NextResponse.json(
        { error: "You don't have permission to manage this organization" },
        { status: 403 }
      );
    }

    const { name } = await request.json();

    if (!name || name.trim().length === 0) {
      return NextResponse.json(
        { error: "Organization name is required" },
        { status: 400 }
      );
    }

    // Update the organization - using the correct method signature
    const updatedOrganization = await workos.organizations.updateOrganization({
      organization: session.organizationId,
      name: name.trim(),
    });

    return NextResponse.json({
      success: true,
      organization: updatedOrganization
    });
  } catch (error: any) {
    console.error("Failed to update organization:", error);

    // Handle specific WorkOS error cases
    if (error?.status === 403) {
      if (error?.rawData?.message?.includes("Default test organizations")) {
        return NextResponse.json(
          { error: "Default test organizations cannot be updated. Please create a new organization to test this feature." },
          { status: 403 }
        );
      }
      return NextResponse.json(
        { error: "You don't have permission to update this organization" },
        { status: 403 }
      );
    }

    return NextResponse.json(
      { error: error?.message || "Failed to update organization" },
      { status: error?.status || 500 }
    );
  }
}