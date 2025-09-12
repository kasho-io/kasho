import crypto from "crypto";

/**
 * Validates image file content by checking magic bytes (file signatures)
 * This prevents MIME type spoofing attacks
 */
export function isValidImageMagicBytes(buffer: Uint8Array): boolean {
  if (buffer.length < 12) return false;

  // Check for common image format magic bytes
  const magicBytes = {
    // JPEG: FF D8 FF
    jpeg: [0xff, 0xd8, 0xff],
    // PNG: 89 50 4E 47 0D 0A 1A 0A
    png: [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a],
    // GIF87a: 47 49 46 38 37 61
    gif87a: [0x47, 0x49, 0x46, 0x38, 0x37, 0x61],
    // GIF89a: 47 49 46 38 39 61
    gif89a: [0x47, 0x49, 0x46, 0x38, 0x39, 0x61],
    // WebP: 52 49 46 46 ?? ?? ?? ?? 57 45 42 50
    webpRiff: [0x52, 0x49, 0x46, 0x46],
    webpSignature: [0x57, 0x45, 0x42, 0x50],
  };

  // Check JPEG
  if (matchesMagicBytes(buffer, magicBytes.jpeg, 0)) {
    return true;
  }

  // Check PNG
  if (matchesMagicBytes(buffer, magicBytes.png, 0)) {
    return true;
  }

  // Check GIF
  if (matchesMagicBytes(buffer, magicBytes.gif87a, 0) || matchesMagicBytes(buffer, magicBytes.gif89a, 0)) {
    return true;
  }

  // Check WebP (RIFF header + WEBP signature at offset 8)
  if (matchesMagicBytes(buffer, magicBytes.webpRiff, 0) && matchesMagicBytes(buffer, magicBytes.webpSignature, 8)) {
    return true;
  }

  return false;
}

/**
 * Helper function to match magic bytes at a specific offset
 */
function matchesMagicBytes(buffer: Uint8Array, magicBytes: number[], offset: number): boolean {
  if (buffer.length < offset + magicBytes.length) return false;

  for (let i = 0; i < magicBytes.length; i++) {
    if (buffer[offset + i] !== magicBytes[i]) {
      return false;
    }
  }
  return true;
}

/**
 * Sanitizes file extension to prevent path traversal attacks
 */
export function sanitizeFileExtension(filename: string): string | null {
  const allowedExtensions = ["jpg", "jpeg", "png", "gif", "webp"];

  // Remove any path components and get just the filename
  const baseFilename = filename.split(/[/\\]/).pop() || "";

  // Get the extension (last dot-separated part)
  const parts = baseFilename.split(".");
  if (parts.length < 2) return null;

  const extension = parts.pop()?.toLowerCase();
  if (!extension || !allowedExtensions.includes(extension)) {
    return null;
  }

  // Remove any special characters from extension
  return extension.replace(/[^a-z]/g, "");
}

/**
 * Generates a CSRF token
 */
export function generateCSRFToken(): string {
  return crypto.randomBytes(32).toString("hex");
}

/**
 * Validates a CSRF token
 */
export function validateCSRFToken(token: string | null, sessionToken: string | null): boolean {
  if (!token || !sessionToken) return false;
  return token === sessionToken;
}

/**
 * Sanitizes user input to prevent XSS attacks
 * For basic text fields - more complex HTML should use a library like DOMPurify
 */
export function sanitizeInput(input: string): string {
  return input
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#x27;")
    .replace(/\//g, "&#x2F;");
}

/**
 * Validates email format
 */
export function isValidEmail(email: string): boolean {
  // More comprehensive email regex that handles edge cases including + tags
  const emailRegex = /^[a-zA-Z0-9][a-zA-Z0-9._+-]*[a-zA-Z0-9]@[a-zA-Z0-9][a-zA-Z0-9.-]*[a-zA-Z0-9]\.[a-zA-Z]{2,}$/;

  // Basic length check
  if (email.length > 255) return false;

  // Check for spaces
  if (email.includes(" ")) return false;

  // Check for double dots
  if (email.includes("..")) return false;

  // Check if starts or ends with dot
  if (email.startsWith(".") || email.endsWith(".")) return false;

  // Check if has dot right before or after @
  if (email.includes(".@") || email.includes("@.")) return false;

  // For very short emails like a@b.co
  if (email.length <= 6) {
    const simpleRegex = /^[a-zA-Z0-9]+@[a-zA-Z0-9]+\.[a-zA-Z]{2,}$/;
    return simpleRegex.test(email);
  }

  // Handle special case where email ends with + or .
  const localPart = email.split("@")[0];
  if (localPart.endsWith("+") || localPart.endsWith(".")) {
    return false;
  }

  return emailRegex.test(email);
}
