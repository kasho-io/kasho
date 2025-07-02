package server

import (
	"context"
	"log"
	"sync"
	"time"

	"kasho/pkg/license"
	"kasho/proto"
)

const (
	defaultLicensePath = "/app/config/license.jwt"
	cacheExpiry        = 5 * time.Minute
)

type Server struct {
	proto.UnimplementedLicenseServer
	validator    *license.Validator
	licensePath  string
	cacheMu      sync.RWMutex
	cachedClaims *license.Claims
	cacheTime    time.Time
}

func New() (*Server, error) {
	validator, err := license.NewValidator()
	if err != nil {
		return nil, err
	}

	s := &Server{
		validator:   validator,
		licensePath: defaultLicensePath,
	}

	// Pre-validate the license on startup
	if _, err := s.getClaims(); err != nil {
		log.Printf("Warning: License validation failed on startup: %v", err)
	}

	return s, nil
}

func (s *Server) getClaims() (*license.Claims, error) {
	s.cacheMu.RLock()
	if s.cachedClaims != nil && time.Since(s.cacheTime) < cacheExpiry {
		claims := s.cachedClaims
		s.cacheMu.RUnlock()
		return claims, nil
	}
	s.cacheMu.RUnlock()

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Double-check after acquiring write lock
	if s.cachedClaims != nil && time.Since(s.cacheTime) < cacheExpiry {
		return s.cachedClaims, nil
	}

	claims, err := s.validator.ValidateFromFile(s.licensePath)
	if err != nil {
		return nil, err
	}

	s.cachedClaims = claims
	s.cacheTime = time.Now()
	return claims, nil
}

func (s *Server) ValidateLicense(ctx context.Context, req *proto.ValidateLicenseRequest) (*proto.ValidateLicenseResponse, error) {
	claims, err := s.getClaims()
	if err != nil {
		return &proto.ValidateLicenseResponse{
			Valid:  false,
			Reason: err.Error(),
		}, nil
	}

	var expiresAt int64
	if claims.RegisteredClaims.ExpiresAt != nil {
		expiresAt = claims.RegisteredClaims.ExpiresAt.Unix()
	}

	return &proto.ValidateLicenseResponse{
		Valid:     true,
		Reason:    "",
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Server) GetLicenseInfo(ctx context.Context, req *proto.GetLicenseInfoRequest) (*proto.GetLicenseInfoResponse, error) {
	claims, err := s.getClaims()
	if err != nil {
		return &proto.GetLicenseInfoResponse{
			Valid: false,
		}, nil
	}

	var issuedAt, expiresAt int64
	if claims.RegisteredClaims.IssuedAt != nil {
		issuedAt = claims.RegisteredClaims.IssuedAt.Unix()
	}
	if claims.RegisteredClaims.ExpiresAt != nil {
		expiresAt = claims.RegisteredClaims.ExpiresAt.Unix()
	}

	return &proto.GetLicenseInfoResponse{
		CustomerId:   claims.RegisteredClaims.Subject,
		CustomerName: claims.Name,
		IssuedAt:     issuedAt,
		ExpiresAt:    expiresAt,
		Valid:        true,
	}, nil
}