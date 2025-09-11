import { WorkOS } from "@workos-inc/node";

if (!process.env.WORKOS_API_KEY) {
  throw new Error("WORKOS_API_KEY environment variable is required");
}

export const workosClient = new WorkOS(process.env.WORKOS_API_KEY);
