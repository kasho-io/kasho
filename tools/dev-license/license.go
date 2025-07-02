package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"kasho/pkg/license"
	"kasho/pkg/version"
)

// ParsePrivateKey parses a PEM-encoded private key
func ParsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not an RSA private key")
	}

	return privateKey, nil
}

// GenerateToken generates a JWT token for the given parameters
func GenerateToken(privateKey *rsa.PrivateKey, customerID, customerName string, issuedAt, expiresAt time.Time) (string, error) {
	claims := license.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "kasho.io",
			Subject:   customerID,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Name: customerName,
		Kasho: license.KashoClaims{
			MajorVersion: version.MajorVersion(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// GenerateLicenseWithTimes generates a license with specific issue and expiry times
func GenerateLicenseWithTimes(privateKey *rsa.PrivateKey, customerID, customerName string, issuedAt, expiresAt time.Time) (string, error) {
	return GenerateToken(privateKey, customerID, customerName, issuedAt, expiresAt)
}