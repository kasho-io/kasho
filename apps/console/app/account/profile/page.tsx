import { withAuth } from "@/lib/auth-wrapper";
import { redirect } from "next/navigation";
import ProfileForm from "./ProfileForm";

export default async function ProfilePage() {
  const { user } = await withAuth();

  if (!user) {
    redirect("/login");
  }

  // Extract user data - WorkOS returns firstName/lastName as direct properties
  // but we store them in metadata for updates
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const userWithFields = user as any; // Type assertion since WorkOS types may vary
  const metadata = userWithFields.metadata || {};

  const profileData = {
    id: userWithFields.id,
    email: userWithFields.email || "",
    emailVerified: userWithFields.emailVerified ?? true,
    // Check both direct properties and metadata
    firstName: userWithFields.firstName || metadata.first_name || "",
    lastName: userWithFields.lastName || metadata.last_name || "",
    profilePictureUrl: userWithFields.profilePictureUrl || metadata.profile_picture_url || "",
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
