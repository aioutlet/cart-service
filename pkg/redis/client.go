package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Address  string
	Password string
	DB       int
	PoolSize int
}

// NewClient creates a new Redis client
func NewClient(cfg RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test the connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
