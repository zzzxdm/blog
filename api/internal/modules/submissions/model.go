package submissions

import "time"

const (
	StatusDraft     = "draft"
	StatusSubmitted = "submitted"
	StatusReturned  = "returned"
	StatusRejected  = "rejected"
	StatusPublished = "published"
	StatusArchived  = "archived"

	VisibilityPublic  = "public"
	VisibilityPrivate = "private"

	ActionApprove = "approve"
	ActionReturn  = "return"
	ActionReject  = "reject"
)

type Submission struct {
	ID                string     `json:"id"`
	AuthorID          string     `json:"authorId"`
	AuthorName        string     `json:"authorName"`
	AuthorAvatar      string     `json:"authorAvatar"`
	Title             string     `json:"title"`
	Summary           string     `json:"summary"`
	Content           string     `json:"content"`
	Category          string     `json:"category"`
	Tags              []string   `json:"tags"`
	CoverImage        string     `json:"coverImage"`
	Slug              string     `json:"slug"`
	Visibility        string     `json:"visibility"`
	Status            string     `json:"status"`
	ReviewNote        string     `json:"reviewNote"`
	ReviewerID        string     `json:"reviewerId,omitempty"`
	ReviewerName      string     `json:"reviewerName,omitempty"`
	PublishedPostSlug string     `json:"publishedPostSlug,omitempty"`
	WordCount         int        `json:"wordCount"`
	Version           int        `json:"version"`
	RiskLevel         string     `json:"riskLevel"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	SubmittedAt       *time.Time `json:"submittedAt,omitempty"`
	ReviewedAt        *time.Time `json:"reviewedAt,omitempty"`
	PublishedAt       *time.Time `json:"publishedAt,omitempty"`
}

type ListQuery struct {
	Status   string
	Keyword  string
	Sort     string
	Page     int
	PageSize int
	All      bool
}

type Stats struct {
	Draft     int `json:"draft"`
	Submitted int `json:"submitted"`
	Returned  int `json:"returned"`
	Rejected  int `json:"rejected"`
	Published int `json:"published"`
	Archived  int `json:"archived"`
	Total     int `json:"total"`
}

type ListResult struct {
	Items    []Submission `json:"items"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
	Total    int          `json:"total"`
	Stats    Stats        `json:"stats"`
}

type SaveRequest struct {
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	Content        string   `json:"content"`
	Category       string   `json:"category"`
	Tags           []string `json:"tags"`
	CoverImage     string   `json:"coverImage"`
	Slug           string   `json:"slug"`
	Visibility     string   `json:"visibility"`
	Submit         bool     `json:"submit"`
	TurnstileToken string   `json:"turnstileToken"`
}

type ReviewRequest struct {
	Action   string `json:"action"`
	Note     string `json:"note"`
	Slug     string `json:"slug"`
	Category string `json:"category"`
}
