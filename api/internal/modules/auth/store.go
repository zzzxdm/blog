package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidSession     = errors.New("invalid session")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidRole        = errors.New("invalid role")
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

type Store interface {
	Authenticate(email string, password string) (User, string, error)
	Register(request RegisterRequest) (User, string, error)
	InviteUser(request InviteUserRequest) (User, string, error)
	UpdateRole(userID string, role string) (User, error)
	UserBySession(token string) (User, error)
	ChangePassword(userID string, currentPassword string, newPassword string) error
	RequestEmailVerification(userID string) (string, error)
	VerifyEmail(token string) (User, error)
	RequestPasswordReset(email string) (string, error)
	ResetPassword(token string, newPassword string) error
	ListSessions(userID string, currentToken string) ([]SessionInfo, error)
	DeleteUserSession(userID string, sessionID string) error
	ExportUserData(userID string, currentToken string) (ExportData, error)
	DeleteUser(userID string) error
	DeleteSession(token string)
}

type MemoryStore struct {
	mu             sync.RWMutex
	usersByID      map[string]User
	usersByEmail   map[string]string
	passwordHashes map[string][]byte
	sessions       map[string]session
	emailTokens    map[string]authToken
	resetTokens    map[string]authToken
	now            func() time.Time
}

func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		usersByID:      map[string]User{},
		usersByEmail:   map[string]string{},
		passwordHashes: map[string][]byte{},
		sessions:       map[string]session{},
		emailTokens:    map[string]authToken{},
		resetTokens:    map[string]authToken{},
		now:            time.Now,
	}

	store.mustSeed(User{
		ID:            "user_linyi",
		Email:         "linyi@example.com",
		DisplayName:   "林一",
		Role:          "user",
		Status:        "active",
		AvatarText:    "林",
		EmailVerified: true,
	}, "password")

	store.mustSeed(User{
		ID:            "user_admin",
		Email:         "admin@example.com",
		DisplayName:   "管理员",
		Role:          "admin",
		Status:        "active",
		AvatarText:    "管",
		EmailVerified: true,
	}, "password")

	store.mustSeed(User{
		ID:            "user_chen",
		Email:         "chen@example.com",
		DisplayName:   "陈默",
		Role:          "user",
		Status:        "active",
		AvatarText:    "陈",
		EmailVerified: true,
	}, "password")

	store.mustSeed(User{
		ID:            "user_market",
		Email:         "market@example.com",
		DisplayName:   "market_user",
		Role:          "user",
		Status:        "muted",
		AvatarText:    "m",
		EmailVerified: true,
	}, "password")

	store.mustSeed(User{
		ID:            "user_noise",
		Email:         "noise@example.com",
		DisplayName:   "noise_2048",
		Role:          "user",
		Status:        "banned",
		AvatarText:    "n",
		EmailVerified: true,
	}, "password")

	return store
}

func (store *MemoryStore) Authenticate(email string, password string) (User, string, error) {
	normalizedEmail := normalizeEmail(email)

	store.mu.RLock()
	userID, ok := store.usersByEmail[normalizedEmail]
	hash := store.passwordHashes[userID]
	user := store.usersByID[userID]
	store.mu.RUnlock()

	if !ok || user.Status == "banned" || user.Status == "deleted" || bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return User{}, "", ErrInvalidCredentials
	}

	token, err := randomToken()
	if err != nil {
		return User{}, "", err
	}

	store.mu.Lock()
	store.sessions[token] = session{
		UserID:    user.ID,
		CreatedAt: store.now(),
		ExpiresAt: store.now().Add(7 * 24 * time.Hour),
	}
	store.mu.Unlock()

	return user, token, nil
}

func (store *MemoryStore) Register(request RegisterRequest) (User, string, error) {
	normalizedEmail := normalizeEmail(request.Email)
	displayName := strings.TrimSpace(request.DisplayName)
	if displayName == "" {
		displayName = strings.Split(normalizedEmail, "@")[0]
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.usersByEmail[normalizedEmail]; ok {
		return User{}, "", ErrEmailExists
	}

	user := User{
		ID:            "user_" + strings.ReplaceAll(strings.Split(normalizedEmail, "@")[0], ".", "_"),
		Email:         normalizedEmail,
		DisplayName:   displayName,
		Role:          "user",
		Status:        "active",
		AvatarText:    firstRune(displayName),
		EmailVerified: false,
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, "", err
	}

	token, err := randomToken()
	if err != nil {
		return User{}, "", err
	}

	store.usersByID[user.ID] = user
	store.usersByEmail[normalizedEmail] = user.ID
	store.passwordHashes[user.ID] = hash
	store.sessions[token] = session{
		UserID:    user.ID,
		CreatedAt: store.now(),
		ExpiresAt: store.now().Add(7 * 24 * time.Hour),
	}

	return user, token, nil
}

func (store *MemoryStore) InviteUser(request InviteUserRequest) (User, string, error) {
	normalizedEmail := normalizeEmail(request.Email)
	displayName := strings.TrimSpace(request.DisplayName)
	if displayName == "" {
		displayName = strings.Split(normalizedEmail, "@")[0]
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if normalizedEmail == "" {
		return User{}, "", ErrInvalidCredentials
	}
	if _, ok := store.usersByEmail[normalizedEmail]; ok {
		return User{}, "", ErrEmailExists
	}

	resetToken, err := randomToken()
	if err != nil {
		return User{}, "", err
	}
	tempPassword, err := randomToken()
	if err != nil {
		return User{}, "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return User{}, "", err
	}

	user := User{
		ID:            uniqueUserID(normalizedEmail),
		Email:         normalizedEmail,
		DisplayName:   displayName,
		Role:          normalizeRole(request.Role),
		Status:        "active",
		AvatarText:    firstRune(displayName),
		EmailVerified: false,
	}

	store.usersByID[user.ID] = user
	store.usersByEmail[normalizedEmail] = user.ID
	store.passwordHashes[user.ID] = hash
	store.resetTokens[resetToken] = authToken{
		UserID:    user.ID,
		ExpiresAt: store.now().Add(30 * time.Minute),
	}

	return user, resetToken, nil
}

func (store *MemoryStore) UpdateRole(userID string, role string) (User, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if !validRole(role) {
		return User{}, ErrInvalidRole
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	user, ok := store.usersByID[userID]
	if !ok {
		return User{}, ErrInvalidSession
	}

	user.Role = role
	store.usersByID[userID] = user

	return user, nil
}

func (store *MemoryStore) UserBySession(token string) (User, error) {
	store.mu.RLock()
	session, ok := store.sessions[token]
	user := store.usersByID[session.UserID]
	store.mu.RUnlock()

	if !ok || session.ExpiresAt.Before(store.now()) {
		return User{}, ErrInvalidSession
	}

	return user, nil
}

func (store *MemoryStore) ChangePassword(userID string, currentPassword string, newPassword string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	hash, ok := store.passwordHashes[userID]
	if !ok || bcrypt.CompareHashAndPassword(hash, []byte(currentPassword)) != nil {
		return ErrInvalidCredentials
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	store.passwordHashes[userID] = newHash
	return nil
}

func (store *MemoryStore) RequestEmailVerification(userID string) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.usersByID[userID]; !ok {
		return "", ErrInvalidSession
	}

	store.emailTokens[token] = authToken{
		UserID:    userID,
		ExpiresAt: store.now().Add(24 * time.Hour),
	}

	return token, nil
}

func (store *MemoryStore) VerifyEmail(token string) (User, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	normalizedToken := strings.TrimSpace(token)
	record, ok := store.emailTokens[normalizedToken]
	if !ok || record.ExpiresAt.Before(store.now()) {
		return User{}, ErrInvalidToken
	}

	user, ok := store.usersByID[record.UserID]
	if !ok {
		return User{}, ErrInvalidToken
	}

	user.EmailVerified = true
	store.usersByID[user.ID] = user
	delete(store.emailTokens, normalizedToken)

	return user, nil
}

func (store *MemoryStore) RequestPasswordReset(email string) (string, error) {
	normalizedEmail := normalizeEmail(email)

	token, err := randomToken()
	if err != nil {
		return "", err
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	userID, ok := store.usersByEmail[normalizedEmail]
	if !ok {
		return "", nil
	}

	user := store.usersByID[userID]
	if user.Status == "banned" || user.Status == "deleted" {
		return "", nil
	}

	store.resetTokens[token] = authToken{
		UserID:    userID,
		ExpiresAt: store.now().Add(30 * time.Minute),
	}

	return token, nil
}

func (store *MemoryStore) ResetPassword(token string, newPassword string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	normalizedToken := strings.TrimSpace(token)
	record, ok := store.resetTokens[normalizedToken]
	if !ok || record.ExpiresAt.Before(store.now()) {
		return ErrInvalidToken
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	store.passwordHashes[record.UserID] = newHash
	for sessionToken, session := range store.sessions {
		if session.UserID == record.UserID {
			delete(store.sessions, sessionToken)
		}
	}
	delete(store.resetTokens, normalizedToken)

	return nil
}

func (store *MemoryStore) ListSessions(userID string, currentToken string) ([]SessionInfo, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if _, ok := store.usersByID[userID]; !ok {
		return nil, ErrInvalidSession
	}

	items := make([]SessionInfo, 0)
	for token, session := range store.sessions {
		if session.UserID != userID || session.ExpiresAt.Before(store.now()) {
			continue
		}

		items = append(items, sessionInfo(token, session, token == currentToken))
	}

	return items, nil
}

func (store *MemoryStore) DeleteUserSession(userID string, sessionID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if !ok || session.UserID != userID {
		return ErrInvalidSession
	}

	delete(store.sessions, sessionID)
	return nil
}

func (store *MemoryStore) ExportUserData(userID string, currentToken string) (ExportData, error) {
	store.mu.RLock()
	user, ok := store.usersByID[userID]
	store.mu.RUnlock()
	if !ok {
		return ExportData{}, ErrInvalidSession
	}

	sessions, err := store.ListSessions(userID, currentToken)
	if err != nil {
		return ExportData{}, err
	}

	return ExportData{
		User:       user,
		Sessions:   sessions,
		ExportedAt: store.now().UTC().Format(time.RFC3339),
	}, nil
}

func (store *MemoryStore) DeleteUser(userID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	user, ok := store.usersByID[userID]
	if !ok {
		return ErrInvalidSession
	}

	user.Status = "deleted"
	store.usersByID[userID] = user
	for token, session := range store.sessions {
		if session.UserID == userID {
			delete(store.sessions, token)
		}
	}

	return nil
}

func (store *MemoryStore) DeleteSession(token string) {
	store.mu.Lock()
	delete(store.sessions, token)
	store.mu.Unlock()
}

func (store *MemoryStore) mustSeed(user User, password string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	email := normalizeEmail(user.Email)
	store.usersByID[user.ID] = user
	store.usersByEmail[email] = user.ID
	store.passwordHashes[user.ID] = hash
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

func sessionInfo(token string, item session, current bool) SessionInfo {
	return SessionInfo{
		ID:        token,
		Device:    "Web 浏览器",
		Current:   current,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		ExpiresAt: item.ExpiresAt.UTC().Format(time.RFC3339),
	}
}
