package users

import (
	"context"
	"database/sql"
	"encoding/json"
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
	if err := repo.ensureAccountSettings(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *SQLRepository) List(ctx context.Context) (UserListResult, error) {
	items, err := repo.queryManagedUsers(ctx, "", nil)
	if err != nil {
		return UserListResult{}, err
	}

	return UserListResult{
		Items: items,
		Total: len(items),
		Stats: countStats(items),
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
	users, err := repo.queryManagedUsers(ctx, "", nil)
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

func (repo *SQLRepository) queryManagedUsers(ctx context.Context, where string, arg any) ([]ManagedUser, error) {
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
			(SELECT count(*)::int FROM comments c WHERE c.author_id = u.id) AS comment_count,
			(SELECT count(*)::int FROM post_bookmarks b WHERE b.user_id = u.id) AS bookmark_count
		FROM users u
		` + where + `
		ORDER BY u.created_at DESC
	`

	var rows *sql.Rows
	var err error
	if where == "" {
		rows, err = repo.db.QueryContext(ctx, query)
	} else {
		rows, err = repo.db.QueryContext(ctx, query, arg)
	}
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	items := make([]ManagedUser, 0)
	for rows.Next() {
		var user ManagedUser
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.DisplayName,
			&user.Role,
			&user.Status,
			&user.AvatarText,
			&user.EmailVerified,
			&user.LastLoginAt,
			&user.RegisteredAt,
			&user.CommentCount,
			&user.BookmarkCount,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		user.ModerationNote = moderationNote(user.Status)
		if settings, err := repo.getAccountSettings(ctx, auth.User{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        user.Role,
			Status:      user.Status,
			AvatarText:  user.AvatarText,
		}); err == nil {
			user.TwoFactor = settings.TwoFactor
		}
		items = append(items, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return items, nil
}

func (repo *SQLRepository) getManagedUser(ctx context.Context, userID string) (ManagedUser, error) {
	items, err := repo.queryManagedUsers(ctx, "WHERE u.id = $1", userID)
	if err != nil {
		return ManagedUser{}, err
	}
	if len(items) == 0 {
		return ManagedUser{}, ErrUserNotFound
	}

	return items[0], nil
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
