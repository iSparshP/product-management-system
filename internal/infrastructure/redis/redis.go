// internal/infrastructure/redis/redis.go

package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	*redis.Client
}

func NewRedisClient(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		// Add more options if needed, e.g., Password, DB, etc.
	})
	return &Client{rdb}
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	result := c.Client.Set(ctx, key, value, expiration)
	return result.Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	result := c.Client.Get(ctx, key)
	return result.Result()
}

func (c *Client) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return c.Client.Del(ctx, keys...)
}

// SetEX sets a key with expiration
func (c *Client) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Set(ctx, key, value, expiration)
}

// GetWithTTL gets a value and its remaining TTL
func (c *Client) GetWithTTL(ctx context.Context, key string) (string, time.Duration, error) {
	pipe := c.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return "", 0, err
	}

	value, err := getCmd.Result()
	if err != nil {
		return "", 0, err
	}

	ttl, err := ttlCmd.Result()
	if err != nil {
		return "", 0, err
	}

	return value, ttl, nil
}

// SetNX sets a key only if it doesn't exist (useful for locks)
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, expiration).Result()
}

// Exists checks if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.Client.Exists(ctx, key).Result()
	return n > 0, err
}
