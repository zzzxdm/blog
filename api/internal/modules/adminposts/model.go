package adminposts

import "time"

const (
	StatusDraft     = "draft"
	StatusReview    = "review"
	StatusScheduled = "scheduled"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
	VisibilityMembers = "members"
)

type AdminPost struct {
	ID                string     `json:"id"`
	Slug              string     `json:"slug"`
	Title             string     `json:"title"`
	Summary           string     `json:"summary"`
	Content           string     `json:"content"`
	Status            string     `json:"status"`
	Visibility        string     `json:"visibility"`
	Category          string     `json:"category"`
	Tags              []string   `json:"tags"`
	CoverImage        string     `json:"coverImage"`
	AuthorID          string     `json:"authorId,omitempty"`
	AuthorName        string     `json:"authorName"`
	ReadingTime       int        `json:"readingTime"`
	ViewCount         int        `json:"viewCount"`
	CommentCount      int        `json:"commentCount"`
	SEOtitle          string     `json:"seoTitle"`
	SEODescription    string     `json:"seoDescription"`
	Version           int        `json:"version"`
	Revisions         []Revision `json:"revisions,omitempty"`
	PublishedPostSlug string     `json:"publishedPostSlug,omitempty"`
	ScheduledAt       *time.Time `json:"scheduledAt,omitempty"`
	PublishedAt       *time.Time `json:"publishedAt,omitempty"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type Revision struct {
	ID             string    `json:"id"`
	Version        int       `json:"version"`
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Summary        string    `json:"summary"`
	Content        string    `json:"content"`
	Status         string    `json:"status"`
	Visibility     string    `json:"visibility"`
	Category       string    `json:"category"`
	Tags           []string  `json:"tags"`
	CoverImage     string    `json:"coverImage"`
	AuthorID       string    `json:"authorId,omitempty"`
	SEOtitle       string    `json:"seoTitle"`
	SEODescription string    `json:"seoDescription"`
	AuthorName     string    `json:"authorName"`
	CreatedAt      time.Time `json:"createdAt"`
}

type RevisionListResult struct {
	Items []Revision `json:"items"`
	Total int        `json:"total"`
}

type PreviewResult struct {
	PreviewURL string    `json:"previewUrl"`
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type Stats struct {
	Published    int    `json:"published"`
	Draft        int    `json:"draft"`
	Review       int    `json:"review"`
	Scheduled    int    `json:"scheduled"`
	MonthlyViews string `json:"monthlyViews"`
	Total        int    `json:"total"`
}

type ListResult struct {
	Items    []AdminPost `json:"items"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Total    int         `json:"total"`
	Stats    Stats       `json:"stats"`
}

type ListQuery struct {
	Keyword  string
	Status   string
	Sort     string
	Page     int
	PageSize int
	All      bool
}

type SaveRequest struct {
	Slug           string   `json:"slug"`
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	Content        string   `json:"content"`
	Status         string   `json:"status"`
	Visibility     string   `json:"visibility"`
	ScheduledAt    string   `json:"scheduledAt"`
	Category       string   `json:"category"`
	Tags           []string `json:"tags"`
	CoverImage     string   `json:"coverImage"`
	SEOtitle       string   `json:"seoTitle"`
	SEODescription string   `json:"seoDescription"`
}
