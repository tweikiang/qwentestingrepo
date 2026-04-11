package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"brave-search-clone/config"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
}

// Use the SearchResult from strategy package to avoid type conflicts
type CachedResult struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
}

func NewCache(cfg *config.RedisConfig) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Cache{
		client: client,
		ctx:    context.Background(),
	}
}

func (c *Cache) Get(query string) ([]CachedResult, bool) {
	key := "search:" + query
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var results []CachedResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, false
	}

	return results, true
}

func (c *Cache) Set(query string, results []CachedResult, ttl time.Duration) error {
	key := "search:" + query
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}

	return c.client.Set(c.ctx, key, data, ttl).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}
