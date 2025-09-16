import { services } from "@/lib/services";
import Image from "next/image";
import Link from "next/link";
import { WorkOSUser } from "@/lib/validation-schemas";

export default async function Navigation() {
  let user = null;
  let canManageOrganization = false;

  try {
    const session = await services.workos.withAuth();
    user = session?.user;
    // Check if user has organization management permission
    canManageOrganization = session?.permissions?.includes("organization:manage") || false;
  } catch {
    // User is not authenticated - that's ok, we'll show Sign In button
  }

  // Extract profile picture from metadata if it exists
  const workosUser = user as WorkOSUser;
  const profilePictureUrl = workosUser?.metadata?.profile_picture_url as string | undefined;

  return (
    <nav className="navbar bg-base-100 shadow-sm">
      <div className="flex-1">
        <Link href="/" className="btn btn-ghost px-2">
          <Image src="/kasho-wordmark-light.png" alt="Kasho" width={100} height={40} className="dark:invert" />
        </Link>
      </div>
      <div className="flex-none">
        {user ? (
          <div className="dropdown dropdown-end">
            <div tabIndex={0} className={`avatar ${!profilePictureUrl ? "avatar-placeholder" : ""} cursor-pointer`}>
              <div className="bg-primary text-primary-content w-10 rounded-full">
                {profilePictureUrl ? (
                  <Image src={profilePictureUrl} alt="Profile" width={40} height={40} className="rounded-full" />
                ) : (
                  <span>{user.email?.charAt(0).toUpperCase() || "U"}</span>
                )}
              </div>
            </div>
            <ul
              tabIndex={0}
              className="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow-lg bg-base-200 rounded-box w-52 border border-base-300"
            >
              <li className="menu-title">
                <span>{user.email}</span>
              </li>
              <li>
                <Link href="/account/profile">Profile</Link>
              </li>
              {canManageOrganization && (
                <li>
                  <Link href="/account/organization">Organization</Link>
                </li>
              )}
              <li className="divider"></li>
              <li>
                <a href="/logout">Sign Out</a>
              </li>
            </ul>
          </div>
        ) : (
          <a href="/login" className="btn btn-primary btn-sm">
            Sign In
          </a>
        )}
      </div>
    </nav>
  );
}
