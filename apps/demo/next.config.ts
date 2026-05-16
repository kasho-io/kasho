import path from "node:path";
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Produce a self-contained server build that can run in a minimal container
  // image without copying the full node_modules tree.
  output: "standalone",
  // Anchor file tracing at the monorepo root so workspace hoisted modules are
  // included in the standalone bundle.
  outputFileTracingRoot: path.join(__dirname, "../.."),
};

export default nextConfig;
