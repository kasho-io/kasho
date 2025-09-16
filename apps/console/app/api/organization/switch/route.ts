import { NextResponse } from "next/server";
import { services } from "@/lib/services";

export async function POST(request: Request) {
  try {
    const session = await services.workos.withAuth();

    if (!session?.user) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const { organizationId } = await request.json();

    if (!organizationId) {
      return NextResponse.json({ error: "Organization ID is required" }, { status: 400 });
    }

    // Switch to the new organization context
    await services.workos.switchToOrganization(organizationId);

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error("Failed to switch organization:", error);
    return NextResponse.json({ error: "Failed to switch organization" }, { status: 500 });
  }
}
