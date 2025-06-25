"use client";

import { ReactNode } from "react";

export function MDXWrapper({ children }: { children: ReactNode }) {
  return (
    <div className="container mx-auto px-6 py-8 max-w-5xl">
      {children}
    </div>
  );
}