import { signOut } from "@workos-inc/authkit-nextjs";

export async function GET() {
  // Sign out the user and redirect to home
  await signOut();
}