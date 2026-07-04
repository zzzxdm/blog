package comments

import "time"

type Comment struct {
	ID         string    `json:"id"`
	PostSlug   string    `json:"postSlug"`
	ParentID   string    `json:"parentId,omitempty"`
	AuthorID   string    `json:"authorId"`
	AuthorName string    `json:"authorName"`
	AvatarText string    `json:"avatarText"`
	Body       string    `json:"body"`
	Status     string    `json:"status"`
	LikeCount  int       `json:"likeCount"`
	IsMine     bool      `json:"isMine"`
	IsAuthor   bool      `json:"isAuthor"`
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
