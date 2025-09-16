"use client";

import { useState, useRef, useEffect } from "react";
import Image from "next/image";

interface ProfileData {
  id: string;
  email: string;
  emailVerified: boolean;
  firstName: string;
  lastName: string;
  profilePictureUrl: string;
}

interface ProfileFormProps {
  initialData: ProfileData;
}

export default function ProfileForm({ initialData }: ProfileFormProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const messageRef = useRef<HTMLDivElement>(null);
  const [formData, setFormData] = useState(initialData);
  const [originalData, setOriginalData] = useState(initialData);
  const [isSaving, setIsSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [previewUrl, setPreviewUrl] = useState(initialData.profilePictureUrl);

  // Focus management for dynamic messages
  useEffect(() => {
    if (message && messageRef.current) {
      // Announce to screen readers and focus for keyboard users
      messageRef.current.focus();
    }
  }, [message]);

  // Check if there are any changes
  const hasChanges =
    formData.email !== originalData.email ||
    formData.firstName !== originalData.firstName ||
    formData.lastName !== originalData.lastName ||
    formData.profilePictureUrl !== originalData.profilePictureUrl;

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => {
      const updated = { ...prev, [name]: value };
      // Handle email verification state based on whether email matches original
      if (name === "email") {
        if (value !== originalData.email) {
          // Email changed from original - mark as unverified
          updated.emailVerified = false;
        } else {
          // Email reverted to original - restore original verification state
          updated.emailVerified = originalData.emailVerified;
        }
      }
      return updated;
    });
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.type.startsWith("image/")) {
      setMessage({ type: "error", text: "Please select an image file" });
      return;
    }

    // Validate file size (4.5MB limit for Vercel Blob)
    if (file.size > 4.5 * 1024 * 1024) {
      setMessage({ type: "error", text: "Image must be less than 4.5MB" });
      return;
    }

    // Create preview
    const reader = new FileReader();
    reader.onloadend = () => {
      setPreviewUrl(reader.result as string);
    };
    reader.readAsDataURL(file);

    // Upload the file
    try {
      const formData = new FormData();
      formData.append("file", file);

      const response = await fetch("/api/user/upload-avatar", {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        throw new Error("Failed to upload image");
      }

      const { url } = await response.json();
      // Only update the form data, not the original data
      // This ensures the "Save Changes" button becomes enabled
      setFormData((prev) => ({ ...prev, profilePictureUrl: url }));
      setPreviewUrl(url);
      setMessage({ type: "success", text: "Image uploaded successfully. Click 'Save Changes' to save your profile." });
    } catch (error) {
      // On upload failure, revert preview to current saved image
      setPreviewUrl(originalData.profilePictureUrl);
      setMessage({ type: "error", text: "Failed to upload image" });
      console.error("Upload error:", error);
      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    setMessage(null);

    try {
      const response = await fetch("/api/user/update", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email: formData.email,
          metadata: {
            first_name: formData.firstName,
            last_name: formData.lastName,
            profile_picture_url: formData.profilePictureUrl,
          },
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to update profile");
      }

      const result = await response.json();
      const updatedUser = result.user;

      // Update both formData and originalData with the server response
      // IMPORTANT: WorkOS stores custom fields in metadata, not as direct properties
      // The firstName/lastName direct properties are immutable and set during user creation
      // We must use metadata for user-editable fields
      const updatedData = {
        id: updatedUser.id || formData.id,
        email: updatedUser.email || formData.email,
        emailVerified: updatedUser.emailVerified ?? formData.emailVerified,
        // Use metadata values if they exist, otherwise keep form values
        // WorkOS doesn't update firstName/lastName direct properties
        firstName: updatedUser.metadata?.first_name || formData.firstName,
        lastName: updatedUser.metadata?.last_name || formData.lastName,
        profilePictureUrl: updatedUser.metadata?.profile_picture_url || formData.profilePictureUrl,
      };

      setFormData(updatedData);
      setOriginalData(updatedData);
      setMessage({ type: "success", text: "Profile updated successfully" });

      // Don't refresh - WorkOS session cache doesn't update immediately
      // The form now has the correct data from the API response
    } catch (error) {
      // On save failure, keep the form data as-is but show error
      // This preserves unsaved changes including uploaded avatar
      setMessage({ type: "error", text: "Failed to update profile. Your changes have not been saved." });
      console.error("Update error:", error);
      // Note: We intentionally don't revert formData here to preserve user's work
      // The hasChanges flag will remain true, allowing retry
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      {!formData.emailVerified && (
        <div className="alert alert-warning">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-6 w-6 shrink-0 stroke-current"
            fill="none"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
          <div>
            <h3 className="font-bold">Email Verification Required</h3>
            <div className="text-sm">
              Your email address needs to be verified. Please sign out and sign in again to receive a verification code.
            </div>
          </div>
        </div>
      )}

      {message && (
        <div
          ref={messageRef}
          className={`alert ${message.type === "success" ? "alert-success" : "alert-error"}`}
          role="alert"
          aria-live="polite"
          tabIndex={-1}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-6 w-6 shrink-0 stroke-current"
            fill="none"
            viewBox="0 0 24 24"
          >
            {message.type === "success" ? (
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            ) : (
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            )}
          </svg>
          <span>{message.text}</span>
        </div>
      )}

      <div className="card bg-base-100 shadow-xl">
        <div className="card-body p-8">
          <form onSubmit={handleSubmit}>
            {/* Profile Picture Section */}
            <div className="mb-8 pb-8 border-b border-base-200">
              <h3 className="text-lg font-semibold mb-4">Profile Picture</h3>
              <div className="flex items-center gap-6">
                <div className="avatar">
                  <div className="w-32 h-32 rounded-full bg-base-200 ring ring-base-300 ring-offset-base-100 ring-offset-2">
                    {previewUrl ? (
                      <Image
                        src={previewUrl}
                        alt="Profile"
                        width={128}
                        height={128}
                        className="rounded-full object-cover"
                      />
                    ) : (
                      <div className="flex items-center justify-center h-full text-4xl font-semibold text-base-content/70">
                        {formData.firstName?.[0]?.toUpperCase() || formData.email?.[0]?.toUpperCase() || "U"}
                      </div>
                    )}
                  </div>
                </div>
                <div className="space-y-2">
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*"
                    onChange={handleFileChange}
                    className="hidden"
                    aria-label="Upload profile picture"
                    id="profile-picture-upload"
                  />
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    className="btn btn-outline btn-sm"
                    aria-label="Change profile picture"
                    aria-describedby="profile-picture-help"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-4 w-4 mr-2"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                      />
                    </svg>
                    Change Picture
                  </button>
                  <p id="profile-picture-help" className="text-xs text-base-content/60">
                    JPG, PNG, GIF or WebP. Max 4.5MB.
                  </p>
                </div>
              </div>
            </div>

            {/* Personal Information Section */}
            <div className="space-y-1 mb-6">
              <h3 className="text-lg font-semibold">Personal Information</h3>
              <p className="text-sm text-base-content/60">Update your personal details and profile information.</p>
            </div>

            {/* Email */}
            <div className="form-control w-full mb-6">
              <label className="label">
                <span className="label-text font-medium">Email Address</span>
                {!formData.emailVerified && (
                  <span className="badge badge-warning badge-sm ml-2" role="status" aria-label="Email not verified">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-3 w-3 mr-1"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                      />
                    </svg>
                    Unverified
                  </span>
                )}
              </label>
              <input
                type="email"
                name="email"
                value={formData.email}
                onChange={handleInputChange}
                className="input input-bordered w-full"
                placeholder="your@email.com"
                required
                aria-label="Email address"
                aria-describedby="email-help"
                aria-invalid={!formData.emailVerified ? "true" : "false"}
              />
              <label className="label">
                <span id="email-help" className="label-text-alt text-base-content/60">
                  {!formData.emailVerified
                    ? "You will need to verify your email the next time you sign in."
                    : formData.email !== originalData.email
                      ? "You'll need to verify your new email address after saving"
                      : ""}
                </span>
              </label>
            </div>

            {/* Name Fields - Side by side on larger screens */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
              <div className="form-control w-full">
                <label className="label">
                  <span className="label-text font-medium">First Name</span>
                </label>
                <input
                  type="text"
                  name="firstName"
                  value={formData.firstName}
                  onChange={handleInputChange}
                  className="input input-bordered w-full"
                  placeholder="John"
                  aria-label="First name"
                />
              </div>

              <div className="form-control w-full">
                <label className="label">
                  <span className="label-text font-medium">Last Name</span>
                </label>
                <input
                  type="text"
                  name="lastName"
                  value={formData.lastName}
                  onChange={handleInputChange}
                  className="input input-bordered w-full"
                  placeholder="Doe"
                  aria-label="Last name"
                />
              </div>
            </div>

            {/* Action Buttons */}
            <div className="divider my-8"></div>
            <div className="flex justify-end">
              <button type="submit" disabled={!hasChanges || isSaving} className="btn btn-primary">
                {isSaving ? (
                  <>
                    <span className="loading loading-spinner loading-sm"></span>
                    Saving...
                  </>
                ) : (
                  <>Save Changes</>
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
