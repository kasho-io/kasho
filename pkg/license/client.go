package license

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kasho/proto"
)

const (
	defaultTimeout      = 5 * time.Second
	cacheExpiry         = 5 * time.Minute
	defaultLicenseAddr  = "licensing:50052"
)

// Client provides methods to interact with the licensing service
type Client struct {
	conn       *grpc.ClientConn
	client     proto.LicenseClient
	cacheMu    sync.RWMutex
	cachedInfo *proto.GetLicenseInfoResponse
	cacheTime  time.Time
}

// Config holds configuration for the license client
type Config struct {
	// Address of the licensing service (defaults to "licensing:50052")
	Address string
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
		return nil, fmt.Errorf("failed to connect to licensing service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: proto.NewLicenseClient(conn),
	}, nil
}

// NewClientFromConn creates a new license client from an existing gRPC connection
// This is primarily used for testing
func NewClientFromConn(conn *grpc.ClientConn) *Client {
	return &Client{
		conn:   conn,
		client: proto.NewLicenseClient(conn),
	}
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
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
	}

	resp, err := c.client.ValidateLicense(ctx, &proto.ValidateLicenseRequest{})
	if err != nil {
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

// StartPeriodicValidation starts a goroutine that periodically validates the license
// Returns a channel that will be closed when validation fails
func (c *Client) StartPeriodicValidation(ctx context.Context, interval time.Duration) <-chan struct{} {
	failChan := make(chan struct{})
	
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer close(failChan)
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.ValidateLicense(ctx); err != nil {
					log.Printf("License validation failed: %v", err)
					return
				}
			}
		}
	}()
	
	return failChan
}