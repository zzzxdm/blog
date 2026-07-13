package auth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisLoginLockStoreLocksAfterFailures(t *testing.T) {
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	store := NewLoginLockStore(client)
	ctx := context.Background()
	now := time.Now()
	email := "redis.lock@example.com"

	for i := 0; i < loginFailureLimit-1; i++ {
		if store.RecordFailure(ctx, email, now) {
			t.Fatalf("unexpected lock on failure %d", i+1)
		}
	}
	if !store.RecordFailure(ctx, email, now) {
		t.Fatal("expected lock after failure limit")
	}
	if !store.IsLocked(ctx, email, now) {
		t.Fatal("expected redis lock")
	}

	store.Clear(ctx, email)
	if store.IsLocked(ctx, email, now) {
		t.Fatal("expected lock cleared")
	}
}
