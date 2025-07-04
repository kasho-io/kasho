"use client";

import DatabaseSetupContent from "@/content/installation/database-setup.mdx";
import { MDXWrapper } from "@/app/components/mdx-wrapper";

export default function DatabaseSetupPage() {
  return (
    <MDXWrapper>
      <DatabaseSetupContent />
    </MDXWrapper>
  );
}
