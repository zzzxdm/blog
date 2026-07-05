package messages

import "time"

const (
	TypeReview  = "review"
	TypeComment = "comment"
	TypeSystem  = "system"
	TypeAdmin   = "admin"
	TypeAccount = "account"

	StatusUnread    = "unread"
	StatusRead      = "read"
	StatusArchived  = "archived"
	StatusScheduled = "scheduled"
)

type Message struct {
	ID            string     `json:"id"`
	RecipientID   string     `json:"recipientId"`
	RecipientName string     `json:"recipientName"`
	SenderID      string     `json:"senderId"`
	SenderName    string     `json:"senderName"`
	Type          string     `json:"type"`
	Priority      string     `json:"priority"`
	Title         string     `json:"title"`
	Body          string     `json:"body"`
	TargetType    string     `json:"targetType,omitempty"`
	TargetID      string     `json:"targetId,omitempty"`
	TargetTitle   string     `json:"targetTitle,omitempty"`
	Status        string     `json:"status"`
	ReadAt        *time.Time `json:"readAt,omitempty"`
	ArchivedAt    *time.Time `json:"archivedAt,omitempty"`
	ScheduledAt   *time.Time `json:"scheduledAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type ListQuery struct {
	Status string
	Type   string
}

type Stats struct {
	Unread    int `json:"unread"`
	Review    int `json:"review"`
	Admin     int `json:"admin"`
	Archived  int `json:"archived"`
	Scheduled int `json:"scheduled"`
	Total     int `json:"total"`
}

type ListResult struct {
	Items []Message `json:"items"`
	Total int       `json:"total"`
	Stats Stats     `json:"stats"`
}

type CreateRequest struct {
	RecipientID   string `json:"recipientId"`
	RecipientName string `json:"recipientName"`
	Type          string `json:"type"`
	Priority      string `json:"priority"`
	Title         string `json:"title"`
	Body          string `json:"body"`
	TargetType    string `json:"targetType"`
	TargetID      string `json:"targetId"`
	TargetTitle   string `json:"targetTitle"`
	ScheduledAt   string `json:"scheduledAt"`
}

type BroadcastRecipient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type BroadcastRequest struct {
	Recipients  []BroadcastRecipient `json:"recipients"`
	Type        string               `json:"type"`
	Priority    string               `json:"priority"`
	Title       string               `json:"title"`
	Body        string               `json:"body"`
	TargetType  string               `json:"targetType"`
	TargetID    string               `json:"targetId"`
	TargetTitle string               `json:"targetTitle"`
	ScheduledAt string               `json:"scheduledAt"`
}
