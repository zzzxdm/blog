package comments

import "time"

type Comment struct {
	ID         string    `json:"id"`
	PostSlug   string    `json:"postSlug"`
	PostTitle  string    `json:"postTitle,omitempty"`
	ParentID   string    `json:"parentId,omitempty"`
	AuthorID   string    `json:"authorId"`
	AuthorName string    `json:"authorName"`
	AvatarText string    `json:"avatarText"`
	Body       string    `json:"body"`
	Status     string    `json:"status"`
	LikeCount  int       `json:"likeCount"`
	ReplyCount int       `json:"replyCount"`
	RiskLevel  string    `json:"riskLevel,omitempty"`
	IsMine     bool      `json:"isMine"`
	IsAuthor   bool      `json:"isAuthor"`
	Liked      bool      `json:"liked"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ListResult struct {
	Items []Comment `json:"items"`
	Total int       `json:"total"`
}

type CreateRequest struct {
	Body     string `json:"body"`
	ParentID string `json:"parentId"`
}

type ListQuery struct {
	Status   string
	Keyword  string
	Sort     string
	Page     int
	PageSize int
	All      bool
}

type ManageStats struct {
	Total    int `json:"total"`
	Pending  int `json:"pending"`
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
	Spam     int `json:"spam"`
	Deleted  int `json:"deleted"`
	Likes    int `json:"likes"`
	Replies  int `json:"replies"`
}

type ManageListResult struct {
	Items    []Comment   `json:"items"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Total    int         `json:"total"`
	Stats    ManageStats `json:"stats"`
}

type StatusRequest struct {
	Status string `json:"status"`
}

type ReportRequest struct {
	Reason string `json:"reason"`
}

type CommentReport struct {
	ID         string    `json:"id"`
	CommentID  string    `json:"commentId"`
	ReporterID string    `json:"reporterId"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ReportListResult struct {
	Items []CommentReport `json:"items"`
	Total int             `json:"total"`
}
