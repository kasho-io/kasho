package server

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

type RedisBuffer struct {
	client *redis.Client
}

func NewRedisBuffer(redisURL string) (*RedisBuffer, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisBuffer{client: client}, nil
}

// AddChange adds a change to the Redis buffer with its LSN as the score
func (b *RedisBuffer) AddChange(ctx context.Context, lsn string, change Change) error {
	score, err := pglogrepl.ParseLSN(lsn)
	if err != nil {
		return fmt.Errorf("failed to parse LSN: %w", err)
	}

	data, err := json.Marshal(change)
	if err != nil {
		return fmt.Errorf("failed to marshal change: %w", err)
	}

	err = b.client.ZAdd(ctx, changesKey, redis.Z{
		Score:  float64(score),
		Member: data,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to add change to Redis: %w", err)
	}

	err = b.client.Expire(ctx, changesKey, changesTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set TTL: %w", err)
	}

	return nil
}

// GetChangesAfter returns all changes after the given LSN
func (b *RedisBuffer) GetChangesAfter(ctx context.Context, lsn string) ([]Change, error) {
	score, err := pglogrepl.ParseLSN(lsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LSN: %w", err)
	}

	results, err := b.client.ZRangeByScore(ctx, changesKey, &redis.ZRangeBy{
		// (%d --> exclude the score itself, > and not >=
		Min:    fmt.Sprintf("(%d", score),
		Max:    "+inf",
		Offset: 0,
		Count:  1000,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get changes from Redis: %w", err)
	}

	changes := make([]Change, 0, len(results))
	for _, result := range results {
		var change Change
		if err := json.Unmarshal([]byte(result), &change); err != nil {
			return nil, fmt.Errorf("failed to unmarshal change: %w", err)
		}
		changes = append(changes, change)
	}

	return changes, nil
}

// Close closes the Redis connection
func (b *RedisBuffer) Close() error {
	return b.client.Close()
}
