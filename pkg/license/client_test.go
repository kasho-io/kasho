package license_test

import (
	"context"
	"testing"
	"time"

	"kasho/pkg/license/testutil"
)

func TestClient_ValidateLicense(t *testing.T) {
	client, cleanup := testutil.NewTestClient(t)
	defer cleanup()

	ctx := context.Background()
	err := client.ValidateLicense(ctx)
	if err != nil {
		t.Fatalf("Expected license validation to succeed, got error: %v", err)
	}
}

func TestClient_GetLicenseInfo(t *testing.T) {
	client, cleanup := testutil.NewTestClient(t)
	defer cleanup()

	ctx := context.Background()
	info, err := client.GetLicenseInfo(ctx)
	if err != nil {
		t.Fatalf("Expected to get license info, got error: %v", err)
	}

	if info.CustomerName != "Test Customer" {
		t.Errorf("Expected customer name 'Test Customer', got '%s'", info.CustomerName)
	}
}

func TestClient_StartPeriodicValidation(t *testing.T) {
	client, cleanup := testutil.NewTestClient(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start periodic validation with a short interval
	failChan := client.StartPeriodicValidation(ctx, 100*time.Millisecond)

	// Let it run for a bit
	select {
	case <-failChan:
		t.Fatal("Periodic validation failed unexpectedly")
	case <-time.After(300 * time.Millisecond):
		// Success - validation ran without failing
	}

	// Cancel context and ensure goroutine exits
	cancel()
	select {
	case <-failChan:
		// Good - channel closed when context cancelled
	case <-time.After(time.Second):
		t.Fatal("Periodic validation goroutine didn't exit after context cancel")
	}
}

func TestClient_MustValidate(t *testing.T) {
	// This test is tricky because MustValidate calls log.Fatal on failure
	// For now, we just test that it doesn't panic with a valid license
	client, cleanup := testutil.NewTestClient(t)
	defer cleanup()

	ctx := context.Background()
	// This should not panic
	client.MustValidate(ctx)
}

func TestClient_Caching(t *testing.T) {
	client, cleanup := testutil.NewTestClient(t)
	defer cleanup()

	ctx := context.Background()
	
	// First call should hit the server
	info1, err := client.GetLicenseInfo(ctx)
	if err != nil {
		t.Fatalf("First GetLicenseInfo failed: %v", err)
	}

	// Second call should use cache (we can't easily verify this without
	// instrumenting the mock server, but at least verify it works)
	info2, err := client.GetLicenseInfo(ctx)
	if err != nil {
		t.Fatalf("Second GetLicenseInfo failed: %v", err)
	}

	if info1.CustomerName != info2.CustomerName {
		t.Errorf("Cached response differs from original")
	}
}