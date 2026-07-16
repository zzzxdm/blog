package cachex

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// TTLStore is a small string cache with TTL, used for public read endpoints.
type TTLStore interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key string, value string, ttl time.Duration)
	Delete(ctx context.Context, keys ...string)
}

type memoryTTLStore struct {
	mu    sync.RWMutex
	items map[string]memoryTTLItem
}

type memoryTTLItem struct {
	value     string
	expiresAt time.Time
}

type redisTTLStore struct {
	client *redis.Client
	memory *memoryTTLStore
}

// NewTTLStore returns Redis-backed cache when available, otherwise process memory.
func NewTTLStore(client *redis.Client) TTLStore {
	if client != nil {
		return &redisTTLStore{client: client, memory: newMemoryTTLStore()}
	}
	return newMemoryTTLStore()
}

func newMemoryTTLStore() *memoryTTLStore {
	return &memoryTTLStore{items: map[string]memoryTTLItem{}}
}

func (s *memoryTTLStore) Get(_ context.Context, key string) (string, bool) {
	s.mu.RLock()
	item, ok := s.items[key]
	s.mu.RUnlock()
	if !ok {
		return "", false
	}
	if time.Now().After(item.expiresAt) {
		s.mu.Lock()
		delete(s.items, key)
		s.mu.Unlock()
		return "", false
	}
	return item.value, true
}

func (s *memoryTTLStore) Set(_ context.Context, key string, value string, ttl time.Duration) {
	if key == "" || ttl <= 0 {
		return
	}
	s.mu.Lock()
	s.items[key] = memoryTTLItem{value: value, expiresAt: time.Now().Add(ttl)}
	s.mu.Unlock()
}

func (s *memoryTTLStore) Delete(_ context.Context, keys ...string) {
	s.mu.Lock()
	for _, key := range keys {
		delete(s.items, key)
	}
	s.mu.Unlock()
}

func (s *redisTTLStore) Get(ctx context.Context, key string) (string, bool) {
	value, err := s.client.Get(ctx, redisCacheKey(key)).Result()
	if err == redis.Nil {
		return "", false
	}
	if err != nil {
		return s.memory.Get(ctx, key)
	}
	return value, true
}

func (s *redisTTLStore) Set(ctx context.Context, key string, value string, ttl time.Duration) {
	if key == "" || ttl <= 0 {
		return
	}
	if err := s.client.Set(ctx, redisCacheKey(key), value, ttl).Err(); err != nil {
		s.memory.Set(ctx, key, value, ttl)
	}
}

func (s *redisTTLStore) Delete(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	redisKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		redisKeys = append(redisKeys, redisCacheKey(key))
	}
	if err := s.client.Del(ctx, redisKeys...).Err(); err != nil {
		s.memory.Delete(ctx, keys...)
	}
}

func redisCacheKey(key string) string {
	return "pubcache:" + key
}

// Common public cache keys.
const (
	CacheKeyPublicSettings    = "public:settings"
	CacheKeyPublicNavigation  = "public:navigation"
	CacheKeySiteStats         = "public:site-stats"
	PublicCacheTTL            = 45 * time.Second
)
