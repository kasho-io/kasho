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
// NOTE: This method is limited to 1000 changes for backward compatibility.
// Use GetChangesAfterBatch for paginated access to larger result sets.
func (b *KVBuffer) GetChangesAfter(ctx context.Context, lsn string) ([]json.RawMessage, error) {
	return b.GetChangesAfterBatch(ctx, lsn, 0, 1000)
}

// GetChangesAfterBatch returns a batch of changes after the given LSN with offset and limit
func (b *KVBuffer) GetChangesAfterBatch(ctx context.Context, lsn string, offset int64, limit int64) ([]json.RawMessage, error) {
	score, err := b.parseLSNToScore(lsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LSN: %w", err)
	}

	// Special case: LSN "0/0" means get all changes including bootstrap
	minScore := fmt.Sprintf("(%g", score)
	if lsn == "0/0" {
		minScore = "-inf"
	}

	results, err := b.client.ZRangeByScore(ctx, changesKey, &redis.ZRangeBy{
		Min:    minScore,
		Max:    "+inf",
		Offset: offset,
		Count:  limit,
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
	// Handle synthetic bootstrap LSNs first (e.g., "0/BOOTSTRAP0000000000000001")
	if len(lsn) > 2 && lsn[:2] == "0/" && len(lsn) > 11 && lsn[2:11] == "BOOTSTRAP" {
		var seq int64
		if n, err := fmt.Sscanf(lsn[2:], "BOOTSTRAP%016d", &seq); n == 1 && err == nil {
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

// Get retrieves a value from Redis by key
func (b *KVBuffer) Get(ctx context.Context, key string) (string, error) {
	val, err := b.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// Set stores a value in Redis with the given key
func (b *KVBuffer) Set(ctx context.Context, key, value string) error {
	err := b.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Close closes the KV connection
func (b *KVBuffer) Close() error {
	return b.client.Close()
}