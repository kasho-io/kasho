package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pg-change-stream/internal/types"

	"github.com/jackc/pglogrepl"
	"github.com/redis/go-redis/v9"
)

const (
	changesKey = "pg:changes"
	changesTTL = 24 * time.Hour
)

type KVBuffer struct {
	client *redis.Client
}

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
func (b *KVBuffer) AddChange(ctx context.Context, lsn string, change types.Change) error {
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
func (b *KVBuffer) GetChangesAfter(ctx context.Context, lsn string) ([]types.Change, error) {
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
		return nil, fmt.Errorf("failed to get changes from KV: %w", err)
	}

	changes := make([]types.Change, 0, len(results))
	for _, result := range results {
		var change types.Change
		if err := json.Unmarshal([]byte(result), &change); err != nil {
			return nil, fmt.Errorf("failed to unmarshal change: %w", err)
		}
		changes = append(changes, change)
	}

	return changes, nil
}

// Close closes the KV connection
func (b *KVBuffer) Close() error {
	return b.client.Close()
}
