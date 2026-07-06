package topics

import "time"

type Topic struct {
	ID           string     `json:"id"`
	Slug         string     `json:"slug"`
	Title        string     `json:"title"`
	Summary      string     `json:"summary"`
	CoverImage   string     `json:"coverImage"`
	ImageAlt     string     `json:"imageAlt"`
	Tone         string     `json:"tone"`
	Status       string     `json:"status"`
	Featured     bool       `json:"featured"`
	SortOrder    int        `json:"sortOrder"`
	Categories   []string   `json:"categories"`
	Tags         []string   `json:"tags"`
	PostCount    int        `json:"postCount"`
	LatestPostAt *time.Time `json:"latestPostAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type SaveRequest struct {
	Slug       string   `json:"slug"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	CoverImage string   `json:"coverImage"`
	ImageAlt   string   `json:"imageAlt"`
	Tone       string   `json:"tone"`
	Status     string   `json:"status"`
	Featured   bool     `json:"featured"`
	SortOrder  int      `json:"sortOrder"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
}

type ListQuery struct {
	Keyword  string
	Status   string
	Featured bool
	All      bool
	Page     int
	PageSize int
}

type ListResult struct {
	Items    []Topic `json:"items"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
	Total    int     `json:"total"`
}
