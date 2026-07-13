package redisx

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Connect opens a Redis client and pings it. Returns nil when addr is empty or Redis is unreachable
// so callers can fall back to in-memory implementations.
func Connect(addr string, password string) *redis.Client {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		slog.Warn("redis unavailable, using in-memory fallbacks", "addr", addr, "error", err)
		_ = client.Close()
		return nil
	}

	slog.Info("redis connected", "addr", addr)
	return client
}
