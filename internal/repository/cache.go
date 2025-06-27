package repository

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache implements caching functionality
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(redisURL string) *RedisCache {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// Fallback to default configuration
		opt = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	client := redis.NewClient(opt)

	return &RedisCache{
		client: client,
		ctx:    context.Background(),
		ttl:    24 * time.Hour, // 24 hour TTL
	}
}

// Get retrieves a value from cache
func (c *RedisCache) Get(key string) (string, error) {
	return c.client.Get(c.ctx, key).Result()
}

// Set stores a value in cache with TTL
func (c *RedisCache) Set(key, value string) error {
	return c.client.Set(c.ctx, key, value, c.ttl).Err()
}

// SetWithTTL stores a value in cache with custom TTL
func (c *RedisCache) SetWithTTL(key, value string, ttl time.Duration) error {
	return c.client.Set(c.ctx, key, value, ttl).Err()
}

// Delete removes a value from cache
func (c *RedisCache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Ping checks if Redis is accessible
func (c *RedisCache) Ping() error {
	return c.client.Ping(c.ctx).Err()
}
