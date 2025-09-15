import { NextRequest, NextResponse } from "next/server";
import { services } from "@/lib/services";
import { isValidImageMagicBytes, sanitizeFileExtension } from "@/lib/security-utils";
import { fileUploadSchema } from "@/lib/validation-schemas";
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
    const { user } = await services.workos.withAuth();

    if (!user) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // Get the file from the request
    const formData = await request.formData();
    const file = formData.get("file") as File;

    if (!file) {
      return NextResponse.json({ error: "No file provided" }, { status: 400 });
    }

    // Validate file metadata with Zod
    const fileValidation = fileUploadSchema.safeParse({
      size: file.size,
      type: file.type,
      name: file.name,
    });

    if (!fileValidation.success) {
      return NextResponse.json({ error: fileValidation.error.issues[0].message }, { status: 400 });
    }

    // Read file content for magic bytes validation
    const buffer = await file.arrayBuffer();
    const uint8Array = new Uint8Array(buffer);

    // Validate actual file content (magic bytes) to prevent MIME type spoofing
    if (!isValidImageMagicBytes(uint8Array)) {
      return NextResponse.json(
        { error: "Invalid image file format. File content does not match image type." },
        { status: 400 },
      );
    }

    // Sanitize and validate file extension
    const sanitizedExtension = sanitizeFileExtension(file.name);
    if (!sanitizedExtension) {
      return NextResponse.json({ error: "Invalid or unsupported file extension" }, { status: 400 });
    }

    // Generate a secure unique filename with sanitized extension
    const timestamp = Date.now();
    const randomSuffix = Math.random().toString(36).substring(2, 8);
    const filename = `avatars/${user.id}-${timestamp}-${randomSuffix}.${sanitizedExtension}`;

    // Convert buffer back to Blob for upload (after validation)
    const validatedBlob = new Blob([buffer], { type: file.type });

    // Upload to Vercel Blob Storage
    const blob = await services.vercelBlob.upload(filename, validatedBlob, {
      access: "public",
      addRandomSuffix: false,
      contentType: file.type,
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
