"use client";

import { WorkOsWidgets } from "@workos-inc/widgets";
import "@workos-inc/widgets/styles.css";

interface WorkOSWidgetProviderProps {
  children: React.ReactNode;
}

export function WorkOSWidgetProvider({ children }: WorkOSWidgetProviderProps) {
  return (
    <WorkOsWidgets
      theme={{
        appearance: "inherit",
        accentColor: "blue",
        radius: "medium",
        fontFamily: "inherit",
      }}
    >
      {children}
    </WorkOsWidgets>
  );
}
