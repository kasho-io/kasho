import type { Services } from "./services/types";
import { TestWorkOSService, TestVercelBlobService } from "./services/services.test";
import { ProductionWorkOSService, ProductionVercelBlobService } from "./services/services.prod";

// Singleton instances
let testServices: Services | null = null;
let prodServices: Services | null = null;

// Get services based on environment - checked at runtime
const getServices = (): Services => {
  const isTestEnvironment = process.env.NODE_ENV === "test" || process.env.MOCK_AUTH === "true";

  if (isTestEnvironment) {
    if (!testServices) {
      testServices = {
        workos: new TestWorkOSService(),
        vercelBlob: new TestVercelBlobService(),
      };
    }
    return testServices;
  }

  if (!prodServices) {
    prodServices = {
      workos: new ProductionWorkOSService(),
      vercelBlob: new ProductionVercelBlobService(),
    };
  }
  return prodServices;
};

// Export services with getters that check environment at runtime
export const services: Services = {
  get workos() {
    return getServices().workos;
  },
  get vercelBlob() {
    return getServices().vercelBlob;
  },
};

// Export convenience methods for backward compatibility
export const withAuth = () => services.workos.withAuth();
export const workosClient = services.workos; // For gradual migration
