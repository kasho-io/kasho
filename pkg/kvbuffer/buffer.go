package kvbuffer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/redis/go-redis/v9"
)

const (
	changesKey    = "kasho:changes"
	changesChannel = "kasho:changes"
	changesTTL    = 24 * time.Hour
)

// Change represents a database change event
// GetPosition returns the position identifier (PostgreSQL LSN, MySQL binlog position, etc.)
type Change interface {
	Type() string
	GetPosition() string
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

// AddChange adds a change to the KV buffer with its position as the score
func (b *KVBuffer) AddChange(ctx context.Context, change Change) error {
	position := change.GetPosition()
	score, err := b.parsePositionToScore(position)
	if err != nil {
		return fmt.Errorf("failed to parse position: %w", err)
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

	err = b.client.Publish(ctx, changesChannel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish change: %w", err)
	}

	return nil
}

// GetChangesAfter returns all changes after the given position
// NOTE: This method is limited to 1000 changes for backward compatibility.
// Use GetChangesAfterBatch for paginated access to larger result sets.
func (b *KVBuffer) GetChangesAfter(ctx context.Context, position string) ([]json.RawMessage, error) {
	return b.GetChangesAfterBatch(ctx, position, 0, 1000)
}

// GetChangesAfterBatch returns a batch of changes after the given position with offset and limit
func (b *KVBuffer) GetChangesAfterBatch(ctx context.Context, position string, offset int64, limit int64) ([]json.RawMessage, error) {
	score, err := b.parsePositionToScore(position)
	if err != nil {
		return nil, fmt.Errorf("failed to parse position: %w", err)
	}

	// Special case: position "0/0" means get all changes including bootstrap
	minScore := fmt.Sprintf("(%g", score)
	if position == "0/0" {
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

// parsePositionToScore converts a database position to a Redis sorted set score
// Supports:
// - PostgreSQL LSN: "0/100" format
// - MySQL binlog: "mysql-bin.000001:4" format (filename:offset)
// - Bootstrap: "0/BOOTSTRAP%016d" format (negative scores for ordering)
func (b *KVBuffer) parsePositionToScore(position string) (float64, error) {
	// Bootstrap positions: "0/BOOTSTRAP%016d" → -1000000 + seq
	if len(position) > 2 && position[:2] == "0/" && len(position) >= 11 && position[2:11] == "BOOTSTRAP" {
		var seq int64
		if n, err := fmt.Sscanf(position[2:], "BOOTSTRAP%016d", &seq); n == 1 && err == nil {
			return float64(-1000000 + seq), nil
		}
		return 0, fmt.Errorf("invalid bootstrap position format: %s", position)
	}

	// MySQL binlog position: "mysql-bin.000001:4" or "binlog.000001:4" → (filenum * 4294967296) + offset
	if strings.Contains(position, ":") && (strings.Contains(position, "bin.") || strings.HasPrefix(position, "binlog.")) {
		parts := strings.Split(position, ":")
		if len(parts) == 2 {
			// Extract file number from "mysql-bin.000001" or "binlog.000001"
			filename := parts[0]
			if idx := strings.LastIndex(filename, "."); idx != -1 {
				fileNumStr := filename[idx+1:]
				fileNum, err := strconv.ParseInt(fileNumStr, 10, 64)
				if err != nil {
					return 0, fmt.Errorf("invalid MySQL binlog file number: %s", position)
				}
				offset, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					return 0, fmt.Errorf("invalid MySQL binlog offset: %s", position)
				}
				// Combine: file number * 4GB + offset for monotonic ordering
				return float64(fileNum)*4294967296 + float64(offset), nil
			}
		}
		return 0, fmt.Errorf("invalid MySQL binlog format: %s", position)
	}

	// PostgreSQL LSN: "0/100" → float64
	if parsedLSN, err := pglogrepl.ParseLSN(position); err == nil {
		return float64(parsedLSN), nil
	}

	return 0, fmt.Errorf("invalid position format: %s", position)
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