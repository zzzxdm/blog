package reactions

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrInvalidReaction = errors.New("invalid reaction")

type Repository interface {
	Get(ctx context.Context, postSlug string, userID string) (Summary, error)
	SetReaction(ctx context.Context, postSlug string, userID string, reaction string) (Summary, error)
	SetBookmark(ctx context.Context, postSlug string, userID string, bookmarked bool) (Summary, error)
	ListBookmarks(ctx context.Context, userID string) ([]Bookmark, error)
}

type postState struct {
	LikeCount     int
	DislikeCount  int
	BookmarkCount int
	Reactions     map[string]string
	Bookmarks     map[string]time.Time
}

type MemoryRepository struct {
	mu    sync.RWMutex
	state map[string]*postState
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		state: map[string]*postState{
			"blog-system-design": {
				LikeCount:     128,
				DislikeCount:  7,
				BookmarkCount: 34,
				Reactions:     map[string]string{"user_linyi": "like"},
				Bookmarks:     map[string]time.Time{"user_linyi": time.Now().Add(-2 * time.Hour)},
			},
			"vue3-content-site-cache-seo": {
				LikeCount:     96,
				DislikeCount:  3,
				BookmarkCount: 18,
				Reactions:     map[string]string{},
				Bookmarks:     map[string]time.Time{"user_linyi": time.Now().Add(-26 * time.Hour)},
			},
			"postgres-redis-blog-boundary": {
				LikeCount:     84,
				DislikeCount:  4,
				BookmarkCount: 25,
				Reactions:     map[string]string{},
				Bookmarks:     map[string]time.Time{},
			},
		},
	}
}

func (repo *MemoryRepository) Get(_ context.Context, postSlug string, userID string) (Summary, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	return repo.summary(postSlug, userID), nil
}

func (repo *MemoryRepository) SetReaction(_ context.Context, postSlug string, userID string, reaction string) (Summary, error) {
	reaction = strings.ToLower(strings.TrimSpace(reaction))
	if reaction != "" && reaction != "like" && reaction != "dislike" {
		return Summary{}, ErrInvalidReaction
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	state := repo.ensure(postSlug)
	previous := state.Reactions[userID]

	if previous == reaction {
		reaction = ""
	}

	if previous == "like" {
		state.LikeCount--
	}
	if previous == "dislike" {
		state.DislikeCount--
	}

	if reaction == "" {
		delete(state.Reactions, userID)
		return repo.summaryLocked(postSlug, userID, state), nil
	}

	state.Reactions[userID] = reaction
	if reaction == "like" {
		state.LikeCount++
	}
	if reaction == "dislike" {
		state.DislikeCount++
	}

	return repo.summaryLocked(postSlug, userID, state), nil
}

func (repo *MemoryRepository) SetBookmark(_ context.Context, postSlug string, userID string, bookmarked bool) (Summary, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	state := repo.ensure(postSlug)
	_, previous := state.Bookmarks[userID]

	if bookmarked && !previous {
		state.BookmarkCount++
		state.Bookmarks[userID] = time.Now()
	}
	if !bookmarked && previous {
		state.BookmarkCount--
		delete(state.Bookmarks, userID)
	}

	return repo.summaryLocked(postSlug, userID, state), nil
}

func (repo *MemoryRepository) ListBookmarks(_ context.Context, userID string) ([]Bookmark, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	items := make([]Bookmark, 0)
	for postSlug, state := range repo.state {
		bookmarkedAt, ok := state.Bookmarks[userID]
		if !ok {
			continue
		}

		items = append(items, Bookmark{
			PostSlug:     postSlug,
			BookmarkedAt: bookmarkedAt,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].BookmarkedAt.After(items[j].BookmarkedAt)
	})

	return items, nil
}

func (repo *MemoryRepository) ensure(postSlug string) *postState {
	state, ok := repo.state[postSlug]
	if ok {
		return state
	}

	state = &postState{
		Reactions: map[string]string{},
		Bookmarks: map[string]time.Time{},
	}
	repo.state[postSlug] = state
	return state
}

func (repo *MemoryRepository) summary(postSlug string, userID string) Summary {
	state, ok := repo.state[postSlug]
	if !ok {
		return Summary{PostSlug: postSlug}
	}

	return repo.summaryLocked(postSlug, userID, state)
}

func (repo *MemoryRepository) summaryLocked(postSlug string, userID string, state *postState) Summary {
	return Summary{
		PostSlug:      postSlug,
		LikeCount:     maxInt(0, state.LikeCount),
		DislikeCount:  maxInt(0, state.DislikeCount),
		BookmarkCount: maxInt(0, state.BookmarkCount),
		MyReaction:    state.Reactions[userID],
		Bookmarked:    !state.Bookmarks[userID].IsZero(),
	}
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
