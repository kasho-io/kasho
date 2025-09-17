"use client";

import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { WorkOSUser } from "@/lib/validation-schemas";

interface NavigationClientProps {
  user: WorkOSUser | null;
  canManageOrganization: boolean;
}

export default function NavigationClient({ user, canManageOrganization }: NavigationClientProps) {
  const router = useRouter();

  // Extract profile picture from metadata if it exists
  const profilePictureUrl = user?.metadata?.profile_picture_url as string | undefined;

  // Close dropdown when menu item is clicked
  const handleDropdownItemClick = () => {
    // Blur the currently focused element to close the dropdown
    if (document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
  };

  // Handle sign in navigation
  const handleSignIn = (e: React.MouseEvent) => {
    e.preventDefault();
    router.push("/login");
  };

  // Handle sign out navigation
  const handleSignOut = (e: React.MouseEvent) => {
    e.preventDefault();
    handleDropdownItemClick();
    router.push("/logout");
  };

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
                <Link href="/account/profile" onClick={handleDropdownItemClick}>
                  Profile
                </Link>
              </li>
              {canManageOrganization && (
                <li>
                  <Link href="/account/organization" onClick={handleDropdownItemClick}>
                    Organization
                  </Link>
                </li>
              )}
              <li className="divider"></li>
              <li>
                <a href="/logout" onClick={handleSignOut}>
                  Sign Out
                </a>
              </li>
            </ul>
          </div>
        ) : (
          <button onClick={handleSignIn} className="btn btn-primary btn-sm">
            Sign In
          </button>
        )}
      </div>
    </nav>
  );
}
