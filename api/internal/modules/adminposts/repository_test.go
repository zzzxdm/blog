package adminposts

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestParseScheduledAt(t *testing.T) {
	want := time.Date(2026, 7, 5, 10, 30, 0, 0, time.FixedZone("CST", 8*60*60))

	got, err := parseScheduledAt(" 2026-07-05T10:30:00+08:00 ")
	if err != nil {
		t.Fatalf("parseScheduledAt returned error: %v", err)
	}
	if got == nil || !got.Equal(want) {
		t.Fatalf("parseScheduledAt returned %v, want %v", got, want)
	}

	empty, err := parseScheduledAt(" ")
	if err != nil {
		t.Fatalf("parseScheduledAt empty returned error: %v", err)
	}
	if empty != nil {
		t.Fatalf("parseScheduledAt empty returned %v, want nil", empty)
	}

	_, err = parseScheduledAt("2026-07-05 10:30")
	if !errors.Is(err, ErrInvalidPost) {
		t.Fatalf("parseScheduledAt invalid returned %v, want %v", err, ErrInvalidPost)
	}
}

func TestMemoryRepositorySaveScheduledRequiresScheduledAt(t *testing.T) {
	repo := &MemoryRepository{
		items:  map[string]AdminPost{},
		nextID: 1,
		now: func() time.Time {
			return time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)
		},
	}

	_, err := repo.Save(context.Background(), "", SaveRequest{
		Title:   "定时发布文章",
		Content: "需要选择发布时间。",
		Status:  StatusScheduled,
	})
	if !errors.Is(err, ErrInvalidPost) {
		t.Fatalf("Save scheduled without scheduledAt returned %v, want %v", err, ErrInvalidPost)
	}
}

func TestMemoryRepositorySaveScheduledStoresScheduledAtAndStats(t *testing.T) {
	now := time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)
	scheduledAt := now.Add(2 * time.Hour)
	repo := &MemoryRepository{
		items:  map[string]AdminPost{},
		nextID: 1,
		now: func() time.Time {
			return now
		},
	}

	post, err := repo.Save(context.Background(), "", SaveRequest{
		Title:       "定时发布文章",
		Content:     "到点之后再发布。",
		Status:      StatusScheduled,
		ScheduledAt: scheduledAt.Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Save scheduled returned error: %v", err)
	}
	if post.ScheduledAt == nil || !post.ScheduledAt.Equal(scheduledAt) {
		t.Fatalf("ScheduledAt = %v, want %v", post.ScheduledAt, scheduledAt)
	}

	result, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if result.Stats.Scheduled != 1 || result.Stats.Total != 1 {
		t.Fatalf("Stats = %+v, want one scheduled post", result.Stats)
	}
}

func TestCountStatsIncludesScheduled(t *testing.T) {
	stats := countStats([]AdminPost{
		{Status: StatusPublished},
		{Status: StatusDraft},
		{Status: StatusReview},
		{Status: StatusScheduled},
	})

	if stats.Published != 1 || stats.Draft != 1 || stats.Review != 1 || stats.Scheduled != 1 || stats.Total != 4 {
		t.Fatalf("countStats = %+v", stats)
	}
}
