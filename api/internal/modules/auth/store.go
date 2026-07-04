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
)

type session struct {
	UserID    string
	ExpiresAt time.Time
}

type Store interface {
	Authenticate(email string, password string) (User, string, error)
	Register(request RegisterRequest) (User, string, error)
	UserBySession(token string) (User, error)
	DeleteSession(token string)
}

type MemoryStore struct {
	mu             sync.RWMutex
	usersByID      map[string]User
	usersByEmail   map[string]string
	passwordHashes map[string][]byte
	sessions       map[string]session
	now            func() time.Time
}

func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		usersByID:      map[string]User{},
		usersByEmail:   map[string]string{},
		passwordHashes: map[string][]byte{},
		sessions:       map[string]session{},
		now:            time.Now,
	}

	store.mustSeed(User{
		ID:          "user_linyi",
		Email:       "linyi@example.com",
		DisplayName: "林一",
		Role:        "user",
		Status:      "active",
		AvatarText:  "林",
	}, "password")

	store.mustSeed(User{
		ID:          "user_admin",
		Email:       "admin@example.com",
		DisplayName: "管理员",
		Role:        "admin",
		Status:      "active",
		AvatarText:  "管",
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
		ID:          "user_" + strings.ReplaceAll(strings.Split(normalizedEmail, "@")[0], ".", "_"),
		Email:       normalizedEmail,
		DisplayName: displayName,
		Role:        "user",
		Status:      "active",
		AvatarText:  firstRune(displayName),
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
