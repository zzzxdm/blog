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
)

type session struct {
	UserID    string
	ExpiresAt time.Time
}

type authToken struct {
	UserID    string
	ExpiresAt time.Time
}

type Store interface {
	Authenticate(email string, password string) (User, string, error)
	Register(request RegisterRequest) (User, string, error)
	UserBySession(token string) (User, error)
	ChangePassword(userID string, currentPassword string, newPassword string) error
	RequestEmailVerification(userID string) (string, error)
	VerifyEmail(token string) (User, error)
	RequestPasswordReset(email string) (string, error)
	ResetPassword(token string, newPassword string) error
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

	return store
}

func (store *MemoryStore) Authenticate(email string, password string) (User, string, error) {
	normalizedEmail := normalizeEmail(email)

	store.mu.RLock()
	userID, ok := store.usersByEmail[normalizedEmail]
	hash := store.passwordHashes[userID]
	user := store.usersByID[userID]
	store.mu.RUnlock()

	if !ok || bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return User{}, "", ErrInvalidCredentials
	}

	token, err := randomToken()
	if err != nil {
		return User{}, "", err
	}

	store.mu.Lock()
	store.sessions[token] = session{
		UserID:    user.ID,
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
		ExpiresAt: store.now().Add(7 * 24 * time.Hour),
	}

	return user, token, nil
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
