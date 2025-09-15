import { defineConfig, devices } from "@playwright/test";

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: "./tests",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: "html",
  use: {
    baseURL: process.env.BASE_URL || "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },

  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],

  webServer: {
    command: "npm run dev",
    url: "http://localhost:3000",
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
    env: {
      ...process.env,
      NODE_ENV: "test",
      WORKOS_CLIENT_ID: "test_client_id",
      WORKOS_API_KEY: "sk_test_placeholder",
      WORKOS_CLIENT_SECRET: "test_secret",
      NEXT_PUBLIC_WORKOS_REDIRECT_URI: "http://localhost:3000/callback",
      WORKOS_COOKIE_PASSWORD: "test_cookie_password_at_least_32_characters_long",
      BLOB_READ_WRITE_TOKEN: "test_blob_token",
    },
  },
});
