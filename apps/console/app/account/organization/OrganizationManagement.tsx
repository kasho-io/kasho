"use client";

import { WorkOSWidgetProvider } from "@/components/WorkOSWidgetProvider";
import { UsersManagement, OrganizationSwitcher, UserProfile } from "@workos-inc/widgets";
import { useState } from "react";

interface OrganizationManagementProps {
  authToken: string;
  organizationId?: string;
}

export function OrganizationManagement({ authToken }: OrganizationManagementProps) {
  const [activeTab, setActiveTab] = useState<"members" | "profile" | "settings">("members");

  return (
    <WorkOSWidgetProvider>
      <div className="space-y-6">
        {/* Organization Switcher */}
        <div className="flex justify-between items-center mb-6">
          <div>
            <label className="text-sm font-medium text-base-content/70">Current Organization</label>
            <OrganizationSwitcher
              authToken={authToken}
              switchToOrganization={() => {
                // Handle organization switch - reload to get new org context
                window.location.reload();
              }}
            />
          </div>
        </div>

        {/* Tabs */}
        <div className="tabs tabs-boxed mb-6">
          <button
            className={`tab ${activeTab === "members" ? "tab-active" : ""}`}
            onClick={() => setActiveTab("members")}
          >
            Team Members
          </button>
          <button
            className={`tab ${activeTab === "profile" ? "tab-active" : ""}`}
            onClick={() => setActiveTab("profile")}
          >
            My Profile
          </button>
          <button
            className={`tab ${activeTab === "settings" ? "tab-active" : ""}`}
            onClick={() => setActiveTab("settings")}
          >
            Settings
          </button>
        </div>

        {/* Tab Content */}
        <div className="min-h-[400px]">
          {activeTab === "members" && (
            <div>
              <h3 className="text-lg font-semibold mb-4">Team Members</h3>
              <p className="text-base-content/70 mb-6">
                Invite new members, manage roles, and control access to your organization.
              </p>
              <UsersManagement authToken={authToken} />
            </div>
          )}

          {activeTab === "profile" && (
            <div>
              <h3 className="text-lg font-semibold mb-4">My Profile</h3>
              <p className="text-base-content/70 mb-6">View and update your personal information.</p>
              <UserProfile authToken={authToken} />
            </div>
          )}

          {activeTab === "settings" && (
            <div>
              <h3 className="text-lg font-semibold mb-4">Organization Settings</h3>
              <p className="text-base-content/70 mb-6">Configure organization-wide settings and preferences.</p>

              <div className="space-y-6">
                <div className="card bg-base-200">
                  <div className="card-body">
                    <h4 className="font-semibold">Organization Admin Portal</h4>
                    <p className="text-sm text-base-content/70 mb-4">
                      Access the Admin Portal to manage your organization&apos;s name, configure SSO, verify domains,
                      and more.
                    </p>
                    <button
                      className="btn btn-primary"
                      onClick={async () => {
                        try {
                          const response = await fetch("/api/admin-portal", {
                            method: "POST",
                          });
                          const { link } = await response.json();
                          if (link) {
                            window.open(link, "_blank");
                          }
                        } catch (error) {
                          console.error("Failed to open admin portal:", error);
                        }
                      }}
                    >
                      Open Admin Portal
                    </button>
                  </div>
                </div>

                <div className="alert alert-info">
                  <span>
                    ℹ️ The Admin Portal allows you to rename your organization, configure SSO, manage domains, and more.
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </WorkOSWidgetProvider>
  );
}
