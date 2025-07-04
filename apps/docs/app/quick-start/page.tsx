"use client";

import QuickStartContent from "@/content/quick-start.mdx";
import { MDXWrapper } from "@/app/components/mdx-wrapper";

export default function QuickStartPage() {
  return (
    <MDXWrapper>
      <QuickStartContent />
    </MDXWrapper>
  );
}
