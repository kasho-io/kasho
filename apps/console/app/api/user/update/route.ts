import { NextRequest, NextResponse } from "next/server";
import { withAuth } from "@/lib/auth-wrapper";
import { workosClient } from "@/lib/workos-client";
import { refreshSession } from "@workos-inc/authkit-nextjs";

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

    // Build update payload for WorkOS API
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const updatePayload: any = {
      userId: user.id,
    };

    // Always include email if provided
    // Don't compare with session user.email as it may be stale after previous updates
    if (email) {
      updatePayload.email = email;
    }

    // Always include metadata if provided
    if (metadata) {
      updatePayload.metadata = metadata;
    }

    // Update user via WorkOS API - expects an object with userId and other fields
    const updatedUser = await workosClient.userManagement.updateUser(updatePayload);

    // Refresh the session to get the updated user data
    // This ensures the cached session is updated with the new email
    try {
      await refreshSession();
    } catch (refreshError) {
      console.error("Failed to refresh session:", refreshError);
      // Continue even if refresh fails - the update was successful
    }

    return NextResponse.json({
      success: true,
      user: updatedUser,
    });
  } catch (error) {
    console.error("Error updating user profile:", error);
    return NextResponse.json({ error: "Failed to update profile" }, { status: 500 });
  }
}
