import { NextRequest, NextResponse } from "next/server";
import { withAuth } from "@workos-inc/authkit-nextjs";
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

    // Build update payload
    const updatePayload: Record<string, string | Record<string, string>> = {
      userId: user.id,
    };

    // Add email if it changed
    if (email && email !== user.email) {
      updatePayload.email = email;
    }

    // Add metadata if provided
    if (metadata) {
      updatePayload.metadata = metadata;
    }

    // Update user via WorkOS API
    const updatedUser = await workosClient.userManagement.updateUser(updatePayload);

    return NextResponse.json({
      success: true,
      user: updatedUser,
    });
  } catch (error) {
    console.error("Error updating user profile:", error);
    return NextResponse.json({ error: "Failed to update profile" }, { status: 500 });
  }
}
