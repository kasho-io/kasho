import { z } from "zod";

/**
 * WorkOS Metadata constraints:
 * - Up to 10 key-value pairs
 * - Keys: Up to 40 characters, ASCII only
 * - Values: Up to 500 characters
 */
const workosMetadataValueSchema = z.string().max(500, "Metadata value must be less than 500 characters").nullable();

const workosMetadataSchema = z.record(z.string(), workosMetadataValueSchema).refine(
  (metadata) => {
    const keys = Object.keys(metadata);
    // Check max 10 keys
    if (keys.length > 10) return false;
    // Check each key: max 40 chars, ASCII only
    return keys.every((key) => key.length <= 40 && /^[\x00-\x7F]*$/.test(key));
  },
  {
    message: "Metadata must have max 10 keys, each up to 40 ASCII characters",
  },
);

/**
 * Schema for user profile update request
 * Based on WorkOS UpdateUserOptions interface
 */
export const userUpdateSchema = z.object({
  email: z
    .string()
    .email("Invalid email format")
    .min(1, "Email is required")
    .max(255, "Email must be less than 255 characters")
    .optional(),
  firstName: z.string().max(100, "First name must be less than 100 characters").nullable().optional(),
  lastName: z.string().max(100, "Last name must be less than 100 characters").nullable().optional(),
  emailVerified: z.boolean().optional(),
  externalId: z.string().max(64, "External ID must be less than 64 characters").nullable().optional(),
  metadata: workosMetadataSchema.optional(),
});

export type UserUpdateRequest = z.infer<typeof userUpdateSchema>;

/**
 * Schema for our custom metadata fields (stored within WorkOS metadata)
 */
export const customMetadataSchema = z.object({
  first_name: z.string().max(100, "First name must be less than 100 characters").optional(),
  last_name: z.string().max(100, "Last name must be less than 100 characters").optional(),
  profile_picture_url: z.string().url("Invalid URL format").max(500, "URL must be less than 500 characters").optional(),
});

export type CustomMetadata = z.infer<typeof customMetadataSchema>;

/**
 * Schema for file upload validation
 */
export const fileUploadSchema = z.object({
  size: z.number().max(4.5 * 1024 * 1024, "File size must be less than 4.5MB"),
  type: z.string().regex(/^image\/(jpeg|jpg|png|gif|webp)$/i, "File must be an image"),
  name: z.string().min(1, "File name is required").max(255, "File name must be less than 255 characters"),
});

export type FileUploadValidation = z.infer<typeof fileUploadSchema>;

/**
 * Complete WorkOS User type based on their interface
 */
export interface WorkOSUser {
  object: "user";
  id: string;
  email: string;
  emailVerified: boolean;
  profilePictureUrl: string | null;
  firstName: string | null;
  lastName: string | null;
  lastSignInAt: string | null;
  createdAt: string;
  updatedAt: string;
  externalId: string | null;
  metadata: Record<string, string>;
}

/**
 * WorkOS UpdateUserOptions type based on their interface
 */
export interface WorkOSUpdateUserOptions {
  userId: string;
  email?: string;
  firstName?: string;
  lastName?: string;
  emailVerified?: boolean;
  password?: string;
  passwordHash?: string;
  passwordHashType?: string;
  externalId?: string;
  metadata?: Record<string, string | null>;
}
