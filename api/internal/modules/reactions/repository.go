package reactions

import (
	"context"
	"errors"
)

var ErrInvalidReaction = errors.New("invalid reaction")

type Repository interface {
	Get(ctx context.Context, postSlug string, userID string) (Summary, error)
	SetReaction(ctx context.Context, postSlug string, userID string, reaction string) (Summary, error)
	SetBookmark(ctx context.Context, postSlug string, userID string, bookmarked bool) (Summary, error)
	ListBookmarks(ctx context.Context, userID string, query BookmarkQuery) (BookmarkPage, error)
}
