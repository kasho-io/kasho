"use client";

import { useFirstTimeOrgSetup } from "@/hooks/useFirstTimeOrgSetup";

interface ClientLayoutProps {
  children: React.ReactNode;
  hasOrganization: boolean;
}

export default function ClientLayout({ children, hasOrganization }: ClientLayoutProps) {
  useFirstTimeOrgSetup(hasOrganization);
  return <>{children}</>;
}
