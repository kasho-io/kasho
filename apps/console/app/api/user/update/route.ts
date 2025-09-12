import { NextRequest, NextResponse } from "next/server";
import { withAuth } from "@/lib/auth-wrapper";
import { workosClient } from "@/lib/workos-client";

export async function POST(request: NextRequest) {
  try {
    // Get the authenticated user
    const { user } = await withAuth();

    if (!user) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // Parse the request body
    const body = await request.json();
    const { email, metadata } = body;

    // In test mode, just return success without calling WorkOS
    if (process.env.NODE_ENV === "test" || process.env.MOCK_AUTH === "true") {
      return NextResponse.json({
        success: true,
        user: {
          ...user,
          email: email || user.email,
          metadata: metadata || user.metadata,
        },
      });
    }

    // Build update payload - must include userId as first parameter
    const updateData: Record<string, string | Record<string, string>> = {};

    // Add email if it changed
    if (email && email !== user.email) {
      updateData.email = email;
    }

    // Always include metadata if provided
    if (metadata) {
      updateData.metadata = metadata;
    }

    // Update user via WorkOS API - userId is the first parameter
    const updatedUser = await workosClient.userManagement.updateUser(user.id, updateData);

    return NextResponse.json({
      success: true,
      user: updatedUser,
    });
  } catch (error) {
    console.error("Error updating user profile:", error);
    return NextResponse.json({ error: "Failed to update profile" }, { status: 500 });
  }
}
