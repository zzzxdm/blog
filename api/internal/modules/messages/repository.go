package messages

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
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

type MemoryRepository struct {
	mu       sync.RWMutex
	messages []Message
	nextID   int
	now      func() time.Time
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		messages: seedMessages(),
		nextID:   100,
		now:      time.Now,
	}
}

func (repo *MemoryRepository) List(_ context.Context, userID string, query ListQuery) (ListResult, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	now := repo.now()
	items := make([]Message, 0)
	for _, message := range repo.messages {
		if message.RecipientID != userID {
			continue
		}

		item := withStatus(message, now)
		if item.Status == StatusScheduled {
			continue
		}
		if !matchesQuery(item, query) {
			continue
		}
		items = append(items, item)
	}

	sortMessages(items)

	return ListResult{
		Items: items,
		Total: len(items),
		Stats: repo.statsLocked(userID, now),
	}, nil
}

func (repo *MemoryRepository) AdminList(_ context.Context, query ListQuery) (ListResult, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	items := make([]Message, 0, len(repo.messages))
	now := repo.now()
	for _, message := range repo.messages {
		item := withStatus(message, now)
		if !matchesQuery(item, query) {
			continue
		}
		items = append(items, item)
	}

	sortMessages(items)

	return ListResult{
		Items: items,
		Total: len(items),
		Stats: repo.adminStatsLocked(now),
	}, nil
}

func (repo *MemoryRepository) Create(_ context.Context, request CreateRequest, sender auth.User) (Message, error) {
	title := strings.TrimSpace(request.Title)
	body := strings.TrimSpace(request.Body)
	recipientID := strings.TrimSpace(request.RecipientID)
	if title == "" || body == "" || recipientID == "" {
		return Message{}, ErrInvalidMessage
	}
	scheduledAt, err := parseScheduledAt(request.ScheduledAt)
	if err != nil {
		return Message{}, ErrInvalidMessage
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	message := Message{
		ID:            fmt.Sprintf("message_%03d", repo.nextID),
		RecipientID:   recipientID,
		RecipientName: defaultString(strings.TrimSpace(request.RecipientName), recipientID),
		SenderID:      defaultString(sender.ID, "system"),
		SenderName:    defaultString(sender.DisplayName, "系统"),
		Type:          normalizeType(request.Type),
		Priority:      defaultString(strings.TrimSpace(request.Priority), "normal"),
		Title:         title,
		Body:          body,
		TargetType:    strings.TrimSpace(request.TargetType),
		TargetID:      strings.TrimSpace(request.TargetID),
		TargetTitle:   strings.TrimSpace(request.TargetTitle),
		Status:        StatusUnread,
		ScheduledAt:   scheduledAt,
		CreatedAt:     repo.now(),
	}

	repo.nextID++
	repo.messages = append(repo.messages, message)

	return withStatus(message, repo.now()), nil
}

func (repo *MemoryRepository) MarkRead(_ context.Context, userID string, messageID string) (Message, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	index := repo.findLocked(userID, messageID)
	if index < 0 {
		return Message{}, ErrMessageNotFound
	}

	now := repo.now()
	if withStatus(repo.messages[index], now).Status == StatusScheduled {
		return Message{}, ErrMessageNotFound
	}
	if repo.messages[index].ReadAt == nil {
		repo.messages[index].ReadAt = &now
	}

	return withStatus(repo.messages[index], now), nil
}

func (repo *MemoryRepository) MarkAllRead(_ context.Context, userID string) (Stats, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	now := repo.now()
	for index := range repo.messages {
		if repo.messages[index].RecipientID != userID || repo.messages[index].ArchivedAt != nil {
			continue
		}
		if withStatus(repo.messages[index], now).Status == StatusScheduled {
			continue
		}
		if repo.messages[index].ReadAt == nil {
			repo.messages[index].ReadAt = &now
		}
	}

	return repo.statsLocked(userID, now), nil
}

func (repo *MemoryRepository) Archive(_ context.Context, userID string, messageID string) (Message, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	index := repo.findLocked(userID, messageID)
	if index < 0 {
		return Message{}, ErrMessageNotFound
	}

	now := repo.now()
	if withStatus(repo.messages[index], now).Status == StatusScheduled {
		return Message{}, ErrMessageNotFound
	}
	if repo.messages[index].ReadAt == nil {
		repo.messages[index].ReadAt = &now
	}
	repo.messages[index].ArchivedAt = &now

	return withStatus(repo.messages[index], now), nil
}

func (repo *MemoryRepository) findLocked(userID string, messageID string) int {
	for index, message := range repo.messages {
		if message.ID == messageID && message.RecipientID == userID {
			return index
		}
	}

	return -1
}

func (repo *MemoryRepository) statsLocked(userID string, now time.Time) Stats {
	stats := Stats{}
	for _, message := range repo.messages {
		if message.RecipientID != userID {
			continue
		}

		item := withStatus(message, now)
		if item.Status == StatusScheduled {
			continue
		}
		stats.Total++
		if item.Status == StatusUnread {
			stats.Unread++
		}
		if item.Status == StatusArchived {
			stats.Archived++
		}
		if item.Type == TypeReview {
			stats.Review++
		}
		if item.Type == TypeAdmin {
			stats.Admin++
		}
	}

	return stats
}

func (repo *MemoryRepository) adminStatsLocked(now time.Time) Stats {
	stats := Stats{}
	for _, message := range repo.messages {
		item := withStatus(message, now)
		stats.Total++
		if item.Status == StatusUnread {
			stats.Unread++
		}
		if item.Status == StatusArchived {
			stats.Archived++
		}
		if item.Status == StatusScheduled {
			stats.Scheduled++
		}
		if item.Type == TypeReview {
			stats.Review++
		}
		if item.Type == TypeAdmin {
			stats.Admin++
		}
	}

	return stats
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

	if status != "" && status != "all" && message.Status != status {
		return false
	}

	if messageType != "" && messageType != "all" && message.Type != messageType {
		return false
	}

	if status == "" && message.Status == StatusArchived {
		return false
	}

	return true
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
			ID:            "message_001",
			RecipientID:   "user_linyi",
			RecipientName: "林一",
			SenderID:      "user_admin",
			SenderName:    "管理员",
			Type:          TypeReview,
			Priority:      "normal",
			Title:         "你的投稿已退回修改",
			Body:          "《如何写一篇可维护的技术文章》需要补充摘要和代码示例。修改后可以重新提交。",
			TargetType:    "submission",
			TargetID:      "submission_003",
			TargetTitle:   "如何写一篇可维护的技术文章",
			CreatedAt:     now.Add(-20 * time.Minute),
		},
		{
			ID:            "message_002",
			RecipientID:   "user_linyi",
			RecipientName: "林一",
			SenderID:      "user_admin",
			SenderName:    "管理员",
			Type:          TypeComment,
			Priority:      "normal",
			Title:         "管理员回复了你的评论",
			Body:          "审核结果会同步到站内信和我的投稿列表。",
			TargetType:    "comment",
			TargetID:      "comment_002",
			TargetTitle:   "如何设计一个内容长期增长的博客系统",
			CreatedAt:     now.Add(-2 * time.Hour),
		},
		{
			ID:            "message_003",
			RecipientID:   "user_linyi",
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
			ID:            "message_004",
			RecipientID:   "user_linyi",
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
			ID:            "message_005",
			RecipientID:   "user_linyi",
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
