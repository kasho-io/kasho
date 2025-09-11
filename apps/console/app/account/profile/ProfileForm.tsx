"use client";

import { useState, useRef } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";

interface ProfileData {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  profilePictureUrl: string;
  bio: string;
  location: string;
}

interface ProfileFormProps {
  initialData: ProfileData;
}

export default function ProfileForm({ initialData }: ProfileFormProps) {
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [formData, setFormData] = useState(initialData);
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [previewUrl, setPreviewUrl] = useState(initialData.profilePictureUrl);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
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
      setFormData((prev) => ({ ...prev, profilePictureUrl: url }));
      setPreviewUrl(url);
      setMessage({ type: "success", text: "Image uploaded successfully" });
    } catch (error) {
      setMessage({ type: "error", text: "Failed to upload image" });
      console.error("Upload error:", error);
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
          metadata: {
            first_name: formData.firstName,
            last_name: formData.lastName,
            profile_picture_url: formData.profilePictureUrl,
            bio: formData.bio,
            location: formData.location,
          },
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to update profile");
      }

      setMessage({ type: "success", text: "Profile updated successfully" });
      setIsEditing(false);
      router.refresh();
    } catch (error) {
      setMessage({ type: "error", text: "Failed to update profile" });
      console.error("Update error:", error);
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = () => {
    setFormData(initialData);
    setPreviewUrl(initialData.profilePictureUrl);
    setIsEditing(false);
    setMessage(null);
  };

  return (
    <div className="space-y-6">
      {message && (
        <div className={`alert ${message.type === "success" ? "alert-success" : "alert-error"}`}>
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
                {isEditing && (
                  <div className="space-y-2">
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept="image/*"
                      onChange={handleFileChange}
                      className="hidden"
                    />
                    <button
                      type="button"
                      onClick={() => fileInputRef.current?.click()}
                      className="btn btn-outline btn-sm"
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
                      Upload New Picture
                    </button>
                    <p className="text-xs text-base-content/60">JPG, PNG, GIF or WebP. Max 4.5MB.</p>
                  </div>
                )}
              </div>
            </div>

            {/* Personal Information Section */}
            <div className="space-y-1 mb-6">
              <h3 className="text-lg font-semibold">Personal Information</h3>
              <p className="text-sm text-base-content/60">Update your personal details and profile information.</p>
            </div>

            {/* Email (read-only) */}
            <div className="form-control w-full mb-6">
              <label className="label">
                <span className="label-text font-medium">Email Address</span>
              </label>
              <input
                type="email"
                value={formData.email}
                disabled
                className="input input-bordered w-full bg-base-200 cursor-not-allowed"
              />
              <label className="label">
                <span className="label-text-alt text-base-content/60">
                  Contact support to change your email address
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
                  disabled={!isEditing}
                  className="input input-bordered w-full"
                  placeholder="John"
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
                  disabled={!isEditing}
                  className="input input-bordered w-full"
                  placeholder="Doe"
                />
              </div>
            </div>

            {/* Location */}
            <div className="form-control w-full mb-6">
              <label className="label">
                <span className="label-text font-medium">Location</span>
              </label>
              <input
                type="text"
                name="location"
                value={formData.location}
                onChange={handleInputChange}
                disabled={!isEditing}
                className="input input-bordered w-full"
                placeholder="San Francisco, CA"
              />
              <label className="label">
                <span className="label-text-alt text-base-content/60">Your city and state/country</span>
              </label>
            </div>

            {/* Bio */}
            <div className="form-control w-full mb-8">
              <label className="label">
                <span className="label-text font-medium">Bio</span>
              </label>
              <textarea
                name="bio"
                value={formData.bio}
                onChange={handleInputChange}
                disabled={!isEditing}
                className="textarea textarea-bordered w-full h-32 resize-none"
                placeholder="Tell us a bit about yourself..."
              />
              <label className="label">
                <span className="label-text-alt text-base-content/60">{formData.bio.length}/500 characters</span>
              </label>
            </div>

            {/* Action Buttons */}
            <div className="divider my-8"></div>
            <div className="flex justify-end gap-3">
              {!isEditing ? (
                <button type="button" onClick={() => setIsEditing(true)} className="btn btn-primary">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                    />
                  </svg>
                  Edit Profile
                </button>
              ) : (
                <>
                  <button type="button" onClick={handleCancel} disabled={isSaving} className="btn btn-ghost">
                    Cancel
                  </button>
                  <button type="submit" disabled={isSaving} className="btn btn-primary">
                    {isSaving ? (
                      <>
                        <span className="loading loading-spinner loading-sm"></span>
                        Saving...
                      </>
                    ) : (
                      <>
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          className="h-5 w-5 mr-2"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        Save Changes
                      </>
                    )}
                  </button>
                </>
              )}
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
