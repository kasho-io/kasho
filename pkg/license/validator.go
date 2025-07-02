package license

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"kasho/pkg/version"
)

// DefaultPublicKey is the embedded public key for license validation
// This will be replaced with the actual public key in production
const DefaultPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhX8UpCHa9LHLUcP1VJ7u
2wKTLENfzPq0jUJEuPUW00UhUUes10plu8ubP+t9jmGjs6l5XEP/Px1Us9NxE8+l
wq4KOZTHee1UdpKviWWJnw+B9ToKtD6uw6i2XMR+Z2PJ3l46ulKgGzXBFlaOxJt9
5a7HbLYIDPOzYqMr9cN0zjwbkyMyret1cI9C7W5cYoD8hSk+ft4ggMa7XiMvJgDz
NbbsQFeqUWXiNGG/8AI1Y9DkxinrEkZwjwdBlo028/lDEnCL83GnVs9e69C3XTfC
q3fIUvRKg2nPtkleDG6nUf7BMTWahnFXz7oXH4rXMQtMaJYRGwIdJRKJXUbCQqpP
MQIDAQAB
-----END PUBLIC KEY-----`

type Validator struct {
	publicKey *rsa.PublicKey
}


func NewValidator() (*Validator, error) {
	return NewValidatorWithKey(DefaultPublicKey)
}

func NewValidatorWithKey(publicKeyPEM string) (*Validator, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key type is not RSA")
	}

	return &Validator{
		publicKey: rsaPub,
	}, nil
}

func (v *Validator) ValidateFromFile(licensePath string) (*Claims, error) {
	licenseData, err := os.ReadFile(licensePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read license file: %w", err)
	}

	return v.Validate(string(licenseData))
}

func (v *Validator) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("license expired on %s", claims.ExpiresAt.Format(time.RFC3339))
	}

	// Check major version compatibility
	if claims.Kasho.MajorVersion > version.MajorVersion() {
		return nil, fmt.Errorf("license requires Kasho major version %d, but running version %d",
			claims.Kasho.MajorVersion, version.MajorVersion())
	}

	return claims, nil
}