package reactions

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryRepositorySetReactionSwitchesAndToggles(t *testing.T) {
	repo := &MemoryRepository{state: map[string]*postState{}}

	summary, err := repo.SetReaction(context.Background(), "post-a", "user-a", "like")
	if err != nil {
		t.Fatalf("SetReaction like returned error: %v", err)
	}
	if summary.LikeCount != 1 || summary.DislikeCount != 0 || summary.MyReaction != "like" {
		t.Fatalf("after like summary = %+v", summary)
	}

	summary, err = repo.SetReaction(context.Background(), "post-a", "user-a", "dislike")
	if err != nil {
		t.Fatalf("SetReaction dislike returned error: %v", err)
	}
	if summary.LikeCount != 0 || summary.DislikeCount != 1 || summary.MyReaction != "dislike" {
		t.Fatalf("after switch summary = %+v", summary)
	}

	summary, err = repo.SetReaction(context.Background(), "post-a", "user-a", "dislike")
	if err != nil {
		t.Fatalf("SetReaction toggle returned error: %v", err)
	}
	if summary.LikeCount != 0 || summary.DislikeCount != 0 || summary.MyReaction != "" {
		t.Fatalf("after toggle summary = %+v", summary)
	}

	_, err = repo.SetReaction(context.Background(), "post-a", "user-a", "star")
	if !errors.Is(err, ErrInvalidReaction) {
		t.Fatalf("SetReaction invalid returned %v, want %v", err, ErrInvalidReaction)
	}
}

func TestMemoryRepositorySetBookmarkIsIdempotent(t *testing.T) {
	repo := &MemoryRepository{state: map[string]*postState{}}

	summary, err := repo.SetBookmark(context.Background(), "post-a", "user-a", true)
	if err != nil {
		t.Fatalf("SetBookmark true returned error: %v", err)
	}
	if summary.BookmarkCount != 1 || !summary.Bookmarked {
		t.Fatalf("after bookmark summary = %+v", summary)
	}

	summary, err = repo.SetBookmark(context.Background(), "post-a", "user-a", true)
	if err != nil {
		t.Fatalf("SetBookmark true twice returned error: %v", err)
	}
	if summary.BookmarkCount != 1 || !summary.Bookmarked {
		t.Fatalf("after duplicate bookmark summary = %+v", summary)
	}

	summary, err = repo.SetBookmark(context.Background(), "post-a", "user-a", false)
	if err != nil {
		t.Fatalf("SetBookmark false returned error: %v", err)
	}
	if summary.BookmarkCount != 0 || summary.Bookmarked {
		t.Fatalf("after unbookmark summary = %+v", summary)
	}
}

func TestMemoryRepositoryListBookmarksSortedNewestFirst(t *testing.T) {
	oldTime := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(time.Hour)
	repo := &MemoryRepository{
		state: map[string]*postState{
			"old": {
				Bookmarks: map[string]time.Time{"user-a": oldTime},
			},
			"new": {
				Bookmarks: map[string]time.Time{"user-a": newTime},
			},
			"other": {
				Bookmarks: map[string]time.Time{"user-b": newTime},
			},
		},
	}

	bookmarks, err := repo.ListBookmarks(context.Background(), "user-a")
	if err != nil {
		t.Fatalf("ListBookmarks returned error: %v", err)
	}
	if len(bookmarks) != 2 || bookmarks[0].PostSlug != "new" || bookmarks[1].PostSlug != "old" {
		t.Fatalf("ListBookmarks = %+v, want newest user-a bookmarks first", bookmarks)
	}
}

func TestReactionDeltas(t *testing.T) {
	cases := []struct {
		name         string
		previous     string
		next         string
		likeDelta    int
		dislikeDelta int
	}{
		{name: "new like", next: "like", likeDelta: 1},
		{name: "like to dislike", previous: "like", next: "dislike", likeDelta: -1, dislikeDelta: 1},
		{name: "dislike to empty", previous: "dislike", dislikeDelta: -1},
		{name: "empty", previous: "", next: "", likeDelta: 0, dislikeDelta: 0},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			likeDelta, dislikeDelta := reactionDeltas(tt.previous, tt.next)
			if likeDelta != tt.likeDelta || dislikeDelta != tt.dislikeDelta {
				t.Fatalf("reactionDeltas(%q, %q) = (%d, %d), want (%d, %d)", tt.previous, tt.next, likeDelta, dislikeDelta, tt.likeDelta, tt.dislikeDelta)
			}
		})
	}
}
