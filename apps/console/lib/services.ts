import type { Services } from "./services/types";
import { TestWorkOSService, TestVercelBlobService } from "./services/services.test";
import { ProductionWorkOSService, ProductionVercelBlobService } from "./services/services.prod";

// Create services based on environment
// This is the ONLY place in the application that checks NODE_ENV
const createServices = (): Services => {
  const isTestEnvironment = process.env.NODE_ENV === "test";

  if (isTestEnvironment) {
    return {
      workos: new TestWorkOSService(),
      vercelBlob: new TestVercelBlobService(),
    };
  }

  return {
    workos: new ProductionWorkOSService(),
    vercelBlob: new ProductionVercelBlobService(),
  };
};

// Export a single instance of services
export const services = createServices();

// Export convenience methods for backward compatibility
export const withAuth = () => services.workos.withAuth();
export const workosClient = services.workos; // For gradual migration
