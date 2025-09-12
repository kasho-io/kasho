import { userUpdateSchema, customMetadataSchema, fileUploadSchema } from "../validation-schemas";

describe("validation-schemas", () => {
  describe("userUpdateSchema", () => {
    it("should validate a valid user update request", () => {
      const validData = {
        email: "test@example.com",
        firstName: "John",
        lastName: "Doe",
        emailVerified: true,
        externalId: "ext123",
        metadata: {
          role: "admin",
          department: "engineering",
        },
      };

      const result = userUpdateSchema.safeParse(validData);
      expect(result.success).toBe(true);
    });

    it("should accept partial updates", () => {
      const partialData = {
        email: "test@example.com",
      };

      const result = userUpdateSchema.safeParse(partialData);
      expect(result.success).toBe(true);
    });

    it("should accept null values for nullable fields", () => {
      const dataWithNulls = {
        firstName: null,
        lastName: null,
        externalId: null,
      };

      const result = userUpdateSchema.safeParse(dataWithNulls);
      expect(result.success).toBe(true);
    });

    it("should reject invalid email format", () => {
      const invalidEmail = {
        email: "not-an-email",
      };

      const result = userUpdateSchema.safeParse(invalidEmail);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("Invalid email format");
      }
    });

    it("should reject email longer than 255 characters", () => {
      const longEmail = {
        email: "a".repeat(250) + "@test.com",
      };

      const result = userUpdateSchema.safeParse(longEmail);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("Email must be less than 255 characters");
      }
    });

    it("should reject firstName longer than 100 characters", () => {
      const longFirstName = {
        firstName: "a".repeat(101),
      };

      const result = userUpdateSchema.safeParse(longFirstName);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("First name must be less than 100 characters");
      }
    });

    it("should reject lastName longer than 100 characters", () => {
      const longLastName = {
        lastName: "a".repeat(101),
      };

      const result = userUpdateSchema.safeParse(longLastName);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("Last name must be less than 100 characters");
      }
    });

    it("should reject externalId longer than 64 characters", () => {
      const longExternalId = {
        externalId: "a".repeat(65),
      };

      const result = userUpdateSchema.safeParse(longExternalId);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("External ID must be less than 64 characters");
      }
    });

    describe("metadata validation", () => {
      it("should accept valid metadata", () => {
        const validMetadata = {
          metadata: {
            key1: "value1",
            key2: "value2",
            key3: null,
          },
        };

        const result = userUpdateSchema.safeParse(validMetadata);
        expect(result.success).toBe(true);
      });

      it("should reject metadata with more than 10 keys", () => {
        const tooManyKeys: Record<string, string> = {};
        for (let i = 0; i < 11; i++) {
          tooManyKeys[`key${i}`] = "value";
        }

        const result = userUpdateSchema.safeParse({ metadata: tooManyKeys });
        expect(result.success).toBe(false);
        if (!result.success) {
          expect(result.error.issues[0].message).toBe("Metadata must have max 10 keys, each up to 40 ASCII characters");
        }
      });

      it("should reject metadata with keys longer than 40 characters", () => {
        const longKey = {
          metadata: {
            ["a".repeat(41)]: "value",
          },
        };

        const result = userUpdateSchema.safeParse(longKey);
        expect(result.success).toBe(false);
        if (!result.success) {
          expect(result.error.issues[0].message).toBe("Metadata must have max 10 keys, each up to 40 ASCII characters");
        }
      });

      it("should reject metadata with non-ASCII keys", () => {
        const nonAsciiKey = {
          metadata: {
            "émoji🎉": "value",
          },
        };

        const result = userUpdateSchema.safeParse(nonAsciiKey);
        expect(result.success).toBe(false);
        if (!result.success) {
          expect(result.error.issues[0].message).toBe("Metadata must have max 10 keys, each up to 40 ASCII characters");
        }
      });

      it("should reject metadata values longer than 500 characters", () => {
        const longValue = {
          metadata: {
            key: "a".repeat(501),
          },
        };

        const result = userUpdateSchema.safeParse(longValue);
        expect(result.success).toBe(false);
        if (!result.success) {
          expect(result.error.issues[0].message).toBe("Metadata value must be less than 500 characters");
        }
      });

      it("should accept null metadata values", () => {
        const nullValue = {
          metadata: {
            key: null,
          },
        };

        const result = userUpdateSchema.safeParse(nullValue);
        expect(result.success).toBe(true);
      });
    });
  });

  describe("customMetadataSchema", () => {
    it("should validate valid custom metadata", () => {
      const validData = {
        first_name: "John",
        last_name: "Doe",
        profile_picture_url: "https://example.com/pic.jpg",
      };

      const result = customMetadataSchema.safeParse(validData);
      expect(result.success).toBe(true);
    });

    it("should accept partial custom metadata", () => {
      const partialData = {
        first_name: "John",
      };

      const result = customMetadataSchema.safeParse(partialData);
      expect(result.success).toBe(true);
    });

    it("should accept empty object", () => {
      const emptyData = {};

      const result = customMetadataSchema.safeParse(emptyData);
      expect(result.success).toBe(true);
    });

    it("should reject first_name longer than 100 characters", () => {
      const longFirstName = {
        first_name: "a".repeat(101),
      };

      const result = customMetadataSchema.safeParse(longFirstName);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("First name must be less than 100 characters");
      }
    });

    it("should reject last_name longer than 100 characters", () => {
      const longLastName = {
        last_name: "a".repeat(101),
      };

      const result = customMetadataSchema.safeParse(longLastName);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("Last name must be less than 100 characters");
      }
    });

    it("should reject invalid URL format", () => {
      const invalidUrl = {
        profile_picture_url: "not-a-url",
      };

      const result = customMetadataSchema.safeParse(invalidUrl);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("Invalid URL format");
      }
    });

    it("should reject URL longer than 500 characters", () => {
      const longUrl = {
        profile_picture_url: "https://example.com/" + "a".repeat(481), // Total: 501 characters
      };

      const result = customMetadataSchema.safeParse(longUrl);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("URL must be less than 500 characters");
      }
    });
  });

  describe("fileUploadSchema", () => {
    it("should validate valid file upload", () => {
      const validFile = {
        size: 1024 * 1024, // 1MB
        type: "image/jpeg",
        name: "photo.jpg",
      };

      const result = fileUploadSchema.safeParse(validFile);
      expect(result.success).toBe(true);
    });

    it("should accept various image types", () => {
      const imageTypes = ["image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/JPEG", "image/PNG"];

      imageTypes.forEach((type) => {
        const file = {
          size: 1024,
          type,
          name: "image.ext",
        };
        const result = fileUploadSchema.safeParse(file);
        expect(result.success).toBe(true);
      });
    });

    it("should reject files larger than 4.5MB", () => {
      const largeFile = {
        size: 5 * 1024 * 1024, // 5MB
        type: "image/jpeg",
        name: "large.jpg",
      };

      const result = fileUploadSchema.safeParse(largeFile);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("File size must be less than 4.5MB");
      }
    });

    it("should accept files exactly 4.5MB", () => {
      const maxSizeFile = {
        size: 4.5 * 1024 * 1024, // Exactly 4.5MB
        type: "image/jpeg",
        name: "max.jpg",
      };

      const result = fileUploadSchema.safeParse(maxSizeFile);
      expect(result.success).toBe(true);
    });

    it("should reject non-image files", () => {
      const nonImageFile = {
        size: 1024,
        type: "application/pdf",
        name: "document.pdf",
      };

      const result = fileUploadSchema.safeParse(nonImageFile);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("File must be an image");
      }
    });

    it("should reject files without a name", () => {
      const noNameFile = {
        size: 1024,
        type: "image/jpeg",
        name: "",
      };

      const result = fileUploadSchema.safeParse(noNameFile);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("File name is required");
      }
    });

    it("should reject files with name longer than 255 characters", () => {
      const longNameFile = {
        size: 1024,
        type: "image/jpeg",
        name: "a".repeat(256),
      };

      const result = fileUploadSchema.safeParse(longNameFile);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toBe("File name must be less than 255 characters");
      }
    });
  });
});
