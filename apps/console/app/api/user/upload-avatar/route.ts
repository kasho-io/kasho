import { NextRequest, NextResponse } from "next/server";
import { withAuth } from "@/lib/auth-wrapper";
import { put } from "@vercel/blob";

export async function POST(request: NextRequest) {
  try {
    // Get the authenticated user
    const { user } = await withAuth();

    if (!user) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // In test mode, just return a mock URL
    if (process.env.NODE_ENV === "test" || process.env.MOCK_AUTH === "true") {
      return NextResponse.json({
        success: true,
        url: "https://example.com/mock-avatar.jpg",
      });
    }

    // Get the file from the request
    const formData = await request.formData();
    const file = formData.get("file") as File;

    if (!file) {
      return NextResponse.json({ error: "No file provided" }, { status: 400 });
    }

    // Validate file type
    if (!file.type.startsWith("image/")) {
      return NextResponse.json({ error: "File must be an image" }, { status: 400 });
    }

    // Validate file size (4.5MB limit for Vercel Blob)
    if (file.size > 4.5 * 1024 * 1024) {
      return NextResponse.json({ error: "File size must be less than 4.5MB" }, { status: 400 });
    }

    // Generate a unique filename
    const timestamp = Date.now();
    const fileExtension = file.name.split(".").pop();
    const filename = `avatars/${user.id}-${timestamp}.${fileExtension}`;

    // Upload to Vercel Blob Storage
    const blob = await put(filename, file, {
      access: "public",
      addRandomSuffix: false,
    });

    return NextResponse.json({
      success: true,
      url: blob.url,
    });
  } catch (error) {
    console.error("Error uploading avatar:", error);
    return NextResponse.json({ error: "Failed to upload avatar" }, { status: 500 });
  }
}
