package main

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)


func main() {
	var (
		privateKeyPath = flag.String("private-key", "", "Path to private key file (required)")
		days           = flag.Int("days", 0, "Number of days until license expires")
		noExpiration   = flag.Bool("no-expiration", false, "Generate a license that never expires")
		output         = flag.String("output", "", "Output file path (defaults to stdout)")
		customerID     = flag.String("customer-id", "", "Customer ID (required)")
		customerName   = flag.String("customer-name", "", "Customer name (required)")
	)
	flag.Parse()

	if *privateKeyPath == "" {
		log.Fatal("Private key path is required (-private-key)")
	}

	if *customerID == "" {
		log.Fatal("Customer ID is required (-customer-id)")
	}

	if *customerName == "" {
		log.Fatal("Customer name is required (-customer-name)")
	}

	// Check for conflicting flags
	if *noExpiration && *days != 0 {
		log.Fatal("Cannot specify both -days and -no-expiration flags")
	}

	// Ensure at least one expiration option is specified
	if !*noExpiration && *days == 0 {
		log.Fatal("Must specify either -days or -no-expiration")
	}

	// Read private key from file
	privateKeyData, err := os.ReadFile(*privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key file: %v", err)
	}

	// Parse the private key
	privateKey, err := ParsePrivateKey(string(privateKeyData))
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	if err := generateLicense(privateKey, *days, *noExpiration, *output, *customerID, *customerName); err != nil {
		log.Fatalf("Failed to generate license: %v", err)
	}
	if *noExpiration {
		log.Printf("Generated non-expiring license")
	} else {
		log.Printf("Generated license valid for %d days", *days)
	}
}

func generateLicense(privateKey *rsa.PrivateKey, days int, noExpiration bool, output, customerID, customerName string) error {
	now := time.Now()
	var expiresAt time.Time
	
	if !noExpiration {
		expiresAt = now.AddDate(0, 0, days)
	}
	// If noExpiration is true, expiresAt remains zero value

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
		if noExpiration {
			log.Printf("Expires: Never")
		} else {
			log.Printf("Expires: %s", expiresAt.Format(time.RFC3339))
		}
	}

	return nil
}