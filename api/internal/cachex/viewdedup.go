package cachex

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// ViewDeduper decides whether a post view should increment counters.
// Implementations are best-effort: on storage errors they typically allow counting.
type ViewDeduper interface {
	// Allow returns true once per visitor+slug within the current UTC day.
	Allow(ctx context.Context, visitorKey string, slug string) bool
}

type memoryViewDeduper struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

// NewViewDeduper returns a Redis-backed deduper when client is non-nil, otherwise memory.
func NewViewDeduper(client *redis.Client) ViewDeduper {
	if client != nil {
		return &redisViewDeduper{client: client, memory: newMemoryViewDeduper()}
	}
	return newMemoryViewDeduper()
}

func newMemoryViewDeduper() *memoryViewDeduper {
	return &memoryViewDeduper{seen: map[string]time.Time{}}
}

func (d *memoryViewDeduper) Allow(_ context.Context, visitorKey string, slug string) bool {
	key := viewKey(visitorKey, slug, time.Now().UTC())
	if key == "" {
		return true
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	// opportunistic cleanup of expired entries
	for k, exp := range d.seen {
		if exp.Before(now) {
			delete(d.seen, k)
		}
	}

	if exp, ok := d.seen[key]; ok && exp.After(now) {
		return false
	}
	// expire at next UTC midnight + small buffer
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, time.UTC)
	d.seen[key] = tomorrow
	return true
}

type redisViewDeduper struct {
	client *redis.Client
	memory *memoryViewDeduper
}

func (d *redisViewDeduper) Allow(ctx context.Context, visitorKey string, slug string) bool {
	key := viewKey(visitorKey, slug, time.Now().UTC())
	if key == "" {
		return true
	}

	ttl := remainingDayTTL(time.Now().UTC())
	ok, err := d.client.SetNX(ctx, "viewdedup:"+key, "1", ttl).Result()
	if err != nil {
		return d.memory.Allow(ctx, visitorKey, slug)
	}
	return ok
}

func viewKey(visitorKey string, slug string, now time.Time) string {
	visitorKey = trimViewToken(visitorKey)
	slug = trimViewToken(slug)
	if visitorKey == "" || slug == "" {
		return ""
	}
	day := now.UTC().Format("2006-01-02")
	sum := sha1.Sum([]byte(visitorKey + "|" + slug + "|" + day))
	return hex.EncodeToString(sum[:])
}

func trimViewToken(value string) string {
	for len(value) > 0 && (value[0] == ' ' || value[0] == '\t') {
		value = value[1:]
	}
	for len(value) > 0 && (value[len(value)-1] == ' ' || value[len(value)-1] == '\t') {
		value = value[:len(value)-1]
	}
	return value
}

func remainingDayTTL(now time.Time) time.Duration {
	now = now.UTC()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, time.UTC)
	ttl := tomorrow.Sub(now)
	if ttl < time.Minute {
		return time.Minute
	}
	return ttl
}

// VisitorKey builds a stable visitor identity from session cookie or client IP.
func VisitorKey(sessionToken string, clientIP string) string {
	if token := trimViewToken(sessionToken); token != "" {
		return "s:" + token
	}
	if ip := trimViewToken(clientIP); ip != "" {
		return "ip:" + ip
	}
	return fmt.Sprintf("anon:%d", time.Now().UnixNano())
}
