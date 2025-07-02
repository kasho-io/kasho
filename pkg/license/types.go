package license

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims for a Kasho license
type Claims struct {
	jwt.RegisteredClaims
	Name  string       `json:"name"`  // Customer name
	Kasho KashoClaims `json:"kasho"`
}

// KashoClaims represents Kasho-specific claims
type KashoClaims struct {
	MajorVersion int `json:"major_version"`
}