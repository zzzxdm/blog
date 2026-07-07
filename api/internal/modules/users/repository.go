package users

import (
	"context"
	"errors"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidStatus = errors.New("invalid user status")
)

type Repository interface {
	List(ctx context.Context, query ListQuery) (UserListResult, error)
	Get(ctx context.Context, userID string) (ManagedUser, error)
	EnsureFromAuth(ctx context.Context, user auth.User) (ManagedUser, error)
	UpdateStatus(ctx context.Context, userID string, status string) (ManagedUser, error)
	GetAccount(ctx context.Context, user auth.User) (AccountSettings, error)
	UpdateAccount(ctx context.Context, user auth.User, settings AccountSettings) (AccountSettings, error)
}

func normalizeAccountSettings(settings AccountSettings, user auth.User) AccountSettings {
	settings.Email = user.Email
	settings.EmailVerified = user.EmailVerified
	if strings.TrimSpace(settings.AvatarText) == "" {
		settings.AvatarText = firstRune(settings.DisplayName)
	} else {
		settings.AvatarText = firstRune(settings.AvatarText)
	}
	settings.TwoFactor = false
	settings.LoginAlert = false
	settings.NotifyReview = true
	settings.NotifyComment = false
	settings.NotifyAnnouncement = true
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

func matchesUserListQuery(user ManagedUser, query ListQuery) bool {
	status := strings.ToLower(strings.TrimSpace(query.Status))
	if status != "" {
		if status == "unverified" {
			if user.EmailVerified {
				return false
			}
		} else if user.Status != status {
			return false
		}
	}

	role := strings.ToLower(strings.TrimSpace(query.Role))
	if role != "" && user.Role != role {
		return false
	}

	keyword := strings.ToLower(strings.TrimSpace(query.Keyword))
	if keyword == "" {
		return true
	}

	text := strings.ToLower(strings.Join([]string{
		user.ID,
		user.Email,
		user.DisplayName,
		user.Role,
		user.Status,
		user.ModerationNote,
	}, " "))

	return strings.Contains(text, keyword)
}

func normalizeUserPage(page int) int {
	if page < 1 {
		return 1
	}

	return page
}

func normalizeUserPageSize(pageSize int) int {
	if pageSize < 1 {
		return 10
	}
	if pageSize > 100 {
		return 100
	}

	return pageSize
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
		EmailVerified:            user.EmailVerified,
		AvatarText:               user.AvatarText,
		Bio:                      "关注内容产品、工程实践和长期写作。",
		TwoFactor:                false,
		LoginAlert:               false,
		NotifyReview:             true,
		NotifyComment:            false,
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
			ID:            "5001",
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
			ID:            "5002",
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
			ID:            "5003",
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
			ID:             "5004",
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
			ID:             "5005",
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
