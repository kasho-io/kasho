import { NextResponse } from "next/server";
import { services } from "@/lib/services";
import { GeneratePortalLinkIntent } from "@workos-inc/node";

export async function POST() {
  try {
    const session = await services.workos.withAuth();

    if (!session?.user || !session.organizationId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // Generate an Admin Portal link for the organization
    const { link } = await services.workos.generatePortalLink({
      organization: session.organizationId,
      intent: GeneratePortalLinkIntent.SSO.toString(),
      returnUrl: `${process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000"}/account/organization`,
    });

    return NextResponse.json({ link });
  } catch (error) {
    console.error("Error generating admin portal link:", error);
    return NextResponse.json({ error: "Failed to generate admin portal link" }, { status: 500 });
  }
}
