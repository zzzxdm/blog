package messages

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrInvalidMessage  = errors.New("invalid message")
)

type Repository interface {
	List(ctx context.Context, userID string, query ListQuery) (ListResult, error)
	AdminList(ctx context.Context, query ListQuery) (ListResult, error)
	Create(ctx context.Context, request CreateRequest, sender auth.User) (Message, error)
	MarkRead(ctx context.Context, userID string, messageID string) (Message, error)
	MarkAllRead(ctx context.Context, userID string) (Stats, error)
	Archive(ctx context.Context, userID string, messageID string) (Message, error)
}

func withStatus(message Message, now time.Time) Message {
	switch {
	case message.ScheduledAt != nil && message.ScheduledAt.After(now):
		message.Status = StatusScheduled
	case message.ArchivedAt != nil:
		message.Status = StatusArchived
	case message.ReadAt != nil:
		message.Status = StatusRead
	default:
		message.Status = StatusUnread
	}

	return message
}

func matchesQuery(message Message, query ListQuery) bool {
	status := strings.ToLower(strings.TrimSpace(query.Status))
	messageType := strings.ToLower(strings.TrimSpace(query.Type))
	keyword := strings.ToLower(strings.TrimSpace(query.Keyword))

	if status == "sent" {
		if message.Status == StatusScheduled || message.Status == StatusArchived {
			return false
		}
	} else if status != "" && status != "all" && message.Status != status {
		return false
	}

	if messageType != "" && messageType != "all" && message.Type != messageType {
		return false
	}

	if status == "" && message.Status == StatusArchived {
		return false
	}

	if keyword != "" && !messageContainsKeyword(message, keyword) {
		return false
	}

	return true
}

func messageContainsKeyword(message Message, keyword string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		message.Title,
		message.Body,
		message.RecipientName,
		message.RecipientID,
		message.TargetTitle,
		message.Type,
		message.Status,
	}, " "))
	return strings.Contains(haystack, keyword)
}

func parseScheduledAt(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	layouts := []string{time.RFC3339, "2006-01-02T15:04"}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed, nil
		}
	}

	return nil, ErrInvalidMessage
}

func sortMessages(items []Message) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func pagedMessageResult(items []Message, stats Stats, query ListQuery) ListResult {
	total := len(items)
	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	paged := items
	if query.All {
		page = 1
		pageSize = total
	} else {
		start := (page - 1) * pageSize
		if start > total {
			start = total
		}
		end := start + pageSize
		if end > total {
			end = total
		}
		paged = items[start:end]
	}

	return ListResult{
		Items:    paged,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Stats:    stats,
	}
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return 10
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

func normalizeType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case TypeReview:
		return TypeReview
	case TypeComment:
		return TypeComment
	case TypeSystem:
		return TypeSystem
	case TypeAccount:
		return TypeAccount
	default:
		return TypeAdmin
	}
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func seedMessages() []Message {
	now := time.Now()
	readAt := now.Add(-22 * time.Hour)
	archivedAt := now.Add(-72 * time.Hour)

	return []Message{
		{
			ID:            "8001",
			RecipientID:   "5001",
			RecipientName: "林一",
			SenderID:      "5002",
			SenderName:    "管理员",
			Type:          TypeReview,
			Priority:      "normal",
			Title:         "你的投稿已退回修改",
			Body:          "《如何写一篇可维护的技术文章》需要补充摘要和代码示例。修改后可以重新提交。",
			TargetType:    "submission",
			TargetID:      "7003",
			TargetTitle:   "如何写一篇可维护的技术文章",
			CreatedAt:     now.Add(-20 * time.Minute),
		},
		{
			ID:            "8002",
			RecipientID:   "5001",
			RecipientName: "林一",
			SenderID:      "5002",
			SenderName:    "管理员",
			Type:          TypeComment,
			Priority:      "normal",
			Title:         "管理员回复了你的评论",
			Body:          "审核结果会同步到站内信和我的投稿列表。",
			TargetType:    "comment",
			TargetID:      "6002",
			TargetTitle:   "如何设计一个内容长期增长的博客系统",
			CreatedAt:     now.Add(-2 * time.Hour),
		},
		{
			ID:            "8003",
			RecipientID:   "5001",
			RecipientName: "林一",
			SenderID:      "system",
			SenderName:    "系统",
			Type:          TypeSystem,
			Priority:      "important",
			Title:         "本周将维护媒体库上传服务",
			Body:          "维护期间图片上传会短暂不可用，阅读不受影响。",
			CreatedAt:     now.Add(-6 * time.Hour),
		},
		{
			ID:            "8004",
			RecipientID:   "5001",
			RecipientName: "林一",
			SenderID:      "system",
			SenderName:    "系统",
			Type:          TypeReview,
			Priority:      "normal",
			Title:         "你的评论已通过审核",
			Body:          "评论现在已展示在文章详情页。",
			ReadAt:        &readAt,
			CreatedAt:     now.Add(-24 * time.Hour),
		},
		{
			ID:            "8005",
			RecipientID:   "5001",
			RecipientName: "林一",
			SenderID:      "system",
			SenderName:    "系统",
			Type:          TypeAccount,
			Priority:      "normal",
			Title:         "邮箱验证成功",
			Body:          "你的账号已经完成邮箱验证。",
			ReadAt:        &readAt,
			ArchivedAt:    &archivedAt,
			CreatedAt:     now.Add(-72 * time.Hour),
		},
	}
}
