"use client";

import HomeContent from "@/content/home.mdx";
import { MDXWrapper } from "@/app/components/mdx-wrapper";

export default function HomePage() {
  return (
    <MDXWrapper>
      <HomeContent />
    </MDXWrapper>
  );
}
