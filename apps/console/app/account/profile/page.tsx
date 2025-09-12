import { withAuth } from "@/lib/auth-wrapper";
import { redirect } from "next/navigation";
import ProfileForm from "./ProfileForm";
import { workosClient } from "@/lib/workos-client";

export default async function ProfilePage() {
  const { user } = await withAuth();

  if (!user) {
    redirect("/login");
  }

  // Fetch fresh user data directly from WorkOS API to avoid cached session issues
  let freshUserData = user;

  // Only fetch fresh data in production mode with real WorkOS auth
  if (process.env.NODE_ENV !== "test" && process.env.MOCK_AUTH !== "true") {
    try {
      const freshUser = await workosClient.userManagement.getUser(user.id);
      freshUserData = freshUser;
    } catch (error) {
      console.error("Failed to fetch fresh user data:", error);
      // Fall back to cached session data
    }
  }

  // Extract user data - WorkOS returns firstName/lastName as direct properties
  // but we store them in metadata for updates
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const userWithFields = freshUserData as any; // Type assertion since WorkOS types may vary
  const metadata = userWithFields.metadata || {};

  const profileData = {
    id: userWithFields.id,
    email: userWithFields.email || "",
    emailVerified: userWithFields.emailVerified ?? true,
    // Prioritize metadata over direct properties since WorkOS stores updates in metadata
    // Direct firstName/lastName are immutable and set only during user creation
    firstName: metadata.first_name || userWithFields.firstName || "",
    lastName: metadata.last_name || userWithFields.lastName || "",
    profilePictureUrl: metadata.profile_picture_url || userWithFields.profilePictureUrl || "",
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
