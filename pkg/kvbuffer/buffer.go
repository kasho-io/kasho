package kvbuffer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/redis/go-redis/v9"
)

const (
	changesKey = "pg:changes"
	changesTTL = 24 * time.Hour
)

// Change represents a database change event
type Change interface {
	Type() string
	GetLSN() string
}

// KVBuffer manages change events in a Redis-backed buffer
type KVBuffer struct {
	client *redis.Client
}

// NewKVBuffer creates a new KV buffer connected to Redis
func NewKVBuffer(kvURL string) (*KVBuffer, error) {
	opts, err := redis.ParseURL(kvURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse KV URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to KV: %w", err)
	}

	return &KVBuffer{client: client}, nil
}

// AddChange adds a change to the KV buffer with its LSN as the score
func (b *KVBuffer) AddChange(ctx context.Context, change Change) error {
	lsn := change.GetLSN()
	score, err := b.parseLSNToScore(lsn)
	if err != nil {
		return fmt.Errorf("failed to parse LSN: %w", err)
	}

	data, err := json.Marshal(change)
	if err != nil {
		return fmt.Errorf("failed to marshal change: %w", err)
	}

	err = b.client.ZAdd(ctx, changesKey, redis.Z{
		Score:  score,
		Member: data,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to add change to KV: %w", err)
	}

	err = b.client.Expire(ctx, changesKey, changesTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set TTL: %w", err)
	}

	err = b.client.Publish(ctx, "pg:changes", data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish change: %w", err)
	}

	return nil
}

// GetChangesAfter returns all changes after the given LSN
func (b *KVBuffer) GetChangesAfter(ctx context.Context, lsn string) ([]json.RawMessage, error) {
	score, err := b.parseLSNToScore(lsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LSN: %w", err)
	}

	results, err := b.client.ZRangeByScore(ctx, changesKey, &redis.ZRangeBy{
		// (%d --> exclude the score itself, > and not >=
		Min:    fmt.Sprintf("(%g", score),
		Max:    "+inf",
		Offset: 0,
		Count:  1000,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get changes from KV: %w", err)
	}

	changes := make([]json.RawMessage, len(results))
	for i, result := range results {
		changes[i] = json.RawMessage(result)
	}

	return changes, nil
}

// parseLSNToScore converts an LSN string to a float64 score for Redis sorted set
func (b *KVBuffer) parseLSNToScore(lsn string) (float64, error) {
	// Handle synthetic bootstrap LSNs first (e.g., "0/BOOTSTRAP00000001")
	if len(lsn) > 2 && lsn[:2] == "0/" && len(lsn) > 11 && lsn[2:11] == "BOOTSTRAP" {
		var seq int64
		if n, err := fmt.Sscanf(lsn[2:], "BOOTSTRAP%08d", &seq); n == 1 && err == nil {
			// Use negative scores for bootstrap LSNs to ensure they sort before real LSNs
			return float64(-1000000 + seq), nil
		}
		// If it starts with BOOTSTRAP but doesn't match the pattern, it's invalid
		return 0, fmt.Errorf("invalid bootstrap LSN format: %s", lsn)
	}

	// Handle PostgreSQL LSN format (e.g., "0/100")
	if parsedLSN, err := pglogrepl.ParseLSN(lsn); err == nil {
		return float64(parsedLSN), nil
	}

	return 0, fmt.Errorf("invalid LSN format: %s", lsn)
}

// Subscribe creates a Redis pubsub subscription
func (b *KVBuffer) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return b.client.Subscribe(ctx, channel)
}

// GetClient returns the underlying Redis client for advanced operations
func (b *KVBuffer) GetClient() *redis.Client {
	return b.client
}

// Close closes the KV connection
func (b *KVBuffer) Close() error {
	return b.client.Close()
}