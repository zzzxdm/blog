package messages

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(ctx context.Context, db *sql.DB) (*SQLRepository, error) {
	repo := &SQLRepository{db: db}
	if err := repo.ensureSeedMessages(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *SQLRepository) List(ctx context.Context, userID string, query ListQuery) (ListResult, error) {
	items, err := repo.queryMessages(ctx, `
		WHERE recipient_id = $1
			AND (scheduled_at IS NULL OR scheduled_at <= now())
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return ListResult{}, err
	}

	filtered := filterMessages(items, query)
	stats := countMessageStats(items)

	return pagedMessageResult(filtered, stats, query), nil
}

func (repo *SQLRepository) AdminList(ctx context.Context, query ListQuery) (ListResult, error) {
	items, err := repo.queryMessages(ctx, `ORDER BY created_at DESC`)
	if err != nil {
		return ListResult{}, err
	}

	filtered := filterMessages(items, query)
	stats := countMessageStats(items)

	return pagedMessageResult(filtered, stats, query), nil
}

func (repo *SQLRepository) Create(ctx context.Context, request CreateRequest, sender auth.User) (Message, error) {
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

	id := fmt.Sprintf("message_%d", time.Now().UnixNano())
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO messages (
			id, recipient_id, recipient_name, sender_id, sender_name, type, priority,
		title, body, target_type, target_id, target_title, scheduled_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, id,
		recipientID,
		defaultString(strings.TrimSpace(request.RecipientName), recipientID),
		defaultString(sender.ID, "system"),
		defaultString(sender.DisplayName, "系统"),
		normalizeType(request.Type),
		defaultString(strings.TrimSpace(request.Priority), "normal"),
		title,
		body,
		strings.TrimSpace(request.TargetType),
		strings.TrimSpace(request.TargetID),
		strings.TrimSpace(request.TargetTitle),
		scheduledAt,
	)
	if err != nil {
		return Message{}, fmt.Errorf("insert message: %w", err)
	}

	return repo.getByID(ctx, id)
}

func (repo *SQLRepository) MarkRead(ctx context.Context, userID string, messageID string) (Message, error) {
	var id string
	err := repo.db.QueryRowContext(ctx, `
		UPDATE messages
		SET read_at = COALESCE(read_at, now())
		WHERE id = $1
			AND recipient_id = $2
			AND (scheduled_at IS NULL OR scheduled_at <= now())
		RETURNING id
	`, messageID, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Message{}, ErrMessageNotFound
		}
		return Message{}, fmt.Errorf("mark message read: %w", err)
	}

	return repo.getByID(ctx, id)
}

func (repo *SQLRepository) MarkAllRead(ctx context.Context, userID string) (Stats, error) {
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE messages
		SET read_at = COALESCE(read_at, now())
		WHERE recipient_id = $1
			AND archived_at IS NULL
			AND (scheduled_at IS NULL OR scheduled_at <= now())
	`, userID); err != nil {
		return Stats{}, fmt.Errorf("mark all messages read: %w", err)
	}

	items, err := repo.queryMessages(ctx, `WHERE recipient_id = $1`, userID)
	if err != nil {
		return Stats{}, err
	}

	return countMessageStats(items), nil
}

func (repo *SQLRepository) Archive(ctx context.Context, userID string, messageID string) (Message, error) {
	var id string
	err := repo.db.QueryRowContext(ctx, `
		UPDATE messages
		SET read_at = COALESCE(read_at, now()),
			archived_at = now()
		WHERE id = $1
			AND recipient_id = $2
			AND (scheduled_at IS NULL OR scheduled_at <= now())
		RETURNING id
	`, messageID, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Message{}, ErrMessageNotFound
		}
		return Message{}, fmt.Errorf("archive message: %w", err)
	}

	return repo.getByID(ctx, id)
}

func (repo *SQLRepository) getByID(ctx context.Context, id string) (Message, error) {
	items, err := repo.queryMessages(ctx, `WHERE id = $1`, id)
	if err != nil {
		return Message{}, err
	}
	if len(items) == 0 {
		return Message{}, ErrMessageNotFound
	}

	return items[0], nil
}

func (repo *SQLRepository) queryMessages(ctx context.Context, whereAndOrder string, args ...any) ([]Message, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			id,
			recipient_id,
			recipient_name,
			sender_id,
			sender_name,
			type,
			priority,
			title,
			body,
			target_type,
			target_id,
			target_title,
				read_at,
				archived_at,
				scheduled_at,
				created_at
		FROM messages
		`+whereAndOrder, args...)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	items := make([]Message, 0)
	for rows.Next() {
		var message Message
		if err := rows.Scan(
			&message.ID,
			&message.RecipientID,
			&message.RecipientName,
			&message.SenderID,
			&message.SenderName,
			&message.Type,
			&message.Priority,
			&message.Title,
			&message.Body,
			&message.TargetType,
			&message.TargetID,
			&message.TargetTitle,
			&message.ReadAt,
			&message.ArchivedAt,
			&message.ScheduledAt,
			&message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		items = append(items, withStatus(message, time.Now()))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	return items, nil
}

func filterMessages(items []Message, query ListQuery) []Message {
	filtered := make([]Message, 0, len(items))
	for _, item := range items {
		if matchesQuery(item, query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func countMessageStats(items []Message) Stats {
	stats := Stats{}
	now := time.Now()
	for _, item := range items {
		item = withStatus(item, now)
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

func (repo *SQLRepository) ensureSeedMessages(ctx context.Context) error {
	for _, message := range seedMessages() {
		if _, err := repo.db.ExecContext(ctx, `
			INSERT INTO messages (
				id, recipient_id, recipient_name, sender_id, sender_name, type, priority,
			title, body, target_type, target_id, target_title, read_at, archived_at, scheduled_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO NOTHING
	`,
			message.ID,
			message.RecipientID,
			message.RecipientName,
			message.SenderID,
			message.SenderName,
			normalizeType(message.Type),
			message.Priority,
			message.Title,
			message.Body,
			message.TargetType,
			message.TargetID,
			message.TargetTitle,
			message.ReadAt,
			message.ArchivedAt,
			message.ScheduledAt,
			message.CreatedAt,
		); err != nil {
			return fmt.Errorf("seed message %s: %w", message.ID, err)
		}
	}

	return nil
}
