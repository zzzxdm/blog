package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type SQLStore struct {
	db  *sql.DB
	now func() time.Time
}

func NewSQLStore(ctx context.Context, db *sql.DB) (*SQLStore, error) {
	store := &SQLStore{
		db:  db,
		now: time.Now,
	}

	if err := store.ensureSeedUsers(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (store *SQLStore) Authenticate(email string, password string) (User, string, error) {
	user, hash, err := store.userByEmail(context.Background(), email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, "", ErrInvalidCredentials
		}
		return User{}, "", err
	}

	if user.Status == "deleted" {
		return User{}, "", ErrAccountDeleted
	}
	if user.Status == "banned" {
		return User{}, "", ErrAccountBanned
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return User{}, "", ErrInvalidCredentials
	}

	token, err := randomToken()
	if err != nil {
		return User{}, "", err
	}

	if _, err := store.db.ExecContext(context.Background(), `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, user.ID, store.now().Add(7*24*time.Hour)); err != nil {
		return User{}, "", fmt.Errorf("insert session: %w", err)
	}

	return user, token, nil
}

func (store *SQLStore) Register(request RegisterRequest) (User, string, error) {
	normalizedEmail := normalizeEmail(request.Email)
	displayName := strings.TrimSpace(request.DisplayName)
	if displayName == "" {
		displayName = strings.Split(normalizedEmail, "@")[0]
	}

	tx, err := store.db.BeginTx(context.Background(), nil)
	if err != nil {
		return User{}, "", err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var existingStatus string
	err = tx.QueryRowContext(context.Background(), `SELECT status FROM users WHERE email = $1`, normalizedEmail).Scan(&existingStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return User{}, "", fmt.Errorf("check email: %w", err)
	}
	if existingStatus == "deleted" {
		return User{}, "", ErrAccountDeleted
	}
	if existingStatus != "" {
		return User{}, "", ErrEmailExists
	}
	userID, err := store.nextUserID(context.Background(), tx)
	if err != nil {
		return User{}, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, "", err
	}

	user := User{
		ID:          userID,
		Email:       normalizedEmail,
		DisplayName: displayName,
		Role:        "user",
		Status:      "active",
		AvatarText:  firstRune(displayName),
	}

	if _, err := tx.ExecContext(context.Background(), `
		INSERT INTO users (id, email, display_name, role, status, avatar_text, email_verified, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.Email, user.DisplayName, user.Role, user.Status, user.AvatarText, user.EmailVerified, string(hash)); err != nil {
		return User{}, "", fmt.Errorf("insert user: %w", err)
	}

	token, err := randomToken()
	if err != nil {
		return User{}, "", err
	}

	if _, err := tx.ExecContext(context.Background(), `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, user.ID, store.now().Add(7*24*time.Hour)); err != nil {
		return User{}, "", fmt.Errorf("insert session: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return User{}, "", err
	}
	committed = true

	return user, token, nil
}

func (store *SQLStore) InviteUser(request InviteUserRequest) (User, InvitationSecrets, error) {
	normalizedEmail := normalizeEmail(request.Email)
	displayName := strings.TrimSpace(request.DisplayName)
	if displayName == "" {
		displayName = strings.Split(normalizedEmail, "@")[0]
	}
	if normalizedEmail == "" {
		return User{}, InvitationSecrets{}, ErrInvalidCredentials
	}

	tx, err := store.db.BeginTx(context.Background(), nil)
	if err != nil {
		return User{}, InvitationSecrets{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var existingStatus string
	err = tx.QueryRowContext(context.Background(), `SELECT status FROM users WHERE email = $1`, normalizedEmail).Scan(&existingStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return User{}, InvitationSecrets{}, fmt.Errorf("check invitation email: %w", err)
	}
	if existingStatus == "deleted" {
		return User{}, InvitationSecrets{}, ErrAccountDeleted
	}
	if existingStatus != "" {
		return User{}, InvitationSecrets{}, ErrEmailExists
	}
	userID, err := store.nextUserID(context.Background(), tx)
	if err != nil {
		return User{}, InvitationSecrets{}, err
	}

	tempPassword, err := randomTemporaryPassword()
	if err != nil {
		return User{}, InvitationSecrets{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return User{}, InvitationSecrets{}, err
	}

	user := User{
		ID:            userID,
		Email:         normalizedEmail,
		DisplayName:   displayName,
		Role:          normalizeRole(request.Role),
		Status:        "active",
		AvatarText:    firstRune(displayName),
		EmailVerified: false,
	}

	if _, err := tx.ExecContext(context.Background(), `
		INSERT INTO users (id, email, display_name, role, status, avatar_text, email_verified, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.Email, user.DisplayName, user.Role, user.Status, user.AvatarText, user.EmailVerified, string(hash)); err != nil {
		return User{}, InvitationSecrets{}, fmt.Errorf("insert invited user: %w", err)
	}

	resetToken, err := randomToken()
	if err != nil {
		return User{}, InvitationSecrets{}, err
	}
	if _, err := tx.ExecContext(context.Background(), `
		INSERT INTO password_reset_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, resetToken, user.ID, store.now().Add(30*time.Minute)); err != nil {
		return User{}, InvitationSecrets{}, fmt.Errorf("insert invitation reset token: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return User{}, InvitationSecrets{}, err
	}
	committed = true

	return user, InvitationSecrets{InitialPassword: tempPassword, ResetToken: resetToken}, nil
}

func (store *SQLStore) UpdateRole(userID string, role string) (User, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if !validRole(role) {
		return User{}, ErrInvalidRole
	}

	result, err := store.db.ExecContext(context.Background(), "UPDATE users SET role = $2 WHERE id = $1", userID, role)
	if err != nil {
		return User{}, fmt.Errorf("update user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return User{}, fmt.Errorf("read user role rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return User{}, ErrInvalidSession
	}

	return store.userByID(context.Background(), userID)
}

func (store *SQLStore) UpdateStatus(userID string, status string) (User, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if !validStatus(status) {
		return User{}, ErrInvalidStatus
	}

	result, err := store.db.ExecContext(context.Background(), "UPDATE users SET status = $2 WHERE id = $1", userID, status)
	if err != nil {
		return User{}, fmt.Errorf("update user status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return User{}, fmt.Errorf("read user status rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return User{}, ErrInvalidSession
	}

	return store.userByID(context.Background(), userID)
}

func (store *SQLStore) UpdateProfile(userID string, displayName string, avatarText string) (User, error) {
	displayName = strings.TrimSpace(displayName)
	avatarText = strings.TrimSpace(avatarText)
	if displayName == "" {
		user, err := store.userByID(context.Background(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return User{}, ErrInvalidSession
			}
			return User{}, err
		}
		displayName = strings.Split(user.Email, "@")[0]
	}
	if avatarText == "" {
		avatarText = firstRune(displayName)
	}

	result, err := store.db.ExecContext(context.Background(), `
		UPDATE users
		SET display_name = $2,
			avatar_text = $3
		WHERE id = $1
	`, userID, displayName, avatarText)
	if err != nil {
		return User{}, fmt.Errorf("update user profile: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return User{}, fmt.Errorf("read user profile rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return User{}, ErrInvalidSession
	}

	return store.userByID(context.Background(), userID)
}

func (store *SQLStore) UserBySession(token string) (User, error) {
	var user User
	err := store.db.QueryRowContext(context.Background(), `
		SELECT u.id, u.email, u.display_name, u.role, u.status, u.avatar_text, u.email_verified
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = $1
			AND s.expires_at > $2
	`, token, store.now()).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.AvatarText, &user.EmailVerified)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrInvalidSession
		}
		return User{}, err
	}
	if user.Status == "banned" || user.Status == "deleted" {
		return User{}, ErrInvalidSession
	}

	return user, nil
}

func (store *SQLStore) SetSessionExpiry(token string, expiresAt time.Time) error {
	result, err := store.db.ExecContext(context.Background(), "UPDATE sessions SET expires_at = $2 WHERE token = $1", token, expiresAt)
	if err != nil {
		return fmt.Errorf("update session expiry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read session expiry rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrInvalidSession
	}

	return nil
}

func (store *SQLStore) ChangePassword(userID string, currentPassword string, newPassword string) error {
	var hash string
	err := store.db.QueryRowContext(context.Background(), "SELECT password_hash FROM users WHERE id = $1", userID).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)) != nil {
		return ErrInvalidCredentials
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if _, err := store.db.ExecContext(context.Background(), "UPDATE users SET password_hash = $2 WHERE id = $1", userID, string(newHash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

func (store *SQLStore) RequestEmailVerification(userID string) (string, error) {
	var exists bool
	if err := store.db.QueryRowContext(context.Background(), "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists); err != nil {
		return "", fmt.Errorf("check verification user: %w", err)
	}
	if !exists {
		return "", ErrInvalidSession
	}

	token, err := randomToken()
	if err != nil {
		return "", err
	}

	if _, err := store.db.ExecContext(context.Background(), `
		INSERT INTO email_verification_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, userID, store.now().Add(24*time.Hour)); err != nil {
		return "", fmt.Errorf("insert email verification token: %w", err)
	}

	return token, nil
}

func (store *SQLStore) VerifyEmail(token string) (User, error) {
	tx, err := store.db.BeginTx(context.Background(), nil)
	if err != nil {
		return User{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var userID string
	err = tx.QueryRowContext(context.Background(), `
		SELECT user_id
		FROM email_verification_tokens
		WHERE token = $1
			AND expires_at > $2
	`, strings.TrimSpace(token), store.now()).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrInvalidToken
		}
		return User{}, fmt.Errorf("load verification token: %w", err)
	}

	var user User
	err = tx.QueryRowContext(context.Background(), `
		UPDATE users
		SET email_verified = true
		WHERE id = $1
		RETURNING id, email, display_name, role, status, avatar_text, email_verified
	`, userID).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.AvatarText, &user.EmailVerified)
	if err != nil {
		return User{}, fmt.Errorf("verify user email: %w", err)
	}

	if _, err := tx.ExecContext(context.Background(), "DELETE FROM email_verification_tokens WHERE user_id = $1", userID); err != nil {
		return User{}, fmt.Errorf("delete verification tokens: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	committed = true

	return user, nil
}

func (store *SQLStore) RequestPasswordReset(email string) (User, string, error) {
	user, _, err := store.userByEmail(context.Background(), email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, "", nil
		}
		return User{}, "", err
	}
	if user.Status == "banned" || user.Status == "deleted" {
		return User{}, "", nil
	}

	token, err := randomToken()
	if err != nil {
		return User{}, "", err
	}

	if _, err := store.db.ExecContext(context.Background(), `
		INSERT INTO password_reset_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, user.ID, store.now().Add(30*time.Minute)); err != nil {
		return User{}, "", fmt.Errorf("insert password reset token: %w", err)
	}

	return user, token, nil
}

func (store *SQLStore) ResetPassword(token string, newPassword string) error {
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := store.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var userID string
	err = tx.QueryRowContext(context.Background(), `
		SELECT user_id
		FROM password_reset_tokens
		WHERE token = $1
			AND expires_at > $2
			AND used_at IS NULL
	`, strings.TrimSpace(token), store.now()).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidToken
		}
		return fmt.Errorf("load password reset token: %w", err)
	}

	if _, err := tx.ExecContext(context.Background(), "UPDATE users SET password_hash = $2, email_verified = true WHERE id = $1", userID, string(newHash)); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	if _, err := tx.ExecContext(context.Background(), "UPDATE password_reset_tokens SET used_at = $2 WHERE token = $1", strings.TrimSpace(token), store.now()); err != nil {
		return fmt.Errorf("mark password reset token used: %w", err)
	}
	if _, err := tx.ExecContext(context.Background(), "DELETE FROM sessions WHERE user_id = $1", userID); err != nil {
		return fmt.Errorf("delete reset user sessions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true

	return nil
}

func (store *SQLStore) ListSessions(userID string, currentToken string) ([]SessionInfo, error) {
	return store.listSessions(context.Background(), userID, currentToken)
}

func (store *SQLStore) DeleteUserSession(userID string, sessionID string) error {
	result, err := store.db.ExecContext(context.Background(), "DELETE FROM sessions WHERE user_id = $1 AND token = $2", userID, sessionID)
	if err != nil {
		return fmt.Errorf("delete user session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user session rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrInvalidSession
	}

	return nil
}

func (store *SQLStore) ExportUserData(userID string, currentToken string) (ExportData, error) {
	user, err := store.userByID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExportData{}, ErrInvalidSession
		}
		return ExportData{}, err
	}

	sessions, err := store.listSessions(context.Background(), userID, currentToken)
	if err != nil {
		return ExportData{}, err
	}

	var commentCount int
	_ = store.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM comments WHERE author_id = $1", userID).Scan(&commentCount)

	var bookmarkCount int
	_ = store.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM post_bookmarks WHERE user_id = $1", userID).Scan(&bookmarkCount)

	return ExportData{
		User:          user,
		Sessions:      sessions,
		CommentCount:  commentCount,
		BookmarkCount: bookmarkCount,
		ExportedAt:    store.now().UTC().Format(time.RFC3339),
	}, nil
}

func (store *SQLStore) DeleteUser(userID string) error {
	tx, err := store.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(context.Background(), "UPDATE users SET status = 'deleted' WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete account rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrInvalidSession
	}

	if _, err := tx.ExecContext(context.Background(), "DELETE FROM sessions WHERE user_id = $1", userID); err != nil {
		return fmt.Errorf("delete account sessions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true

	return nil
}

func (store *SQLStore) DeleteSession(token string) {
	_, _ = store.db.ExecContext(context.Background(), `DELETE FROM sessions WHERE token = $1`, token)
}

func (store *SQLStore) listSessions(ctx context.Context, userID string, currentToken string) ([]SessionInfo, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT token, created_at, expires_at
		FROM sessions
		WHERE user_id = $1
			AND expires_at > $2
		ORDER BY created_at DESC
	`, userID, store.now())
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()

	items := make([]SessionInfo, 0)
	for rows.Next() {
		var token string
		var createdAt time.Time
		var expiresAt time.Time
		if err := rows.Scan(&token, &createdAt, &expiresAt); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}

		items = append(items, SessionInfo{
			ID:        token,
			Device:    "Web 浏览器",
			Current:   token == currentToken,
			CreatedAt: createdAt.UTC().Format(time.RFC3339),
			ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}

	return items, nil
}

func (store *SQLStore) userByEmail(ctx context.Context, email string) (User, string, error) {
	var user User
	var hash string
	err := store.db.QueryRowContext(ctx, `
		SELECT id, email, display_name, role, status, avatar_text, email_verified, password_hash
		FROM users
		WHERE email = $1
	`, normalizeEmail(email)).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.AvatarText, &user.EmailVerified, &hash)

	return user, hash, err
}

func (store *SQLStore) userByID(ctx context.Context, id string) (User, error) {
	var user User
	err := store.db.QueryRowContext(ctx, `
		SELECT id, email, display_name, role, status, avatar_text, email_verified
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.AvatarText, &user.EmailVerified)

	return user, err
}

func (store *SQLStore) ensureSeedUsers(ctx context.Context) error {
	seedUsers := []struct {
		User     User
		Password string
	}{
		{
			User: User{
				ID:            "5001",
				Email:         "linyi@example.com",
				DisplayName:   "林一",
				Role:          "user",
				Status:        "active",
				AvatarText:    "林",
				EmailVerified: true,
			},
			Password: "password",
		},
		{
			User: User{
				ID:            "5002",
				Email:         "admin@example.com",
				DisplayName:   "管理员",
				Role:          "admin",
				Status:        "active",
				AvatarText:    "管",
				EmailVerified: true,
			},
			Password: "password",
		},
		{
			User: User{
				ID:            "5003",
				Email:         "chen@example.com",
				DisplayName:   "陈默",
				Role:          "user",
				Status:        "active",
				AvatarText:    "陈",
				EmailVerified: true,
			},
			Password: "password",
		},
		{
			User: User{
				ID:            "5004",
				Email:         "market@example.com",
				DisplayName:   "market_user",
				Role:          "user",
				Status:        "muted",
				AvatarText:    "m",
				EmailVerified: true,
			},
			Password: "password",
		},
		{
			User: User{
				ID:            "5005",
				Email:         "noise@example.com",
				DisplayName:   "noise_2048",
				Role:          "user",
				Status:        "banned",
				AvatarText:    "n",
				EmailVerified: true,
			},
			Password: "password",
		},
	}

	for _, seed := range seedUsers {
		hash, err := bcrypt.GenerateFromPassword([]byte(seed.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		if _, err := store.db.ExecContext(ctx, `
				INSERT INTO users (id, email, display_name, role, status, avatar_text, email_verified, password_hash)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (id) DO NOTHING
			`, seed.User.ID, normalizeEmail(seed.User.Email), seed.User.DisplayName, seed.User.Role, seed.User.Status, seed.User.AvatarText, seed.User.EmailVerified, string(hash)); err != nil {
			return fmt.Errorf("seed user %s: %w", seed.User.ID, err)
		}
	}

	return nil
}

func (store *SQLStore) nextUserID(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}) (string, error) {
	for attempts := 0; attempts < 5; attempts++ {
		id, err := randomUserID()
		if err != nil {
			return "", err
		}

		var exists bool
		if err := queryer.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, id).Scan(&exists); err != nil {
			return "", fmt.Errorf("check generated user id: %w", err)
		}
		if !exists {
			return id, nil
		}
	}

	return "", errors.New("failed to generate unique user id")
}
