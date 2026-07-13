package auth

import (
	"context"
	"testing"
	"time"
)

func TestMemoryLoginLockStoreLocksAfterFailures(t *testing.T) {
	store := newMemoryLoginLockStore()
	ctx := context.Background()
	now := time.Now()
	email := "Lock.User@Example.com"

	for i := 0; i < loginFailureLimit-1; i++ {
		if store.RecordFailure(ctx, email, now) {
			t.Fatalf("unexpected lock on failure %d", i+1)
		}
		if store.IsLocked(ctx, email, now) {
			t.Fatalf("unexpected lock state on failure %d", i+1)
		}
	}

	if !store.RecordFailure(ctx, email, now) {
		t.Fatal("expected lock after reaching failure limit")
	}
	if !store.IsLocked(ctx, email, now) {
		t.Fatal("expected account to be locked")
	}
	if store.IsLocked(ctx, email, now.Add(loginLockDuration+time.Second)) {
		t.Fatal("expected lock to expire")
	}
}

func TestMemoryLoginLockStoreClear(t *testing.T) {
	store := newMemoryLoginLockStore()
	ctx := context.Background()
	now := time.Now()
	email := "clear@example.com"

	for i := 0; i < loginFailureLimit; i++ {
		store.RecordFailure(ctx, email, now)
	}
	if !store.IsLocked(ctx, email, now) {
		t.Fatal("expected lock before clear")
	}

	store.Clear(ctx, email)
	if store.IsLocked(ctx, email, now) {
		t.Fatal("expected lock cleared")
	}
}
