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
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-4 w-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
          />
        </svg>
        <span className="ml-2">{currentOrg?.name || "Select Organization"}</span>
        <svg className="ml-1 h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
          <path
            fillRule="evenodd"
            d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
            clipRule="evenodd"
          />
        </svg>
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
