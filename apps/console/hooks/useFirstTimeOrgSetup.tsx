"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

export function useFirstTimeOrgSetup(hasOrganization: boolean) {
  const router = useRouter();

  useEffect(() => {
    // Skip if user already has an organization
    if (hasOrganization) {
      return;
    }

    async function checkAndSetupOrg() {
      try {
        const response = await fetch("/api/organization/setup", {
          method: "POST",
        });

        const data = await response.json();

        // If organization was created, refresh the page to update permissions
        if (data.requiresRefresh) {
          router.refresh();
          // Force a hard refresh to ensure navigation updates
          window.location.reload();
        }
      } catch (error) {
        console.error("Error checking organization setup:", error);
      }
    }

    checkAndSetupOrg();
  }, [hasOrganization, router]);
}
