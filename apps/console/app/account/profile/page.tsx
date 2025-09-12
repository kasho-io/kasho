import { withAuth } from "@/lib/auth-wrapper";
import { redirect } from "next/navigation";
import ProfileForm from "./ProfileForm";

export default async function ProfilePage() {
  const { user } = await withAuth();

  if (!user) {
    redirect("/login");
  }

  // Extract metadata fields if they exist
  const userWithMetadata = user as {
    metadata?: Record<string, string>;
    emailVerified?: boolean;
  };
  const metadata = userWithMetadata.metadata || {};
  const profileData = {
    id: user.id,
    email: user.email || "",
    emailVerified: userWithMetadata.emailVerified ?? true,
    firstName: metadata.first_name || "",
    lastName: metadata.last_name || "",
    profilePictureUrl: metadata.profile_picture_url || "",
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
