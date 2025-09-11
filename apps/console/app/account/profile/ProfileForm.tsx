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
    <div className="card bg-base-100 shadow-xl">
      <div className="card-body">
        {message && (
          <div className={`alert ${message.type === "success" ? "alert-success" : "alert-error"} mb-4`}>
            <span>{message.text}</span>
          </div>
        )}

        <form onSubmit={handleSubmit}>
          {/* Profile Picture */}
          <div className="form-control mb-6">
            <label className="label">
              <span className="label-text">Profile Picture</span>
            </label>
            <div className="flex items-center gap-4">
              <div className="avatar">
                <div className="w-24 rounded-full bg-base-200">
                  {previewUrl ? (
                    <Image src={previewUrl} alt="Profile" width={96} height={96} className="rounded-full" />
                  ) : (
                    <div className="flex items-center justify-center h-full text-3xl text-base-content">
                      {formData.firstName?.[0]?.toUpperCase() || formData.email?.[0]?.toUpperCase() || "U"}
                    </div>
                  )}
                </div>
              </div>
              {isEditing && (
                <div>
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
                    className="btn btn-sm btn-outline"
                  >
                    Change Picture
                  </button>
                </div>
              )}
            </div>
          </div>

          {/* Email (read-only) */}
          <div className="form-control mb-4">
            <label className="label">
              <span className="label-text">Email</span>
            </label>
            <input type="email" value={formData.email} disabled className="input input-bordered bg-base-200" />
            <label className="label">
              <span className="label-text-alt text-base-content/60">Email cannot be changed here</span>
            </label>
          </div>

          {/* First Name */}
          <div className="form-control mb-4">
            <label className="label">
              <span className="label-text">First Name</span>
            </label>
            <input
              type="text"
              name="firstName"
              value={formData.firstName}
              onChange={handleInputChange}
              disabled={!isEditing}
              className="input input-bordered"
              placeholder="Enter your first name"
            />
          </div>

          {/* Last Name */}
          <div className="form-control mb-4">
            <label className="label">
              <span className="label-text">Last Name</span>
            </label>
            <input
              type="text"
              name="lastName"
              value={formData.lastName}
              onChange={handleInputChange}
              disabled={!isEditing}
              className="input input-bordered"
              placeholder="Enter your last name"
            />
          </div>

          {/* Location */}
          <div className="form-control mb-4">
            <label className="label">
              <span className="label-text">Location</span>
            </label>
            <input
              type="text"
              name="location"
              value={formData.location}
              onChange={handleInputChange}
              disabled={!isEditing}
              className="input input-bordered"
              placeholder="e.g., San Francisco, CA"
            />
          </div>

          {/* Bio */}
          <div className="form-control mb-6">
            <label className="label">
              <span className="label-text">Bio</span>
            </label>
            <textarea
              name="bio"
              value={formData.bio}
              onChange={handleInputChange}
              disabled={!isEditing}
              className="textarea textarea-bordered h-24"
              placeholder="Tell us about yourself"
            />
          </div>

          {/* Action Buttons */}
          <div className="flex gap-2">
            {!isEditing ? (
              <button type="button" onClick={() => setIsEditing(true)} className="btn btn-primary">
                Edit Profile
              </button>
            ) : (
              <>
                <button type="submit" disabled={isSaving} className="btn btn-primary">
                  {isSaving ? <span className="loading loading-spinner loading-sm"></span> : "Save Changes"}
                </button>
                <button type="button" onClick={handleCancel} disabled={isSaving} className="btn btn-ghost">
                  Cancel
                </button>
              </>
            )}
          </div>
        </form>
      </div>
    </div>
  );
}
