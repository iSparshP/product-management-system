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
