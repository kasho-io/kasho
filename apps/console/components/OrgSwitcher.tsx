"use client";

import { useState } from "react";

interface Organization {
  id: string;
  name: string;
}

interface OrgSwitcherProps {
  organizations: Organization[];
  currentOrganizationId?: string;
}

export function OrgSwitcher({ organizations, currentOrganizationId }: OrgSwitcherProps) {
  const [isSwitching, setIsSwitching] = useState(false);

  // Find the current organization
  const currentOrg = organizations.find((org) => org.id === currentOrganizationId);

  // Don't render if there's only one organization or none
  if (organizations.length <= 1) {
    return null;
  }

  const handleSwitchOrganization = async (organizationId: string) => {
    if (organizationId === currentOrganizationId || isSwitching) return;

    setIsSwitching(true);
    try {
      const response = await fetch("/api/organization/switch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ organizationId }),
      });

      if (response.ok) {
        // Reload to refresh the context with new organization
        window.location.reload();
      } else {
        const error = await response.json();
        console.error("Failed to switch organization:", error);
        setIsSwitching(false);
      }
    } catch (error) {
      console.error("Error switching organization:", error);
      setIsSwitching(false);
    }
  };

  return (
    <div className="dropdown dropdown-end">
      <div tabIndex={0} role="button" className="btn btn-ghost btn-sm">
        <span>🏢</span>
        <span className="ml-2">{currentOrg?.name || "Select Organization"}</span>
        <span>⏷</span>
      </div>
      <ul
        tabIndex={0}
        className="dropdown-content menu p-2 shadow-lg bg-base-100 rounded-box w-52 mt-2 border border-base-300"
      >
        <li className="menu-title">
          <span>Switch Organization</span>
        </li>
        {organizations.map((org) => (
          <li key={org.id}>
            <a
              onClick={() => handleSwitchOrganization(org.id)}
              className={`${org.id === currentOrganizationId ? "active" : ""} ${isSwitching ? "disabled" : ""}`}
            >
              {org.name}
              {org.id === currentOrganizationId && (
                <span className="badge badge-primary badge-xs ml-auto">Current</span>
              )}
            </a>
          </li>
        ))}
      </ul>
    </div>
  );
}
