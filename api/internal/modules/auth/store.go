package auth

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrAccountBanned      = errors.New("account banned")
	ErrAccountDeleted     = errors.New("account deleted")
	ErrInvalidSession     = errors.New("invalid session")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidStatus      = errors.New("invalid status")
)

type session struct {
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type authToken struct {
	UserID    string
	ExpiresAt time.Time
}

type InvitationSecrets struct {
	InitialPassword string
	ResetToken      string
}

type Store interface {
	Authenticate(email string, password string) (User, string, error)
	Register(request RegisterRequest) (User, string, error)
	InviteUser(request InviteUserRequest) (User, InvitationSecrets, error)
	UpdateRole(userID string, role string) (User, error)
	UpdateStatus(userID string, status string) (User, error)
	UpdateProfile(userID string, displayName string, avatarText string) (User, error)
	UserBySession(token string) (User, error)
	SetSessionExpiry(token string, expiresAt time.Time) error
	ChangePassword(userID string, currentPassword string, newPassword string) error
	RequestEmailVerification(userID string) (string, error)
	VerifyEmail(token string) (User, error)
	RequestPasswordReset(email string) (User, string, error)
	ResetPassword(token string, newPassword string) error
	ListSessions(userID string, currentToken string) ([]SessionInfo, error)
	DeleteUserSession(userID string, sessionID string) error
	ExportUserData(userID string, currentToken string) (ExportData, error)
	DeleteUser(userID string) error
	DeleteSession(token string)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func randomTemporaryPassword() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func randomUserID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
	return "usr_" + strings.ToLower(encoded), nil
}

func firstRune(value string) string {
	for _, item := range strings.TrimSpace(value) {
		return string(item)
	}

	return "用"
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "editor", "author":
		return strings.ToLower(strings.TrimSpace(role))
	default:
		return "author"
	}
}

func validRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "user", "author", "editor", "admin":
		return true
	default:
		return false
	}
}

func validStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "muted", "banned", "deleted":
		return true
	default:
		return false
	}
}

func sessionInfo(token string, item session, current bool) SessionInfo {
	return SessionInfo{
		ID:        token,
		Device:    "Web 浏览器",
		Current:   current,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		ExpiresAt: item.ExpiresAt.UTC().Format(time.RFC3339),
	}
}
