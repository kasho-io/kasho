package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"kasho/pkg/license"
)

// testPrivateKey is a test-only private key
const testPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCFfxSkIdr0sctR
w/VUnu7bApMsQ1/M+rSNQkS49RbTRSFRR6zXSmW7y5s/632OYaOzqXlcQ/8/HVSz
03ETz6XCrgo5lMd57VR2kq+JZYmfD4H1Ogq0Pq7DqLZcxH5nY8neXjq6UqAbNcEW
Vo7Em33lrsdstggM87Nioyv1w3TOPBuTIzKt63Vwj0LtblxigPyFKT5+3iCAxrte
Iy8mAPM1tuxAV6pRZeI0Yb/wAjVj0OTGKesSRnCPB0GWjTbz+UMScIvzcadWz17r
0LddN8Krd8hS9EqDac+2SV4MbqdR/sExNZqGcVfPuhcfitcxC0xolhEbAh0lEold
RsJCqk8xAgMBAAECggEAEJwBG3LC5VMBswioi4DCwykM2qL/VmeS80hdeI21609c
l9pNHPQ6uCTbChbopkBrt7zMRriHF3k0wrL6DLb3LsOsrgFe2EK5EY+qm3OwrcOm
LbatAkQSRnHFTsF74K0/vpUGxYvmM7x6a6cVWEzoWx1l3pa/Za0kko8utwM8kzQl
UOD4bWFMYcQB/g/H4zFMKnGxxHqvXrSiUa/qHnQR5Bvckt5a+jCiR3YphnpPWCQl
tUqE6vtKfUrhaXc0KxwNE3iiqne/udxbmIFduFBaqYVg5TTzuQsPj6Jr3pz0lPXl
WWJYFewccoHM740DMiwBpTK9Wd/pfJ0Fi3ATl5L+YQKBgQC7BkMGA9UMT9U6a0tw
GDQvIp8TN8PSFsv0+FvHkgfltJDhgl0ybu6s6UiiVtwZm8WKsM+AWF+qwvZAa3hL
noRU/zcA9TvAeu8QCx/looyzuayCqvQVWyVytl0/lyLFctBFYVaiLpA1+NiRv6Zu
vEG7MWDRLRzK2CQbAEqZoBTI3QKBgQC2uwStkFoLWNmP4lb6g2chyNTCKOLNtgUK
Y/NdVNyY9VH5c5q3aYq4tmT/0KlZDTGIoR6R/IQ0+25KTyrYTFhap/yBagKYYc6u
Sxd51NmOWhbMAt1UTUoO7aaP06VShaBQgjUy+2G3k7/Rz9nF2jhmSpkyOLQM5lCB
kBvmFlpQZQKBgQCpku3mYEcl2KS/SVWnF2sJadzOMfvNW3ombaHZ++BJOEU9E1Sp
S8gA46xF9XeviHu+Wr5p4rcrP4bDti3mcp4N6zHWHoTE6zIjW9LaBV6J/soZ2CNj
0bbMoek+pSyT1pxcq/s/JfT/2teSnzCqqur2bbkZMEww53UlPkhlrq3pyQKBgBFG
lSajox+3grort/VvPuzew96nZunz734P/Q4x27lKWDmxSEtW2xqjg+D7pUcaDDjS
osVCjm1D6CV2XqKcdS38+85wa1ZkyNmJl+qYyQjAU69uBebWd835genPJK4sm/+A
j+8F/TMR8OyxLfGatAJXwywQWFVv4OSe70RNkLRRAoGAC/XiWml0/ftD6pG/Lp5J
vLFn6+MeuTYiBXOJvrSoxSWWTnj5/kZs+fLQY5iO6bm1hVl/7+/rBkLO5Q4D5+Pg
i1WoU5Te6wAvO8fRdN7DKHgQFMG5+HjTv1mc5H6OsKKSApdzO4/u+DbZyFRsBk2A
5Qpgj7SvYH/0m1HFMc9Uf6A=
-----END PRIVATE KEY-----`

// getTestPrivateKey returns the test private key
func getTestPrivateKey(t *testing.T) string {
	t.Helper()
	return testPrivateKey
}

func TestParsePrivateKey(t *testing.T) {
	tests := []struct {
		name    string
		pemData string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid key",
			pemData: getTestPrivateKey(t),
			wantErr: false,
		},
		{
			name:    "invalid PEM",
			pemData: "not a valid pem",
			wantErr: true,
			errMsg:  "failed to parse PEM block",
		},
		{
			name: "invalid key format",
			pemData: `-----BEGIN PRIVATE KEY-----
aW52YWxpZCBkYXRh
-----END PRIVATE KEY-----`,
			wantErr: true,
			errMsg:  "failed to parse private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := ParsePrivateKey(tt.pemData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			}
			if !tt.wantErr && key == nil {
				t.Error("Expected private key to be returned")
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	privateKey, err := ParsePrivateKey(getTestPrivateKey(t))
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	token, err := GenerateToken(privateKey, "test-customer", "Test Customer", now, expiresAt)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Verify the token can be parsed and validated
	validator, err := license.NewValidatorWithKey(license.DefaultPublicKey)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	claims, err := validator.Validate(token)
	if err != nil {
		t.Errorf("Failed to validate generated token: %v", err)
	}

	// Verify claims if validation succeeded
	if claims != nil {
		if claims.RegisteredClaims.Subject != "test-customer" {
			t.Errorf("Expected subject 'test-customer', got %s", claims.RegisteredClaims.Subject)
		}
		if claims.Name != "Test Customer" {
			t.Errorf("Expected name 'Test Customer', got %s", claims.Name)
		}
		if claims.Kasho.MajorVersion != 0 {
			t.Errorf("Expected major version 0 (test mode), got %d", claims.Kasho.MajorVersion)
		}
	}
}

func TestGenerateLicense(t *testing.T) {
	privateKey, err := ParsePrivateKey(getTestPrivateKey(t))
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	t.Run("stdout output", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := generateLicense(privateKey, 30, false, "", "test-id", "Test Name")
		if err != nil {
			t.Errorf("generateLicense() error = %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		// Read captured output
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		// Verify it's a JWT
		if len(output) == 0 {
			t.Error("Expected JWT output")
		}

		// Should have 3 parts separated by dots
		parts := 0
		for _, c := range output {
			if c == '.' {
				parts++
			}
		}
		if parts != 2 { // 2 dots = 3 parts
			t.Errorf("Expected JWT with 3 parts, got %d", parts+1)
		}
	})

	t.Run("file output", func(t *testing.T) {
		tmpDir := t.TempDir()
		outFile := filepath.Join(tmpDir, "test.jwt")

		err := generateLicense(privateKey, 7, false, outFile, "file-test", "File Test")
		if err != nil {
			t.Errorf("generateLicense() error = %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(outFile); os.IsNotExist(err) {
			t.Error("Expected license file to be created")
		}

		// Read and validate the license
		content, err := os.ReadFile(outFile)
		if err != nil {
			t.Fatalf("Failed to read license file: %v", err)
		}

		validator, err := license.NewValidatorWithKey(license.DefaultPublicKey)
		if err != nil {
			t.Fatalf("Failed to create validator: %v", err)
		}

		claims, err := validator.Validate(string(content))
		if err != nil {
			t.Errorf("Failed to validate license from file: %v", err)
		}

		if claims.RegisteredClaims.Subject != "file-test" {
			t.Errorf("Expected subject 'file-test', got %s", claims.RegisteredClaims.Subject)
		}
	})
}

func TestTokenExpiration(t *testing.T) {
	privateKey, err := ParsePrivateKey(getTestPrivateKey(t))
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	t.Run("expired token", func(t *testing.T) {
		// Generate a token that expired yesterday
		now := time.Now()
		issuedAt := now.Add(-48 * time.Hour)
		expiresAt := now.Add(-24 * time.Hour)

		token, err := GenerateToken(privateKey, "expired-test", "Expired Test", issuedAt, expiresAt)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		validator, err := license.NewValidatorWithKey(license.DefaultPublicKey)
		if err != nil {
			t.Fatalf("Failed to create validator: %v", err)
		}

		_, err = validator.Validate(token)
		if err == nil {
			t.Error("Expected validation to fail for expired token")
		}
		if !contains(err.Error(), "expired") {
			t.Errorf("Expected error to mention expiration, got: %v", err)
		}
	})

	t.Run("future token", func(t *testing.T) {
		// Generate a token valid for 365 days
		now := time.Now()
		expiresAt := now.Add(365 * 24 * time.Hour)

		token, err := GenerateToken(privateKey, "future-test", "Future Test", now, expiresAt)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		validator, err := license.NewValidatorWithKey(license.DefaultPublicKey)
		if err != nil {
			t.Fatalf("Failed to create validator: %v", err)
		}

		claims, err := validator.Validate(token)
		if err != nil {
			t.Errorf("Failed to validate future token: %v", err)
		}

		// Verify expiration is in the future
		if claims.RegisteredClaims.ExpiresAt.Before(now) {
			t.Error("Expected token to expire in the future")
		}
	})

	t.Run("non-expiring token", func(t *testing.T) {
		// Generate a token with no expiration
		now := time.Now()
		var zeroTime time.Time

		token, err := GenerateToken(privateKey, "perpetual-test", "Perpetual Test", now, zeroTime)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		validator, err := license.NewValidatorWithKey(license.DefaultPublicKey)
		if err != nil {
			t.Fatalf("Failed to create validator: %v", err)
		}

		claims, err := validator.Validate(token)
		if err != nil {
			t.Errorf("Failed to validate non-expiring token: %v", err)
		}

		// Verify no expiration is set
		if claims.RegisteredClaims.ExpiresAt != nil {
			t.Error("Expected ExpiresAt to be nil for non-expiring token")
		}
	})
}

func TestJWTStructure(t *testing.T) {
	privateKey, err := ParsePrivateKey(getTestPrivateKey(t))
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	token, err := GenerateToken(privateKey, "struct-test", "Structure Test", now, expiresAt)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Parse without validation to check structure
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, &license.Claims{})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	claims, ok := parsedToken.Claims.(*license.Claims)
	if !ok {
		t.Fatal("Failed to get claims from token")
	}

	// Verify all expected fields
	if claims.RegisteredClaims.Issuer != "kasho.io" {
		t.Errorf("Expected issuer 'kasho.io', got %s", claims.RegisteredClaims.Issuer)
	}
	if claims.RegisteredClaims.Subject != "struct-test" {
		t.Errorf("Expected subject 'struct-test', got %s", claims.RegisteredClaims.Subject)
	}
	if claims.Name != "Structure Test" {
		t.Errorf("Expected name 'Structure Test', got %s", claims.Name)
	}
	if claims.Kasho.MajorVersion != 0 {
		t.Errorf("Expected major version 0 (test mode), got %d", claims.Kasho.MajorVersion)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}