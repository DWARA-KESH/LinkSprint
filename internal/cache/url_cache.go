package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type URLCache struct {
	client *redis.Client
}

func NewURLCache(client *redis.Client) *URLCache {
	return &URLCache{client: client}
}

func (c *URLCache) Set(ctx context.Context, shortCode, originalURL string, expiration time.Duration) error {
	return c.client.Set(ctx, shortCode, originalURL, expiration).Err()
}

func (c *URLCache) Get(ctx context.Context, shortCode string) (string, error) {
	return c.client.Get(ctx, shortCode).Result()
}
