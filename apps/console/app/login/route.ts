import { getSignInUrl } from "@workos-inc/authkit-nextjs";
import { redirect } from "next/navigation";

export async function GET() {
  // Get the sign-in URL from WorkOS
  const signInUrl = await getSignInUrl();

  // Redirect the user to WorkOS for authentication
  redirect(signInUrl);
}
