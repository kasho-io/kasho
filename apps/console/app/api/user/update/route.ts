import { NextRequest, NextResponse } from "next/server";
import { withAuth } from "@/lib/auth-wrapper";
import { workosClient } from "@/lib/workos-client";
import { refreshSession } from "@workos-inc/authkit-nextjs";
import { customMetadataSchema } from "@/lib/validation-schemas";
import { z } from "zod";
import { isRequestValid } from "@/lib/csrf-protection";

export async function POST(request: NextRequest) {
  try {
    // Validate CSRF protection
    const requestValidation = isRequestValid(request, {
      allowedMethods: ["POST"],
    });

    if (!requestValidation.valid) {
      return NextResponse.json({ error: requestValidation.error }, { status: 403 });
    }

    // Get the authenticated user
    const { user } = await withAuth();

    if (!user) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // Parse the request body
    const body = await request.json();

    // Create a validation schema for the request body
    const requestSchema = z.object({
      email: z.string().email("Invalid email format").max(255).optional(),
      metadata: customMetadataSchema.optional(),
    });

    // Validate request body
    const validation = requestSchema.safeParse(body);
    if (!validation.success) {
      return NextResponse.json({ error: validation.error.issues[0].message }, { status: 400 });
    }

    const { email, metadata } = validation.data;

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
    // Using the actual WorkOS SDK type
    const updatePayload: Parameters<typeof workosClient.userManagement.updateUser>[0] = {
      userId: user.id,
    };

    // Always include email if provided
    // Don't compare with session user.email as it may be stale after previous updates
    if (email) {
      updatePayload.email = email;
    }

    // Always include metadata if provided
    // Convert undefined values to null for WorkOS API compatibility
    if (metadata) {
      const workosMetadata: Record<string, string | null> = {};
      for (const [key, value] of Object.entries(metadata)) {
        workosMetadata[key] = value ?? null;
      }
      updatePayload.metadata = workosMetadata;
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
