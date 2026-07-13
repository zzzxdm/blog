package auth

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoginLockStore tracks failed login attempts and temporary locks.
type LoginLockStore interface {
	IsLocked(ctx context.Context, email string, now time.Time) bool
	RecordFailure(ctx context.Context, email string, now time.Time) bool
	Clear(ctx context.Context, email string)
}

type memoryLoginLockStore struct {
	mu       sync.Mutex
	failures map[string]loginFailure
}

func newMemoryLoginLockStore() *memoryLoginLockStore {
	return &memoryLoginLockStore{failures: map[string]loginFailure{}}
}

func (store *memoryLoginLockStore) IsLocked(_ context.Context, email string, now time.Time) bool {
	key := normalizeEmail(email)
	if key == "" {
		return false
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	failure := store.failures[key]
	if failure.LockedUntil.IsZero() {
		return false
	}
	if failure.LockedUntil.After(now) {
		return true
	}

	delete(store.failures, key)
	return false
}

func (store *memoryLoginLockStore) RecordFailure(_ context.Context, email string, now time.Time) bool {
	key := normalizeEmail(email)
	if key == "" {
		return false
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	failure := store.failures[key]
	failure.Count++
	if failure.Count >= loginFailureLimit {
		failure.LockedUntil = now.Add(loginLockDuration)
	}
	store.failures[key] = failure

	return !failure.LockedUntil.IsZero() && failure.LockedUntil.After(now)
}

func (store *memoryLoginLockStore) Clear(_ context.Context, email string) {
	key := normalizeEmail(email)
	if key == "" {
		return
	}

	store.mu.Lock()
	delete(store.failures, key)
	store.mu.Unlock()
}

type redisLoginLockStore struct {
	client *redis.Client
	memory *memoryLoginLockStore
}

func newRedisLoginLockStore(client *redis.Client) LoginLockStore {
	return &redisLoginLockStore{
		client: client,
		memory: newMemoryLoginLockStore(),
	}
}

func (store *redisLoginLockStore) IsLocked(ctx context.Context, email string, now time.Time) bool {
	key := normalizeEmail(email)
	if key == "" {
		return false
	}

	lockKey := loginLockKey(key)
	ttl, err := store.client.TTL(ctx, lockKey).Result()
	if err != nil {
		slog.Warn("redis login lock check failed, falling back to memory", "error", err)
		return store.memory.IsLocked(ctx, email, now)
	}
	if ttl > 0 {
		return true
	}

	return false
}

func (store *redisLoginLockStore) RecordFailure(ctx context.Context, email string, now time.Time) bool {
	key := normalizeEmail(email)
	if key == "" {
		return false
	}

	countKey := loginFailCountKey(key)
	lockKey := loginLockKey(key)

	count, err := store.client.Incr(ctx, countKey).Result()
	if err != nil {
		slog.Warn("redis login failure incr failed, falling back to memory", "error", err)
		return store.memory.RecordFailure(ctx, email, now)
	}

	if count == 1 {
		_ = store.client.Expire(ctx, countKey, loginLockDuration).Err()
	}

	if count >= int64(loginFailureLimit) {
		if err := store.client.Set(ctx, lockKey, "1", loginLockDuration).Err(); err != nil {
			slog.Warn("redis login lock set failed, falling back to memory", "error", err)
			return store.memory.RecordFailure(ctx, email, now)
		}
		_ = store.client.Del(ctx, countKey).Err()
		return true
	}

	return false
}

func (store *redisLoginLockStore) Clear(ctx context.Context, email string) {
	key := normalizeEmail(email)
	if key == "" {
		return
	}

	if err := store.client.Del(ctx, loginFailCountKey(key), loginLockKey(key)).Err(); err != nil {
		slog.Warn("redis login lock clear failed, falling back to memory", "error", err)
		store.memory.Clear(ctx, email)
		return
	}
	store.memory.Clear(ctx, email)
}

func loginFailCountKey(email string) string {
	return fmt.Sprintf("loginfail:%s", email)
}

func loginLockKey(email string) string {
	return fmt.Sprintf("loginlock:%s", email)
}

// NewLoginLockStore returns a Redis-backed store when client is non-nil, otherwise memory.
func NewLoginLockStore(client *redis.Client) LoginLockStore {
	if client == nil {
		return newMemoryLoginLockStore()
	}
	return newRedisLoginLockStore(client)
}
