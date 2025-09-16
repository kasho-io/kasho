"use client";

import { WorkOsWidgets } from "@workos-inc/widgets";
import { useState } from "react";

interface WorkOSWidgetProviderProps {
  children: React.ReactNode;
}

export function WorkOSWidgetProvider({ children }: WorkOSWidgetProviderProps) {
  const [appearance, setAppearance] = useState<"dark" | "light">("light");

  return (
    <WorkOsWidgets
      theme={{
        appearance,
        accentColor: "violet", // Closer to DaisyUI's primary color
        radius: "large", // DaisyUI uses more rounded corners
        fontFamily: "inherit", // Use the app's font (Inter)
      }}
    >
      {children}
    </WorkOsWidgets>
  );
}
