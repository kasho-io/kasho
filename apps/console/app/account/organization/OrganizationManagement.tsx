"use client";

import { WorkOSWidgetProvider } from "@/components/WorkOSWidgetProvider";
import { UsersManagement, OrganizationSwitcher } from "@workos-inc/widgets";
import { useState, useEffect } from "react";

interface OrganizationManagementProps {
  authToken: string;
  organizationId?: string;
  currentOrganizationName?: string;
  userPermissions?: string[];
}

export function OrganizationManagement({
  authToken,
  organizationId,
  currentOrganizationName,
  userPermissions = [],
}: OrganizationManagementProps) {
  const [activeTab, setActiveTab] = useState<"members" | "settings">("members");
  const [organizationName, setOrganizationName] = useState(currentOrganizationName || "");
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateMessage, setUpdateMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  // Check if user has permission to manage organization
  const canManageOrganization = userPermissions.includes("organization:manage");

  return (
    <WorkOSWidgetProvider>
      <div className="space-y-6">
        {/* Organization Switcher */}
        <div className="flex justify-between items-center mb-6">
          <div>
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
        <div className="tabs tabs-border mb-6">
          <button
            className={`tab ${activeTab === "members" ? "tab-active" : ""}`}
            onClick={() => setActiveTab("members")}
          >
            Team Members
          </button>
          {canManageOrganization && (
            <button
              className={`tab ${activeTab === "settings" ? "tab-active" : ""}`}
              onClick={() => setActiveTab("settings")}
            >
              Settings
            </button>
          )}
        </div>

        {/* Tab Content */}
        <div className="min-h-[400px]">
          {activeTab === "members" && (
            <div>
              <UsersManagement authToken={authToken} />
            </div>
          )}

          {activeTab === "settings" && canManageOrganization && (
            <div>
              <div className="space-y-6">
                {/* Organization Name */}
                <div className="card bg-base-200">
                  <div className="card-body">
                    <form
                      onSubmit={async (e) => {
                        e.preventDefault();
                        setIsUpdating(true);
                        setUpdateMessage(null);

                        try {
                          const response = await fetch("/api/organization/update", {
                            method: "PUT",
                            headers: { "Content-Type": "application/json" },
                            body: JSON.stringify({ name: organizationName }),
                          });

                          if (response.ok) {
                            setUpdateMessage({
                              type: "success",
                              text: "Organization name updated successfully!",
                            });
                            // Optionally reload to update the organization switcher
                            setTimeout(() => window.location.reload(), 1500);
                          } else {
                            const error = await response.json();
                            setUpdateMessage({
                              type: "error",
                              text: error.error || "Failed to update organization name",
                            });
                          }
                        } catch (error) {
                          setUpdateMessage({
                            type: "error",
                            text: "An error occurred while updating the organization name",
                          });
                        } finally {
                          setIsUpdating(false);
                        }
                      }}
                    >
                      <div className="form-control w-full max-w-md">
                        <label className="label">
                          <span className="label-text">Organization Name</span>
                        </label>
                        <input
                          type="text"
                          value={organizationName}
                          onChange={(e) => setOrganizationName(e.target.value)}
                          className="input input-bordered w-full"
                          placeholder="Enter organization name"
                          required
                          disabled={isUpdating}
                        />
                      </div>

                      {updateMessage && (
                        <div
                          className={`alert ${updateMessage.type === "success" ? "alert-success" : "alert-error"} mt-4`}
                        >
                          <span>{updateMessage.text}</span>
                        </div>
                      )}

                      <div className="mt-4 flex justify-end">
                        <button
                          type="submit"
                          className="btn btn-primary"
                          disabled={
                            isUpdating || !organizationName.trim() || organizationName === currentOrganizationName
                          }
                        >
                          {isUpdating ? "Saving..." : "Save Changes"}
                        </button>
                      </div>
                    </form>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </WorkOSWidgetProvider>
  );
}
