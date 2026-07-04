package adminposts

import (
	"context"
	"errors"
	"testing"
	"time"

	"blog/api/internal/modules/posts"
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

func TestMemoryRepositorySaveScheduledRequiresPublicVisibility(t *testing.T) {
	repo := &MemoryRepository{
		items:  map[string]AdminPost{},
		nextID: 1,
		now: func() time.Time {
			return time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)
		},
	}

	_, err := repo.Save(context.Background(), "", SaveRequest{
		Title:       "定时发布文章",
		Content:     "需要公开后才能定时发布。",
		Status:      StatusScheduled,
		Visibility:  VisibilityPrivate,
		ScheduledAt: time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
	})
	if !errors.Is(err, ErrPostNotPublic) {
		t.Fatalf("Save private scheduled returned %v, want %v", err, ErrPostNotPublic)
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

func TestMemoryRepositoryPublishDuePublishesOnlyDuePublicScheduledPosts(t *testing.T) {
	now := time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)
	due := now.Add(-time.Minute)
	future := now.Add(time.Minute)
	publisher := &recordingPublisher{}
	repo := &MemoryRepository{
		items: map[string]AdminPost{
			"due": {
				ID:          "due",
				Slug:        "due-post",
				Title:       "到期文章",
				Content:     "到期后应该发布。",
				Status:      StatusScheduled,
				Visibility:  VisibilityPublic,
				ScheduledAt: &due,
				AuthorName:  "管理员",
				UpdatedAt:   now,
			},
			"future": {
				ID:          "future",
				Slug:        "future-post",
				Title:       "未来文章",
				Content:     "还没到发布时间。",
				Status:      StatusScheduled,
				Visibility:  VisibilityPublic,
				ScheduledAt: &future,
				AuthorName:  "管理员",
				UpdatedAt:   now,
			},
			"private": {
				ID:          "private",
				Slug:        "private-post",
				Title:       "私密文章",
				Content:     "私密文章不能发布到公开站点。",
				Status:      StatusScheduled,
				Visibility:  VisibilityPrivate,
				ScheduledAt: &due,
				AuthorName:  "管理员",
				UpdatedAt:   now,
			},
		},
		now: func() time.Time {
			return now
		},
	}

	count, err := repo.PublishDue(context.Background(), publisher, now)
	if err != nil {
		t.Fatalf("PublishDue returned error: %v", err)
	}
	if count != 1 || len(publisher.inputs) != 1 {
		t.Fatalf("PublishDue published %d items and recorded %d inputs, want 1", count, len(publisher.inputs))
	}

	published, err := repo.Get(context.Background(), "due")
	if err != nil {
		t.Fatalf("Get due returned error: %v", err)
	}
	if published.Status != StatusPublished || published.PublishedPostSlug != "due-post" {
		t.Fatalf("due post = %+v, want published with slug due-post", published)
	}

	futurePost, _ := repo.Get(context.Background(), "future")
	privatePost, _ := repo.Get(context.Background(), "private")
	if futurePost.Status != StatusScheduled || privatePost.Status != StatusScheduled {
		t.Fatalf("future/private posts should remain scheduled, got future=%s private=%s", futurePost.Status, privatePost.Status)
	}
}

func TestCountStatsIncludesScheduled(t *testing.T) {
	now := time.Date(2026, time.July, 15, 10, 0, 0, 0, time.UTC)
	lastMonth := now.AddDate(0, -1, 0)
	stats := countStatsAt([]AdminPost{
		{Status: StatusPublished, ViewCount: 16800, PublishedAt: &now},
		{Status: StatusPublished, ViewCount: 5000, PublishedAt: &lastMonth},
		{Status: StatusDraft},
		{Status: StatusReview},
		{Status: StatusScheduled},
	}, now)

	if stats.Published != 2 || stats.Draft != 1 || stats.Review != 1 || stats.Scheduled != 1 || stats.Total != 5 {
		t.Fatalf("countStats = %+v", stats)
	}
	if stats.MonthlyViews != "16.8k" {
		t.Fatalf("MonthlyViews = %q, want 16.8k", stats.MonthlyViews)
	}
}

type recordingPublisher struct {
	inputs []posts.PublishInput
}

func (publisher *recordingPublisher) Publish(_ context.Context, input posts.PublishInput) (posts.Post, error) {
	publisher.inputs = append(publisher.inputs, input)
	return posts.Post{
		Slug:         input.Slug,
		ViewCount:    12,
		CommentCount: 3,
	}, nil
}
