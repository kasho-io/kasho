import { withAuth } from "@workos-inc/authkit-nextjs";
import { redirect } from "next/navigation";
import ProfileForm from "./ProfileForm";

export default async function ProfilePage() {
  const { user } = await withAuth();

  if (!user) {
    redirect("/login");
  }

  // Extract metadata fields if they exist
  const metadata = (user as { metadata?: Record<string, string> }).metadata || {};
  const profileData = {
    id: user.id,
    email: user.email || "",
    firstName: metadata.first_name || "",
    lastName: metadata.last_name || "",
    profilePictureUrl: metadata.profile_picture_url || "",
    bio: metadata.bio || "",
    location: metadata.location || "",
  };

  return (
    <div className="min-h-screen bg-base-200">
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-3xl mx-auto">
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-base-content mb-2">Account Settings</h1>
            <p className="text-base-content/70">Manage your personal information and preferences</p>
          </div>
          <ProfileForm initialData={profileData} />
        </div>
      </div>
    </div>
  );
}
