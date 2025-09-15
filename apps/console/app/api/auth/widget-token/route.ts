import { NextResponse } from "next/server";
import { withAuth } from "@/lib/auth-wrapper";
import { workosClient } from "@/lib/workos-client";

export async function GET() {
  try {
    const session = await withAuth();

    if (!session?.user || !session.organizationId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // Get widget token from WorkOS
    const token = await workosClient.widgets.getToken({
      userId: session.user.id,
      organizationId: session.organizationId,
      scopes: ["widgets:users-table:manage"],
    });

    // Set CORS headers for WorkOS widgets
    const headers = new Headers({
      "Access-Control-Allow-Origin": "https://api.workos.com",
      "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type, Authorization",
      "Access-Control-Allow-Credentials": "true",
    });

    return NextResponse.json(
      {
        token: token,
        organizationId: session.organizationId,
      },
      { headers },
    );
  } catch (error) {
    console.error("Error getting widget token:", error);
    return NextResponse.json({ error: "Failed to get widget token" }, { status: 500 });
  }
}

export async function OPTIONS() {
  // Handle preflight requests
  return new NextResponse(null, {
    status: 200,
    headers: {
      "Access-Control-Allow-Origin": "https://api.workos.com",
      "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type, Authorization",
      "Access-Control-Allow-Credentials": "true",
    },
  });
}
