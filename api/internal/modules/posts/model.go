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
	Category     string    `json:"category"`
	Tags         []string  `json:"tags"`
	CoverImage   string    `json:"coverImage"`
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
	Category   string
	Tags       []string
	CoverImage string
	AuthorName string
}

type Publisher interface {
	Publish(ctx context.Context, input PublishInput) (Post, error)
}

type ViewRecorder interface {
	RecordView(ctx context.Context, slug string) (Post, error)
}
