import { withAuth, getSignInUrl, getSignUpUrl, signOut } from "@workos-inc/authkit-nextjs";
import { redirect } from "next/navigation";

export default async function AuthTestPage() {
  const { user } = await withAuth();

  const signInUrl = await getSignInUrl();

  async function handleSignOut() {
    "use server";
    await signOut();
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-base-200">
      <div className="card w-96 bg-base-100 shadow-xl">
        <div className="card-body">
          <h2 className="card-title">Authentication Test</h2>
          {user ? (
            <>
              <p>Signed in as: {user.email}</p>
              <form action={handleSignOut}>
                <button type="submit" className="btn btn-primary">
                  Sign Out
                </button>
              </form>
            </>
          ) : (
            <>
              <p>Not signed in</p>
              <a href={signInUrl} className="btn btn-primary">
                Sign In
              </a>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
