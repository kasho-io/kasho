import {
  isValidImageMagicBytes,
  sanitizeFileExtension,
  generateCSRFToken,
  validateCSRFToken,
  sanitizeInput,
  isValidEmail,
} from "../security-utils";

describe("security-utils", () => {
  describe("isValidImageMagicBytes", () => {
    it("should validate JPEG magic bytes", () => {
      const jpegBytes = new Uint8Array([0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01]);
      expect(isValidImageMagicBytes(jpegBytes)).toBe(true);
    });

    it("should validate PNG magic bytes", () => {
      const pngBytes = new Uint8Array([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d]);
      expect(isValidImageMagicBytes(pngBytes)).toBe(true);
    });

    it("should validate GIF87a magic bytes", () => {
      const gif87Bytes = new Uint8Array([0x47, 0x49, 0x46, 0x38, 0x37, 0x61, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00]);
      expect(isValidImageMagicBytes(gif87Bytes)).toBe(true);
    });

    it("should validate GIF89a magic bytes", () => {
      const gif89Bytes = new Uint8Array([0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00]);
      expect(isValidImageMagicBytes(gif89Bytes)).toBe(true);
    });

    it("should validate WebP magic bytes", () => {
      const webpBytes = new Uint8Array([
        0x52,
        0x49,
        0x46,
        0x46, // RIFF
        0x00,
        0x00,
        0x00,
        0x00, // File size (placeholder)
        0x57,
        0x45,
        0x42,
        0x50, // WEBP
      ]);
      expect(isValidImageMagicBytes(webpBytes)).toBe(true);
    });

    it("should reject non-image bytes", () => {
      const pdfBytes = new Uint8Array([0x25, 0x50, 0x44, 0x46, 0x2d, 0x31, 0x2e, 0x34, 0x0a, 0x25, 0xc7, 0xec]);
      expect(isValidImageMagicBytes(pdfBytes)).toBe(false);
    });

    it("should reject empty buffer", () => {
      const emptyBytes = new Uint8Array([]);
      expect(isValidImageMagicBytes(emptyBytes)).toBe(false);
    });

    it("should reject buffer too small", () => {
      const smallBytes = new Uint8Array([0xff, 0xd8]);
      expect(isValidImageMagicBytes(smallBytes)).toBe(false);
    });

    it("should reject random bytes", () => {
      const randomBytes = new Uint8Array([0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b]);
      expect(isValidImageMagicBytes(randomBytes)).toBe(false);
    });

    it("should reject malformed WebP (correct RIFF but wrong signature)", () => {
      const badWebpBytes = new Uint8Array([
        0x52,
        0x49,
        0x46,
        0x46, // RIFF
        0x00,
        0x00,
        0x00,
        0x00, // File size
        0x00,
        0x00,
        0x00,
        0x00, // Wrong signature
      ]);
      expect(isValidImageMagicBytes(badWebpBytes)).toBe(false);
    });
  });

  describe("sanitizeFileExtension", () => {
    it("should return valid image extensions", () => {
      expect(sanitizeFileExtension("image.jpg")).toBe("jpg");
      expect(sanitizeFileExtension("photo.jpeg")).toBe("jpeg");
      expect(sanitizeFileExtension("pic.png")).toBe("png");
      expect(sanitizeFileExtension("animation.gif")).toBe("gif");
      expect(sanitizeFileExtension("modern.webp")).toBe("webp");
    });

    it("should handle uppercase extensions", () => {
      expect(sanitizeFileExtension("IMAGE.JPG")).toBe("jpg");
      expect(sanitizeFileExtension("PHOTO.JPEG")).toBe("jpeg");
      expect(sanitizeFileExtension("PIC.PNG")).toBe("png");
    });

    it("should handle mixed case extensions", () => {
      expect(sanitizeFileExtension("image.JpG")).toBe("jpg");
      expect(sanitizeFileExtension("photo.JpEg")).toBe("jpeg");
    });

    it("should handle path traversal attempts", () => {
      expect(sanitizeFileExtension("../../../etc/passwd.jpg")).toBe("jpg");
      expect(sanitizeFileExtension("..\\..\\windows\\system32\\config.png")).toBe("png");
      expect(sanitizeFileExtension("/etc/passwd/fake.gif")).toBe("gif");
    });

    it("should handle multiple dots in filename", () => {
      expect(sanitizeFileExtension("image.test.backup.jpg")).toBe("jpg");
      expect(sanitizeFileExtension("photo...png")).toBe("png");
    });

    it("should return null for non-image extensions", () => {
      expect(sanitizeFileExtension("document.pdf")).toBeNull();
      expect(sanitizeFileExtension("script.js")).toBeNull();
      expect(sanitizeFileExtension("data.txt")).toBeNull();
      expect(sanitizeFileExtension("video.mp4")).toBeNull();
    });

    it("should return null for files without extensions", () => {
      expect(sanitizeFileExtension("README")).toBeNull();
      expect(sanitizeFileExtension("image")).toBeNull();
      expect(sanitizeFileExtension("")).toBeNull();
    });

    it("should return null for invalid extensions", () => {
      expect(sanitizeFileExtension("image.jpg.exe")).toBeNull();
      expect(sanitizeFileExtension("photo.png.bat")).toBeNull();
    });

    it("should reject extensions with special characters", () => {
      // Extensions with special characters are considered invalid
      expect(sanitizeFileExtension("image.j!p@g")).toBeNull();
      expect(sanitizeFileExtension("photo.p$n%g")).toBeNull();
    });

    it("should handle edge cases", () => {
      expect(sanitizeFileExtension(".jpg")).toBe("jpg");
      expect(sanitizeFileExtension("..jpg")).toBe("jpg");
      expect(sanitizeFileExtension("...jpg")).toBe("jpg");
    });
  });

  describe("generateCSRFToken", () => {
    it("should generate a token of correct length", () => {
      const token = generateCSRFToken();
      expect(token).toHaveLength(64); // 32 bytes * 2 (hex encoding)
    });

    it("should generate unique tokens", () => {
      const tokens = new Set();
      for (let i = 0; i < 100; i++) {
        tokens.add(generateCSRFToken());
      }
      expect(tokens.size).toBe(100);
    });

    it("should generate hexadecimal strings", () => {
      const token = generateCSRFToken();
      expect(/^[0-9a-f]+$/.test(token)).toBe(true);
    });
  });

  describe("validateCSRFToken", () => {
    it("should validate matching tokens", () => {
      const token = "abc123def456";
      expect(validateCSRFToken(token, token)).toBe(true);
    });

    it("should reject non-matching tokens", () => {
      const token1 = "abc123def456";
      const token2 = "xyz789ghi012";
      expect(validateCSRFToken(token1, token2)).toBe(false);
    });

    it("should reject null tokens", () => {
      expect(validateCSRFToken(null, "token")).toBe(false);
      expect(validateCSRFToken("token", null)).toBe(false);
      expect(validateCSRFToken(null, null)).toBe(false);
    });

    it("should reject empty string tokens", () => {
      expect(validateCSRFToken("", "token")).toBe(false);
      expect(validateCSRFToken("token", "")).toBe(false);
    });

    it("should be case-sensitive", () => {
      expect(validateCSRFToken("ABC123", "abc123")).toBe(false);
    });
  });

  describe("sanitizeInput", () => {
    it("should escape HTML special characters", () => {
      expect(sanitizeInput("<script>alert('XSS')</script>")).toBe(
        "&lt;script&gt;alert(&#x27;XSS&#x27;)&lt;&#x2F;script&gt;",
      );
    });

    it("should escape less than symbol", () => {
      expect(sanitizeInput("<div>")).toBe("&lt;div&gt;");
    });

    it("should escape greater than symbol", () => {
      expect(sanitizeInput("5 > 3")).toBe("5 &gt; 3");
    });

    it("should escape double quotes", () => {
      expect(sanitizeInput('Hello "World"')).toBe("Hello &quot;World&quot;");
    });

    it("should escape single quotes", () => {
      expect(sanitizeInput("It's a test")).toBe("It&#x27;s a test");
    });

    it("should escape forward slashes", () => {
      expect(sanitizeInput("http://example.com/path")).toBe("http:&#x2F;&#x2F;example.com&#x2F;path");
    });

    it("should handle multiple special characters", () => {
      expect(sanitizeInput(`<img src="x" onerror='alert("XSS")'/>`)).toBe(
        `&lt;img src=&quot;x&quot; onerror=&#x27;alert(&quot;XSS&quot;)&#x27;&#x2F;&gt;`,
      );
    });

    it("should handle empty string", () => {
      expect(sanitizeInput("")).toBe("");
    });

    it("should handle normal text without changes", () => {
      expect(sanitizeInput("Hello World 123")).toBe("Hello World 123");
    });

    it("should handle all special characters in one string", () => {
      expect(sanitizeInput(`<>"'/`)).toBe(`&lt;&gt;&quot;&#x27;&#x2F;`);
    });
  });

  describe("isValidEmail", () => {
    it("should validate correct email formats", () => {
      expect(isValidEmail("user@example.com")).toBe(true);
      expect(isValidEmail("john.doe@company.co.uk")).toBe(true);
      expect(isValidEmail("test+tag@domain.org")).toBe(true);
      expect(isValidEmail("user123@test-domain.com")).toBe(true);
      expect(isValidEmail("a@b.co")).toBe(true);
    });

    it("should reject invalid email formats", () => {
      expect(isValidEmail("notanemail")).toBe(false);
      expect(isValidEmail("@example.com")).toBe(false);
      expect(isValidEmail("user@")).toBe(false);
      expect(isValidEmail("user@.com")).toBe(false);
      expect(isValidEmail("user..name@example.com")).toBe(false);
      expect(isValidEmail("user name@example.com")).toBe(false);
      expect(isValidEmail("user@exam ple.com")).toBe(false);
    });

    it("should reject emails with multiple @ symbols", () => {
      expect(isValidEmail("user@@example.com")).toBe(false);
      expect(isValidEmail("user@domain@example.com")).toBe(false);
    });

    it("should reject emails without domain extension", () => {
      expect(isValidEmail("user@localhost")).toBe(false);
      expect(isValidEmail("user@domain")).toBe(false);
    });

    it("should reject empty string", () => {
      expect(isValidEmail("")).toBe(false);
    });

    it("should reject emails longer than 255 characters", () => {
      const longEmail = "a".repeat(250) + "@test.com";
      expect(isValidEmail(longEmail)).toBe(false);
    });

    it("should accept emails exactly 255 characters", () => {
      const maxEmail = "a".repeat(246) + "@test.com"; // 246 + 9 = 255
      expect(isValidEmail(maxEmail)).toBe(true);
    });

    it("should handle edge cases", () => {
      expect(isValidEmail("   user@example.com   ")).toBe(false); // Has spaces
      expect(isValidEmail("user@example..com")).toBe(false); // Double dots
      expect(isValidEmail(".user@example.com")).toBe(false); // Starts with dot
      expect(isValidEmail("user.@example.com")).toBe(false); // Ends with dot before @
    });
  });
});
