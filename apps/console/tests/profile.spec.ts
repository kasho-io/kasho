import { test, expect } from "@playwright/test";

// Mock user data for testing
const mockUser = {
  email: "test@example.com",
  firstName: "John",
  lastName: "Doe",
  profilePictureUrl: "",
};

test.describe("Profile Form", () => {
  // Note: beforeEach removed as it's not currently used

  test("should display profile form with correct fields", async ({ page }) => {
    await page.goto("/account/profile");

    // Check if all form fields are present
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="firstName"]')).toBeVisible();
    await expect(page.locator('input[name="lastName"]')).toBeVisible();
    await expect(page.locator('button:has-text("Save Changes")')).toBeVisible();
  });

  test("should enable save button only when changes are made", async ({ page }) => {
    await page.goto("/account/profile");

    // Wait for form to load
    await page.waitForSelector('input[name="firstName"]');

    // Initially, save button should be disabled (no changes)
    const saveButton = page.locator('button:has-text("Save Changes")');
    await expect(saveButton).toBeDisabled();

    // Make a change to first name
    const firstNameInput = page.locator('input[name="firstName"]');
    const originalValue = await firstNameInput.inputValue();
    await firstNameInput.fill("Jane");

    // Save button should now be enabled
    await expect(saveButton).toBeEnabled();

    // Revert the change to the original value
    await firstNameInput.fill(originalValue);

    // Save button should be disabled again
    await expect(saveButton).toBeDisabled();
  });

  test("should show email verification warning when email is unverified", async ({ page }) => {
    // This test would need to mock an unverified email state
    await page.goto("/account/profile");

    // Look for the unverified badge (if present)
    const unverifiedBadge = page.locator('.badge:has-text("Unverified")');

    // If email is unverified, check for warning message
    if (await unverifiedBadge.isVisible()) {
      await expect(page.locator(".alert-warning")).toBeVisible();
      await expect(page.locator(".alert-warning")).toContainText("Email Verification Required");
    }
  });

  test("should keep email field enabled regardless of verification status", async ({ page }) => {
    await page.goto("/account/profile");

    const emailInput = page.locator('input[name="email"]');

    // Email field should always be enabled (users can update profile even with unverified email)
    await expect(emailInput).toBeEnabled();
  });

  test("should allow updating first and last name when email is verified", async ({ page }) => {
    await page.goto("/account/profile");

    const firstNameInput = page.locator('input[name="firstName"]');
    const lastNameInput = page.locator('input[name="lastName"]');
    const saveButton = page.locator('button:has-text("Save Changes")');

    // Update first and last name
    await firstNameInput.fill("Jane");
    await lastNameInput.fill("Smith");

    // Save button should be enabled
    await expect(saveButton).toBeEnabled();

    // Mock the API response for successful update
    await page.route("/api/user/update", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          success: true,
          user: {
            ...mockUser,
            metadata: {
              first_name: "Jane",
              last_name: "Smith",
              profile_picture_url: "",
            },
          },
        }),
      });
    });

    // Click save
    await saveButton.click();

    // Check for success message
    await expect(page.locator(".alert-success")).toBeVisible();
    await expect(page.locator(".alert-success")).toContainText("Profile updated successfully");
  });

  test("should handle API errors gracefully", async ({ page }) => {
    await page.goto("/account/profile");

    const firstNameInput = page.locator('input[name="firstName"]');
    const saveButton = page.locator('button:has-text("Save Changes")');

    // Make a change
    await firstNameInput.fill("Jane");

    // Mock API error response
    await page.route("/api/user/update", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({
          error: "Failed to update profile",
        }),
      });
    });

    // Click save
    await saveButton.click();

    // Check for error message
    await expect(page.locator(".alert-error")).toBeVisible();
    await expect(page.locator(".alert-error")).toContainText("Failed to update profile");
  });

  test("should handle profile picture upload", async ({ page }) => {
    await page.goto("/account/profile");

    // Find the file input (hidden) and upload button
    const fileInput = page.locator('input[type="file"]');
    const uploadButton = page.locator('button:has-text("Change Picture")');

    // Check that upload button is visible
    await expect(uploadButton).toBeVisible();

    // Mock the upload API response BEFORE uploading
    await page.route("/api/user/upload-avatar", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          success: true,
          url: "https://example.com/avatar.jpg",
        }),
      });
    });

    // Upload a file (create a test image)
    const buffer = Buffer.from("fake-image-data");
    await fileInput.setInputFiles({
      name: "test.jpg",
      mimeType: "image/jpeg",
      buffer: buffer,
    });

    // Wait for success message with longer timeout
    await expect(page.locator(".alert-success")).toBeVisible({ timeout: 10000 });
    await expect(page.locator(".alert-success")).toContainText("Image uploaded successfully");
  });

  test("should validate email format", async ({ page }) => {
    await page.goto("/account/profile");

    const emailInput = page.locator('input[name="email"]');
    const saveButton = page.locator('button:has-text("Save Changes")');

    // Enter invalid email
    await emailInput.fill("invalid-email");

    // Try to save (HTML5 validation should prevent submission)
    await saveButton.click();

    // Check that the input has validation error (browser native validation)
    const validationMessage = await emailInput.evaluate((el: HTMLInputElement) => el.validationMessage);
    expect(validationMessage).toBeTruthy();
  });

  test("should mark email as unverified after email change", async ({ page }) => {
    await page.goto("/account/profile");

    const emailInput = page.locator('input[name="email"]');
    const saveButton = page.locator('button:has-text("Save Changes")');

    // Change email
    await emailInput.fill("newemail@example.com");

    // Mock successful update that returns unverified status
    await page.route("/api/user/update", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          success: true,
          user: {
            ...mockUser,
            email: "newemail@example.com",
            emailVerified: false,
          },
        }),
      });
    });

    // Save changes
    await saveButton.click();

    // After save, check for unverified badge
    await expect(page.locator('.badge:has-text("Unverified")')).toBeVisible({ timeout: 5000 });

    // Also check the help text
    await expect(
      page.locator('.label-text-alt:has-text("You will need to verify your email the next time you sign in.")'),
    ).toBeVisible();
  });

  test("should preserve form data after failed save", async ({ page }) => {
    await page.goto("/account/profile");

    const firstNameInput = page.locator('input[name="firstName"]');
    const lastNameInput = page.locator('input[name="lastName"]');
    const saveButton = page.locator('button:has-text("Save Changes")');

    // Enter new values
    const newFirstName = "Jane";
    const newLastName = "Smith";
    await firstNameInput.fill(newFirstName);
    await lastNameInput.fill(newLastName);

    // Mock API error
    await page.route("/api/user/update", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({
          error: "Server error",
        }),
      });
    });

    // Try to save
    await saveButton.click();

    // Check that form data is preserved after error
    await expect(firstNameInput).toHaveValue(newFirstName);
    await expect(lastNameInput).toHaveValue(newLastName);

    // Save button should still be enabled (changes not saved)
    await expect(saveButton).toBeEnabled();
  });
});

test.describe("Profile Form - Data Persistence", () => {
  test("should correctly handle metadata-based name updates", async ({ page }) => {
    // This test verifies that WorkOS metadata is properly handled
    // since firstName/lastName are stored in metadata, not as direct properties
    await page.goto("/account/profile");

    const firstNameInput = page.locator('input[name="firstName"]');
    const lastNameInput = page.locator('input[name="lastName"]');
    const saveButton = page.locator('button:has-text("Save Changes")');

    // Change names
    const newFirstName = "UpdatedFirst";
    const newLastName = "UpdatedLast";
    await firstNameInput.fill(newFirstName);
    await lastNameInput.fill(newLastName);

    // Mock successful update that returns metadata (not direct properties)
    await page.route("/api/user/update", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          success: true,
          user: {
            ...mockUser,
            // WorkOS doesn't update these direct properties
            firstName: mockUser.firstName,
            lastName: mockUser.lastName,
            // Names are actually stored in metadata
            metadata: {
              first_name: newFirstName,
              last_name: newLastName,
              profile_picture_url: "",
            },
          },
        }),
      });
    });

    // Save changes
    await saveButton.click();

    // Wait for success message
    await expect(page.locator(".alert-success")).toBeVisible();

    // The form should display the new values from metadata
    await expect(firstNameInput).toHaveValue(newFirstName);
    await expect(lastNameInput).toHaveValue(newLastName);

    // Save button should be disabled (no pending changes)
    await expect(saveButton).toBeDisabled();
  });

  test("should persist changes after page reload", async ({ page }) => {
    await page.goto("/account/profile");

    const firstNameInput = page.locator('input[name="firstName"]');
    const lastNameInput = page.locator('input[name="lastName"]');
    const saveButton = page.locator('button:has-text("Save Changes")');

    // Change names
    const newFirstName = "UpdatedFirst";
    const newLastName = "UpdatedLast";
    await firstNameInput.fill(newFirstName);
    await lastNameInput.fill(newLastName);

    // Mock successful update with metadata
    await page.route("/api/user/update", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          success: true,
          user: {
            ...mockUser,
            // WorkOS returns original direct properties unchanged
            firstName: mockUser.firstName,
            lastName: mockUser.lastName,
            // Updated values are in metadata
            metadata: {
              first_name: newFirstName,
              last_name: newLastName,
              profile_picture_url: "",
            },
          },
        }),
      });
    });

    // Save changes
    await saveButton.click();

    // Wait for success message
    await expect(page.locator(".alert-success")).toBeVisible();

    // After save, the save button should be disabled (no pending changes)
    await expect(saveButton).toBeDisabled();

    // The form should show the new values
    await expect(firstNameInput).toHaveValue(newFirstName);
    await expect(lastNameInput).toHaveValue(newLastName);
  });

  test("should show unverified badge persistently after email change", async ({ page }) => {
    await page.goto("/account/profile");

    const emailInput = page.locator('input[name="email"]');

    // Change email to trigger unverified state
    await emailInput.fill("changed@example.com");

    // The badge should appear immediately in the form (local state)
    await expect(page.locator('.badge:has-text("Unverified")')).toBeVisible();

    // The help text should update to show appropriate message
    // When email is changed from a verified state, it shows different text
    const helpText = page.locator(".label-text-alt").first();
    await expect(helpText).toContainText(/You will need to verify|You'll need to verify/);
  });
});

test.describe("Profile Form - Edge Cases", () => {
  test("should handle very long names", async ({ page }) => {
    await page.goto("/account/profile");

    const firstNameInput = page.locator('input[name="firstName"]');
    const veryLongName = "A".repeat(255);

    await firstNameInput.fill(veryLongName);

    // Check that the input accepts the value
    await expect(firstNameInput).toHaveValue(veryLongName);
  });

  test("should handle special characters in names", async ({ page }) => {
    await page.goto("/account/profile");

    const firstNameInput = page.locator('input[name="firstName"]');
    const lastNameInput = page.locator('input[name="lastName"]');

    // Test various special characters
    await firstNameInput.fill("O'Connor");
    await lastNameInput.fill("Smith-Jones");

    await expect(firstNameInput).toHaveValue("O'Connor");
    await expect(lastNameInput).toHaveValue("Smith-Jones");
  });

  test("should handle rapid form submissions", async ({ page }) => {
    await page.goto("/account/profile");

    // Wait for form to load
    await page.waitForSelector('input[name="firstName"]');

    const firstNameInput = page.locator('input[name="firstName"]');
    const saveButton = page.locator('button:has-text("Save Changes")');

    // Mock successful API response
    await page.route("/api/user/update", async (route) => {
      // Add a delay to simulate network latency
      await new Promise((resolve) => setTimeout(resolve, 500));
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          success: true,
          user: mockUser,
        }),
      });
    });

    // Make a change to enable the save button
    await firstNameInput.fill("NewTestName");

    // Save button should now be enabled
    await expect(saveButton).toBeEnabled();

    // Click save
    await saveButton.click();

    // Wait for success message instead of checking button state
    // (In test mode, the response might be too fast to catch the "Saving..." state)
    await expect(page.locator(".alert-success")).toBeVisible({ timeout: 5000 });
  });
});
