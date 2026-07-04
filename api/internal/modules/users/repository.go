package users

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"blog/api/internal/modules/auth"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidStatus = errors.New("invalid user status")
)

type Repository interface {
	List(ctx context.Context) (UserListResult, error)
	Get(ctx context.Context, userID string) (ManagedUser, error)
	EnsureFromAuth(ctx context.Context, user auth.User) (ManagedUser, error)
	UpdateStatus(ctx context.Context, userID string, status string) (ManagedUser, error)
	GetAccount(ctx context.Context, user auth.User) (AccountSettings, error)
	UpdateAccount(ctx context.Context, user auth.User, settings AccountSettings) (AccountSettings, error)
}

type MemoryRepository struct {
	mu       sync.RWMutex
	users    map[string]ManagedUser
	accounts map[string]AccountSettings
	now      func() time.Time
}

func NewMemoryRepository() *MemoryRepository {
	users := seedUsers()
	accounts := map[string]AccountSettings{}
	for _, user := range users {
		accounts[user.ID] = accountFromUser(user)
	}

	return &MemoryRepository{
		users:    users,
		accounts: accounts,
		now:      time.Now,
	}
}

func (repo *MemoryRepository) List(_ context.Context) (UserListResult, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	items := make([]ManagedUser, 0, len(repo.users))
	for _, user := range repo.users {
		items = append(items, user)
	}

	return UserListResult{
		Items: items,
		Total: len(items),
		Stats: countStats(items),
	}, nil
}

func (repo *MemoryRepository) Get(_ context.Context, userID string) (ManagedUser, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	user, ok := repo.users[userID]
	if !ok {
		return ManagedUser{}, ErrUserNotFound
	}

	return user, nil
}

func (repo *MemoryRepository) EnsureFromAuth(_ context.Context, user auth.User) (ManagedUser, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	managed, ok := repo.users[user.ID]
	if !ok {
		managed = ManagedUser{
			ID:            user.ID,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			Role:          user.Role,
			Status:        user.Status,
			AvatarText:    user.AvatarText,
			EmailVerified: user.EmailVerified,
			RegisteredAt:  repo.now(),
			LastLoginAt:   repo.now(),
		}
	} else {
		managed.Email = user.Email
		managed.DisplayName = user.DisplayName
		managed.Role = user.Role
		managed.Status = user.Status
		managed.AvatarText = user.AvatarText
		managed.EmailVerified = user.EmailVerified
	}

	repo.users[user.ID] = managed
	if _, ok := repo.accounts[user.ID]; !ok {
		repo.accounts[user.ID] = accountFromUser(managed)
	}

	return managed, nil
}

func (repo *MemoryRepository) UpdateStatus(_ context.Context, userID string, status string) (ManagedUser, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if !validStatus(status) {
		return ManagedUser{}, ErrInvalidStatus
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	user, ok := repo.users[userID]
	if !ok {
		return ManagedUser{}, ErrUserNotFound
	}

	user.Status = status
	if status == "muted" {
		user.ModerationNote = "管理员已限制该用户评论和投稿。"
	}
	if status == "active" {
		user.ModerationNote = ""
	}
	if status == "banned" {
		user.ModerationNote = "账号已封禁，禁止登录和互动。"
	}

	repo.users[userID] = user
	return user, nil
}

func (repo *MemoryRepository) GetAccount(_ context.Context, user auth.User) (AccountSettings, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	settings, ok := repo.accounts[user.ID]
	if !ok {
		managed := ManagedUser{
			ID:            user.ID,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			Role:          user.Role,
			Status:        user.Status,
			AvatarText:    user.AvatarText,
			EmailVerified: true,
			RegisteredAt:  repo.now(),
			LastLoginAt:   repo.now(),
		}
		repo.users[user.ID] = managed
		settings = accountFromUser(managed)
		repo.accounts[user.ID] = settings
	}

	return normalizeAccountSettings(settings, user), nil
}

func (repo *MemoryRepository) UpdateAccount(_ context.Context, user auth.User, settings AccountSettings) (AccountSettings, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	settings = normalizeAccountSettings(settings, user)
	settings.UpdatedAt = repo.now()
	repo.accounts[user.ID] = settings

	managed, ok := repo.users[user.ID]
	if ok {
		managed.DisplayName = settings.DisplayName
		managed.AvatarText = settings.AvatarText
		managed.TwoFactor = settings.TwoFactor
		repo.users[user.ID] = managed
	}

	return settings, nil
}

func normalizeAccountSettings(settings AccountSettings, user auth.User) AccountSettings {
	settings.Email = user.Email
	settings.AvatarText = firstRune(settings.DisplayName)
	settings.TwoFactor = false
	settings.LoginAlert = false
	settings.EmailNotification = false
	settings.ProfileCompleteness = profileCompleteness(settings)
	settings.SecurityLevel = securityLevel(settings)

	return settings
}

func countStats(items []ManagedUser) UserStats {
	stats := UserStats{Total: len(items)}
	for _, user := range items {
		if user.EmailVerified {
			stats.EmailVerified++
		}
		if user.Role == "author" || user.Role == "admin" || user.Role == "editor" {
			stats.Authors++
		}
		if user.Status == "muted" {
			stats.Muted++
		}
		if user.Status == "banned" {
			stats.Banned++
		}
	}

	return stats
}

func validStatus(status string) bool {
	switch status {
	case "active", "muted", "banned", "deleted":
		return true
	default:
		return false
	}
}

func accountFromUser(user ManagedUser) AccountSettings {
	username := strings.Split(user.Email, "@")[0]
	settings := AccountSettings{
		DisplayName:              user.DisplayName,
		Username:                 username,
		Email:                    user.Email,
		AvatarText:               user.AvatarText,
		Bio:                      "关注内容产品、工程实践和长期写作。",
		TwoFactor:                false,
		LoginAlert:               false,
		NotifyReview:             true,
		NotifyComment:            true,
		NotifyAnnouncement:       true,
		EmailNotification:        false,
		PublicProfile:            true,
		PublicBookmarks:          false,
		ProfileURL:               "https://blog.example.com/authors/" + username,
		Timezone:                 "Asia/Shanghai",
		LoginDeviceCount:         2,
		PublicPostCount:          3,
		CurrentDeviceDescription: "Windows Chrome，上海，今天 16:20",
		LastDeviceDescription:    "iPhone Safari，上海，昨天 22:08",
		UpdatedAt:                time.Now(),
	}
	settings.ProfileCompleteness = profileCompleteness(settings)
	settings.SecurityLevel = securityLevel(settings)

	return settings
}

func profileCompleteness(settings AccountSettings) int {
	score := 40
	if strings.TrimSpace(settings.DisplayName) != "" {
		score += 15
	}
	if strings.TrimSpace(settings.Username) != "" {
		score += 15
	}
	if strings.TrimSpace(settings.Bio) != "" {
		score += 15
	}
	if settings.TwoFactor {
		score += 15
	}
	if score > 100 {
		return 100
	}

	return score
}

func securityLevel(settings AccountSettings) string {
	if settings.TwoFactor && settings.LoginAlert {
		return "高"
	}
	if settings.LoginAlert {
		return "中"
	}

	return "低"
}

func firstRune(value string) string {
	for _, item := range strings.TrimSpace(value) {
		return string(item)
	}

	return "用"
}

func seedUsers() map[string]ManagedUser {
	now := time.Now()
	users := []ManagedUser{
		{
			ID:            "user_linyi",
			Email:         "linyi@example.com",
			DisplayName:   "林一",
			Role:          "user",
			Status:        "active",
			AvatarText:    "林",
			EmailVerified: true,
			CommentCount:  42,
			BookmarkCount: 18,
			LastLoginAt:   now.Add(-1 * time.Hour),
			RegisteredAt:  now.AddDate(0, 0, -22),
		},
		{
			ID:            "user_admin",
			Email:         "admin@example.com",
			DisplayName:   "管理员",
			Role:          "admin",
			Status:        "active",
			AvatarText:    "管",
			EmailVerified: true,
			TwoFactor:     false,
			CommentCount:  128,
			BookmarkCount: 6,
			LastLoginAt:   now,
			RegisteredAt:  now.AddDate(0, -2, 0),
		},
		{
			ID:            "user_chen",
			Email:         "chen@example.com",
			DisplayName:   "陈默",
			Role:          "user",
			Status:        "active",
			AvatarText:    "陈",
			EmailVerified: true,
			CommentCount:  16,
			BookmarkCount: 7,
			LastLoginAt:   now.Add(-3 * time.Hour),
			RegisteredAt:  now.AddDate(0, 0, -15),
		},
		{
			ID:             "user_market",
			Email:          "market@example.com",
			DisplayName:    "market_user",
			Role:           "user",
			Status:         "muted",
			AvatarText:     "m",
			EmailVerified:  true,
			CommentCount:   9,
			BookmarkCount:  0,
			LastLoginAt:    now.Add(-12 * time.Minute),
			RegisteredAt:   now.AddDate(0, 0, -3),
			ModerationNote: "推广链接举报 3 次",
		},
		{
			ID:             "user_noise",
			Email:          "noise@example.com",
			DisplayName:    "noise_2048",
			Role:           "user",
			Status:         "banned",
			AvatarText:     "n",
			EmailVerified:  false,
			CommentCount:   21,
			BookmarkCount:  0,
			LastLoginAt:    now.Add(-7 * time.Hour),
			RegisteredAt:   now.AddDate(0, 0, -8),
			ModerationNote: "未验证邮箱",
		},
	}

	result := map[string]ManagedUser{}
	for _, user := range users {
		result[user.ID] = user
	}

	return result
}
