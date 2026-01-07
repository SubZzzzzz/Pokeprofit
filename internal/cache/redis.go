package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps redis.Client with additional functionality.
type Client struct {
	*redis.Client
}

// Config holds Redis connection settings.
type Config struct {
	URL          string
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
}

// DefaultConfig returns default Redis configuration.
func DefaultConfig(url string) Config {
	return Config{
		URL:          url,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 2,
	}
}

// New creates a new Redis client.
func New(cfg Config) (*Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	opts.MaxRetries = cfg.MaxRetries
	opts.PoolSize = cfg.PoolSize
	opts.MinIdleConns = cfg.MinIdleConns

	client := redis.NewClient(opts)

	return &Client{Client: client}, nil
}

// Connect creates a new Redis connection with context timeout.
func Connect(ctx context.Context, url string) (*Client, error) {
	cfg := DefaultConfig(url)
	client, err := New(cfg)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

// Health checks the Redis connection.
func (c *Client) Health(ctx context.Context) error {
	return c.Ping(ctx).Err()
}

// SetWithTTL sets a key with a TTL.
func (c *Client) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl).Err()
}

// GetString gets a string value by key.
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	val, err := c.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Increment increments a key's value and returns the new value.
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	return c.Incr(ctx, key).Result()
}

// IncrementWithExpiry increments a key and sets an expiry if not exists.
func (c *Client) IncrementWithExpiry(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := c.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incrCmd.Val(), nil
}

// Delete removes a key.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.Del(ctx, keys...).Err()
}

// Exists checks if a key exists.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.Client.Exists(ctx, key).Result()
	return n > 0, err
}
