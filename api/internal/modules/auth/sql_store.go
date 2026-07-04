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

	if user.Status != "active" || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
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

	var exists bool
	if err := tx.QueryRowContext(context.Background(), `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, normalizedEmail).Scan(&exists); err != nil {
		return User{}, "", fmt.Errorf("check email: %w", err)
	}
	if exists {
		return User{}, "", ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, "", err
	}

	user := User{
		ID:          uniqueUserID(normalizedEmail),
		Email:       normalizedEmail,
		DisplayName: displayName,
		Role:        "user",
		Status:      "active",
		AvatarText:  firstRune(displayName),
	}

	if _, err := tx.ExecContext(context.Background(), `
		INSERT INTO users (id, email, display_name, role, status, avatar_text, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, user.ID, user.Email, user.DisplayName, user.Role, user.Status, user.AvatarText, string(hash)); err != nil {
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

func (store *SQLStore) UserBySession(token string) (User, error) {
	var user User
	err := store.db.QueryRowContext(context.Background(), `
		SELECT u.id, u.email, u.display_name, u.role, u.status, u.avatar_text
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = $1
			AND s.expires_at > now()
	`, token).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.AvatarText)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrInvalidSession
		}
		return User{}, err
	}

	return user, nil
}

func (store *SQLStore) DeleteSession(token string) {
	_, _ = store.db.ExecContext(context.Background(), `DELETE FROM sessions WHERE token = $1`, token)
}

func (store *SQLStore) userByEmail(ctx context.Context, email string) (User, string, error) {
	var user User
	var hash string
	err := store.db.QueryRowContext(ctx, `
		SELECT id, email, display_name, role, status, avatar_text, password_hash
		FROM users
		WHERE email = $1
	`, normalizeEmail(email)).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.AvatarText, &hash)

	return user, hash, err
}

func (store *SQLStore) ensureSeedUsers(ctx context.Context) error {
	seedUsers := []struct {
		User     User
		Password string
	}{
		{
			User: User{
				ID:          "user_linyi",
				Email:       "linyi@example.com",
				DisplayName: "林一",
				Role:        "user",
				Status:      "active",
				AvatarText:  "林",
			},
			Password: "password",
		},
		{
			User: User{
				ID:          "user_admin",
				Email:       "admin@example.com",
				DisplayName: "管理员",
				Role:        "admin",
				Status:      "active",
				AvatarText:  "管",
			},
			Password: "password",
		},
		{
			User: User{
				ID:          "user_chen",
				Email:       "chen@example.com",
				DisplayName: "陈默",
				Role:        "user",
				Status:      "active",
				AvatarText:  "陈",
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
			INSERT INTO users (id, email, display_name, role, status, avatar_text, password_hash)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				email = EXCLUDED.email,
				display_name = EXCLUDED.display_name,
				role = EXCLUDED.role,
				status = EXCLUDED.status,
				avatar_text = EXCLUDED.avatar_text
		`, seed.User.ID, normalizeEmail(seed.User.Email), seed.User.DisplayName, seed.User.Role, seed.User.Status, seed.User.AvatarText, string(hash)); err != nil {
			return fmt.Errorf("seed user %s: %w", seed.User.ID, err)
		}
	}

	return nil
}

func uniqueUserID(email string) string {
	local := strings.Split(normalizeEmail(email), "@")[0]
	local = strings.ReplaceAll(local, ".", "_")
	local = strings.ReplaceAll(local, "-", "_")
	if local == "" {
		local = "user"
	}

	return fmt.Sprintf("user_%s_%d", local, time.Now().UnixNano())
}
