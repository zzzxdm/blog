package posts

import (
	"context"
	"time"
)

type Post struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	Content      string    `json:"content"`
	Visibility   string    `json:"visibility"`
	Category     string    `json:"category"`
	Tags         []string  `json:"tags"`
	CoverImage   string    `json:"coverImage"`
	AuthorID     string    `json:"authorId"`
	AuthorName   string    `json:"authorName"`
	ReadingTime  int       `json:"readingTime"`
	ViewCount    int       `json:"viewCount"`
	LikeCount    int       `json:"likeCount"`
	DislikeCount int       `json:"dislikeCount"`
	CommentCount int       `json:"commentCount"`
	PublishedAt  time.Time `json:"publishedAt"`
}

type ListQuery struct {
	Keyword  string
	Category string
	Tag      string
	Author   string
	Sort     string
	Page     int
	PageSize int
}

type ListResult struct {
	Items    []Post `json:"items"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Total    int    `json:"total"`
}

type SiteStats struct {
	PostCount int `json:"postCount"`
	ViewCount int `json:"viewCount"`
	WordCount int `json:"wordCount"`
}

type PublishInput struct {
	Slug       string
	Title      string
	Summary    string
	Content    string
	Visibility string
	Category   string
	Tags       []string
	CoverImage string
	AuthorID   string
	AuthorName string
}

type Publisher interface {
	Publish(ctx context.Context, input PublishInput) (Post, error)
}

type SubmissionPublisher interface {
	PublishSubmission(ctx context.Context, input PublishInput, existingSlug string) (Post, error)
}

type AdminPublisher interface {
	PublishAdmin(ctx context.Context, input PublishInput, existingSlug string) (Post, error)
}

type Archiver interface {
	Archive(ctx context.Context, slug string) error
}

type Restorer interface {
	Restore(ctx context.Context, slug string) error
}

type ViewRecorder interface {
	RecordView(ctx context.Context, slug string) (Post, error)
}

type RestrictedViewRecorder interface {
	RecordRestrictedView(ctx context.Context, slug string, viewer Viewer) (Post, error)
}

type RestrictedGetter interface {
	GetBySlugForViewer(ctx context.Context, slug string, viewer Viewer) (Post, error)
}

type PrivateLister interface {
	ListPrivate(ctx context.Context, viewer Viewer, query ListQuery) (ListResult, error)
}

type Viewer struct {
	ID   string
	Role string
}
