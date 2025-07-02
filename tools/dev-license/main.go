package main

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"kasho/pkg/license"
)

// DevPrivateKey is the private key for development licenses
// This matches the public key embedded in the licensing service
const DevPrivateKey = `-----BEGIN PRIVATE KEY-----
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

func main() {
	var (
		days         = flag.Int("days", 30, "Number of days until license expires")
		output       = flag.String("output", "", "Output file path (defaults to stdout)")
		customerID   = flag.String("customer-id", "dev-customer", "Customer ID")
		customerName = flag.String("customer-name", "Development Customer", "Customer name")
		watch        = flag.Bool("watch", false, "Watch mode - regenerate license before expiration")
		watchBuffer  = flag.Int("watch-buffer", 1, "Days before expiration to regenerate (only with -watch)")
	)
	flag.Parse()

	if *watch && *output == "" {
		log.Fatal("Watch mode requires an output file (-output)")
	}

	// Parse the private key
	privateKey, err := ParsePrivateKey(DevPrivateKey)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	if *watch {
		watchAndRegenerate(privateKey, *days, *output, *customerID, *customerName, *watchBuffer)
	} else {
		if err := generateLicense(privateKey, *days, *output, *customerID, *customerName); err != nil {
			log.Fatalf("Failed to generate license: %v", err)
		}
		log.Printf("Generated license valid for %d days", *days)
	}
}

func generateLicense(privateKey *rsa.PrivateKey, days int, output, customerID, customerName string) error {
	now := time.Now()
	expiresAt := now.AddDate(0, 0, days)

	tokenString, err := GenerateToken(privateKey, customerID, customerName, now, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	if output == "" {
		fmt.Println(tokenString)
	} else {
		if err := os.WriteFile(output, []byte(tokenString), 0644); err != nil {
			return fmt.Errorf("failed to write license file: %w", err)
		}
		log.Printf("License written to %s", output)
		log.Printf("Expires: %s", expiresAt.Format(time.RFC3339))
	}

	return nil
}

func watchAndRegenerate(privateKey *rsa.PrivateKey, days int, output, customerID, customerName string, bufferDays int) {
	log.Printf("Watch mode enabled. Will regenerate license %d days before expiration", bufferDays)
	
	// Generate initial license
	if err := generateLicense(privateKey, days, output, customerID, customerName); err != nil {
		log.Fatalf("Failed to generate initial license: %v", err)
	}

	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("🔍 Checking license status...")
		
		// Check if license needs regeneration
		validator, err := license.NewValidator()
		if err != nil {
			log.Printf("Warning: Failed to create validator: %v", err)
			continue
		}

		claims, err := validator.ValidateFromFile(output)
		if err != nil {
			log.Printf("License validation failed: %v", err)
			log.Printf("Regenerating license due to validation error...")
			if err := generateLicense(privateKey, days, output, customerID, customerName); err != nil {
				log.Printf("Error: Failed to regenerate license: %v", err)
			}
			continue
		}

		log.Printf("License validation successful")

		if claims.ExpiresAt != nil {
			daysUntilExpiry := time.Until(claims.ExpiresAt.Time).Hours() / 24
			log.Printf("Days until expiry: %.3f (threshold: %d)", daysUntilExpiry, bufferDays)
			
			if daysUntilExpiry <= float64(bufferDays) {
				log.Printf("Regeneration needed: true - License expires in %.1f days, regenerating...", daysUntilExpiry)
				if err := generateLicense(privateKey, days, output, customerID, customerName); err != nil {
					log.Printf("Error: Failed to regenerate license: %v", err)
				}
			} else {
				log.Printf("Regeneration needed: false - License is still valid for %.1f days", daysUntilExpiry)
			}
		} else {
			log.Printf("Warning: License has no expiration time")
		}
	}
}