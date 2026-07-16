package cachex

import (
	"context"
	"testing"
	"time"
)

func TestMemoryViewDeduperAllowsOncePerDay(t *testing.T) {
	d := newMemoryViewDeduper()
	ctx := context.Background()

	if !d.Allow(ctx, "ip:1.2.3.4", "hello-world") {
		t.Fatal("first view should count")
	}
	if d.Allow(ctx, "ip:1.2.3.4", "hello-world") {
		t.Fatal("second view same day should be suppressed")
	}
	if !d.Allow(ctx, "ip:1.2.3.4", "other-post") {
		t.Fatal("different slug should count")
	}
	if !d.Allow(ctx, "ip:9.9.9.9", "hello-world") {
		t.Fatal("different visitor should count")
	}
}

func TestVisitorKeyPrefersSession(t *testing.T) {
	key := VisitorKey("sess-token", "1.1.1.1")
	if key != "s:sess-token" {
		t.Fatalf("unexpected visitor key: %s", key)
	}
	key = VisitorKey("", "1.1.1.1")
	if key != "ip:1.1.1.1" {
		t.Fatalf("unexpected ip visitor key: %s", key)
	}
}

func TestMemoryTTLStoreExpires(t *testing.T) {
	store := newMemoryTTLStore()
	ctx := context.Background()
	store.Set(ctx, "k", `{"ok":true}`, 20*time.Millisecond)
	if _, ok := store.Get(ctx, "k"); !ok {
		t.Fatal("expected cache hit")
	}
	time.Sleep(30 * time.Millisecond)
	if _, ok := store.Get(ctx, "k"); ok {
		t.Fatal("expected cache miss after ttl")
	}
}
