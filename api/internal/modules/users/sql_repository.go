package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"blog/api/internal/database"
	"blog/api/internal/modules/auth"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(ctx context.Context, db *sql.DB) (*SQLRepository, error) {
	repo := &SQLRepository{db: db}
	if err := repo.ensureAccountSettings(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *SQLRepository) List(ctx context.Context, query ListQuery) (UserListResult, error) {
	where, args := userListWhere(query)
	stats, err := repo.userStats(ctx)
	if err != nil {
		return UserListResult{}, err
	}

	total, err := repo.countManagedUsers(ctx, where, args)
	if err != nil {
		return UserListResult{}, err
	}

	page := normalizeUserPage(query.Page)
	pageSize := normalizeUserPageSize(query.PageSize)
	suffix := ""
	queryArgs := args
	if query.All {
		page = 1
		pageSize = total
	} else {
		suffix = fmt.Sprintf("LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		queryArgs = append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	}

	items, err := repo.queryManagedUsers(ctx, where, queryArgs, suffix)
	if err != nil {
		return UserListResult{}, err
	}

	return UserListResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Stats:    stats,
	}, nil
}

func (repo *SQLRepository) Get(ctx context.Context, userID string) (ManagedUser, error) {
	return repo.getManagedUser(ctx, userID)
}

func (repo *SQLRepository) EnsureFromAuth(ctx context.Context, user auth.User) (ManagedUser, error) {
	return repo.getManagedUser(ctx, user.ID)
}

func (repo *SQLRepository) UpdateStatus(ctx context.Context, userID string, status string) (ManagedUser, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if !validStatus(status) {
		return ManagedUser{}, ErrInvalidStatus
	}

	result, err := repo.db.ExecContext(ctx, "UPDATE users SET status = $2 WHERE id = $1", userID, status)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("update user status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ManagedUser{}, fmt.Errorf("read user status rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ManagedUser{}, ErrUserNotFound
	}

	return repo.getManagedUser(ctx, userID)
}

func (repo *SQLRepository) GetAccount(ctx context.Context, user auth.User) (AccountSettings, error) {
	settings, err := repo.getAccountSettings(ctx, user)
	if err != nil {
		return AccountSettings{}, err
	}

	return settings, nil
}

func (repo *SQLRepository) UpdateAccount(ctx context.Context, user auth.User, settings AccountSettings) (AccountSettings, error) {
	settings = normalizeAccountSettings(settings, user)
	settings.UpdatedAt = time.Now()

	if err := repo.saveAccountSettings(ctx, user.ID, settings); err != nil {
		return AccountSettings{}, err
	}

	if _, err := repo.db.ExecContext(ctx, `
		UPDATE users
		SET display_name = $2, avatar_text = $3
		WHERE id = $1
	`, user.ID, settings.DisplayName, settings.AvatarText); err != nil {
		return AccountSettings{}, fmt.Errorf("update account user profile: %w", err)
	}

	return settings, nil
}

func (repo *SQLRepository) ensureAccountSettings(ctx context.Context) error {
	users, err := repo.queryManagedUsers(ctx, "", nil, "")
	if err != nil {
		return err
	}

	for _, user := range users {
		settings := accountFromUser(user)
		if err := repo.ensureAccountSettingsForUser(ctx, user.ID, settings); err != nil {
			return err
		}
	}

	return nil
}

func (repo *SQLRepository) queryManagedUsers(ctx context.Context, where string, args []any, suffix string) ([]ManagedUser, error) {
	query := `
			SELECT
				u.id,
			u.email,
			u.display_name,
			u.role,
			u.status,
			u.avatar_text,
			u.email_verified,
			COALESCE((SELECT max(s.created_at) FROM sessions s WHERE s.user_id = u.id), u.created_at) AS last_login_at,
			u.created_at,
				(SELECT count(*) FROM comments c WHERE c.author_id = u.id) AS comment_count,
				(SELECT count(*) FROM post_bookmarks b WHERE b.user_id = u.id) AS bookmark_count
			FROM users u
			` + where + `
			ORDER BY u.created_at DESC
			` + suffix + `
		`

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}

	items := make([]ManagedUser, 0)
	for rows.Next() {
		var user ManagedUser
		var lastLoginAt database.FlexibleTime
		var registeredAt database.FlexibleTime
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.DisplayName,
			&user.Role,
			&user.Status,
			&user.AvatarText,
			&user.EmailVerified,
			&lastLoginAt,
			&registeredAt,
			&user.CommentCount,
			&user.BookmarkCount,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		user.LastLoginAt = lastLoginAt.Time
		user.RegisteredAt = registeredAt.Time
		user.ModerationNote = moderationNote(user.Status)
		items = append(items, user)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close user rows: %w", err)
	}

	for index := range items {
		user := items[index]
		if settings, err := repo.getAccountSettings(ctx, auth.User{
			ID:            user.ID,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			Role:          user.Role,
			Status:        user.Status,
			AvatarText:    user.AvatarText,
			EmailVerified: user.EmailVerified,
		}); err == nil {
			items[index].TwoFactor = settings.TwoFactor
		}
	}

	return items, nil
}

func (repo *SQLRepository) getManagedUser(ctx context.Context, userID string) (ManagedUser, error) {
	items, err := repo.queryManagedUsers(ctx, "WHERE u.id = $1", []any{userID}, "")
	if err != nil {
		return ManagedUser{}, err
	}
	if len(items) == 0 {
		return ManagedUser{}, ErrUserNotFound
	}

	return items[0], nil
}

func userListWhere(query ListQuery) (string, []any) {
	conditions := make([]string, 0, 3)
	args := make([]any, 0, 3)

	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		args = append(args, "%"+keyword+"%")
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, "(lower(u.id) LIKE '%' || lower("+placeholder+") || '%' OR lower(u.email) LIKE '%' || lower("+placeholder+") || '%' OR lower(u.display_name) LIKE '%' || lower("+placeholder+") || '%' OR lower(u.role) LIKE '%' || lower("+placeholder+") || '%' OR lower(u.status) LIKE '%' || lower("+placeholder+") || '%')")
	}

	if status := strings.ToLower(strings.TrimSpace(query.Status)); status != "" {
		if status == "unverified" {
			conditions = append(conditions, "u.email_verified = false")
		} else {
			args = append(args, status)
			conditions = append(conditions, fmt.Sprintf("u.status = $%d", len(args)))
		}
	}

	if role := strings.ToLower(strings.TrimSpace(query.Role)); role != "" {
		args = append(args, role)
		conditions = append(conditions, fmt.Sprintf("u.role = $%d", len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (repo *SQLRepository) countManagedUsers(ctx context.Context, where string, args []any) (int, error) {
	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT count(*) FROM users u "+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return total, nil
}

func (repo *SQLRepository) userStats(ctx context.Context) (UserStats, error) {
	var stats UserStats
	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			count(*),
			COALESCE(sum(CASE WHEN email_verified THEN 1 ELSE 0 END), 0),
			COALESCE(sum(CASE WHEN role IN ('author', 'admin', 'editor') THEN 1 ELSE 0 END), 0),
			COALESCE(sum(CASE WHEN status = 'muted' THEN 1 ELSE 0 END), 0),
			COALESCE(sum(CASE WHEN status = 'banned' THEN 1 ELSE 0 END), 0)
		FROM users
	`).Scan(&stats.Total, &stats.EmailVerified, &stats.Authors, &stats.Muted, &stats.Banned); err != nil {
		return UserStats{}, fmt.Errorf("count user stats: %w", err)
	}

	return stats, nil
}

func (repo *SQLRepository) getAccountSettings(ctx context.Context, user auth.User) (AccountSettings, error) {
	var data []byte
	err := repo.db.QueryRowContext(ctx, "SELECT data FROM account_settings WHERE user_id = $1", user.ID).Scan(&data)
	if err != nil {
		if err != sql.ErrNoRows {
			return AccountSettings{}, fmt.Errorf("load account settings: %w", err)
		}

		settings := accountFromUser(ManagedUser{
			ID:           user.ID,
			Email:        user.Email,
			DisplayName:  user.DisplayName,
			Role:         user.Role,
			Status:       user.Status,
			AvatarText:   user.AvatarText,
			RegisteredAt: userCreatedAt(ctx, repo.db, user.ID),
			LastLoginAt:  userCreatedAt(ctx, repo.db, user.ID),
		})
		if err := repo.ensureAccountSettingsForUser(ctx, user.ID, settings); err != nil {
			return AccountSettings{}, err
		}
		return settings, nil
	}

	var settings AccountSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return AccountSettings{}, fmt.Errorf("decode account settings: %w", err)
	}
	settings = normalizeAccountSettings(settings, user)

	return settings, nil
}

func (repo *SQLRepository) ensureAccountSettingsForUser(ctx context.Context, userID string, settings AccountSettings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal account settings: %w", err)
	}

	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO account_settings (user_id, data)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO NOTHING
	`, userID, data); err != nil {
		return fmt.Errorf("seed account settings: %w", err)
	}

	return nil
}

func (repo *SQLRepository) saveAccountSettings(ctx context.Context, userID string, settings AccountSettings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal account settings: %w", err)
	}

	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO account_settings (user_id, data)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET data = EXCLUDED.data
	`, userID, data); err != nil {
		return fmt.Errorf("save account settings: %w", err)
	}

	return nil
}

func moderationNote(status string) string {
	switch status {
	case "muted":
		return "管理员已限制该用户评论和投稿。"
	case "banned":
		return "账号已封禁，禁止登录和互动。"
	default:
		return ""
	}
}

func userCreatedAt(ctx context.Context, db *sql.DB, userID string) time.Time {
	var createdAt time.Time
	_ = db.QueryRowContext(ctx, "SELECT created_at FROM users WHERE id = $1", userID).Scan(&createdAt)
	if createdAt.IsZero() {
		return time.Now()
	}

	return createdAt
}
