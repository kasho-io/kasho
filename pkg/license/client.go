package license

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kasho/proto/kasho/proto"
)

const (
	defaultTimeout      = 5 * time.Second
	cacheExpiry         = 5 * time.Minute
	defaultLicenseAddr  = "licensing:50053"
)

// Client provides methods to interact with the licensing service
type Client struct {
	conn         *grpc.ClientConn
	client       proto.LicenseClient
	cacheMu      sync.RWMutex
	cachedInfo   *proto.GetLicenseInfoResponse
	cacheTime    time.Time
	allowOffline bool
}

// Config holds configuration for the license client
type Config struct {
	// Address of the licensing service (defaults to "licensing:50053")
	Address string
	// AllowOffline allows the service to continue if licensing service is unavailable (for development)
	AllowOffline bool
	// Timeout for RPC calls (defaults to 5 seconds)
	Timeout time.Duration
}

// NewClient creates a new license client
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.Address == "" {
		cfg.Address = defaultLicenseAddr
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	conn, err := grpc.Dial(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		if cfg.AllowOffline {
			log.Printf("Warning: Failed to connect to licensing service at %s: %v (continuing in offline mode)", cfg.Address, err)
			return &Client{
				allowOffline: true,
			}, nil
		}
		return nil, fmt.Errorf("failed to connect to licensing service: %w", err)
	}

	return &Client{
		conn:         conn,
		client:       proto.NewLicenseClient(conn),
		allowOffline: cfg.AllowOffline,
	}, nil
}

// Close closes the connection to the licensing service
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ValidateLicense checks if the current license is valid
func (c *Client) ValidateLicense(ctx context.Context) error {
	if c.allowOffline && c.client == nil {
		log.Println("Warning: Running in offline mode, license validation skipped")
		return nil
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
	}

	resp, err := c.client.ValidateLicense(ctx, &proto.ValidateLicenseRequest{})
	if err != nil {
		if c.allowOffline {
			log.Printf("Warning: License validation failed: %v (continuing in offline mode)", err)
			return nil
		}
		return fmt.Errorf("license validation failed: %w", err)
	}

	if !resp.Valid {
		return fmt.Errorf("license is invalid: %s", resp.Reason)
	}

	// Log successful validation with expiry info
	if resp.ExpiresAt > 0 {
		expiresAt := time.Unix(resp.ExpiresAt, 0)
		daysUntilExpiry := time.Until(expiresAt).Hours() / 24
		if daysUntilExpiry < 30 {
			log.Printf("Warning: License expires in %.0f days on %s", daysUntilExpiry, expiresAt.Format(time.RFC3339))
		} else {
			log.Printf("License validated successfully, expires in %.0f days", daysUntilExpiry)
		}
	} else {
		log.Printf("License validated successfully")
	}

	return nil
}

// GetLicenseInfo returns information about the current license (with caching)
func (c *Client) GetLicenseInfo(ctx context.Context) (*proto.GetLicenseInfoResponse, error) {
	if c.allowOffline && c.client == nil {
		return &proto.GetLicenseInfoResponse{
			Valid:        true,
			CustomerName: "Development Mode",
		}, nil
	}

	// Check cache
	c.cacheMu.RLock()
	if c.cachedInfo != nil && time.Since(c.cacheTime) < cacheExpiry {
		info := c.cachedInfo
		c.cacheMu.RUnlock()
		return info, nil
	}
	c.cacheMu.RUnlock()

	// Fetch from service
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
	}

	resp, err := c.client.GetLicenseInfo(ctx, &proto.GetLicenseInfoRequest{})
	if err != nil {
		if c.allowOffline {
			log.Printf("Warning: Failed to get license info: %v (returning offline mode info)", err)
			return &proto.GetLicenseInfoResponse{
				Valid:        true,
				CustomerName: "Development Mode",
			}, nil
		}
		return nil, fmt.Errorf("failed to get license info: %w", err)
	}

	// Update cache
	c.cacheMu.Lock()
	c.cachedInfo = resp
	c.cacheTime = time.Now()
	c.cacheMu.Unlock()

	return resp, nil
}

// MustValidate validates the license and exits the program if invalid
func (c *Client) MustValidate(ctx context.Context) {
	if err := c.ValidateLicense(ctx); err != nil {
		log.Fatalf("License validation failed: %v", err)
	}
}