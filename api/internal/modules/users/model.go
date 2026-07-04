package users

import "time"

type ManagedUser struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	DisplayName    string    `json:"displayName"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	AvatarText     string    `json:"avatarText"`
	EmailVerified  bool      `json:"emailVerified"`
	TwoFactor      bool      `json:"twoFactor"`
	CommentCount   int       `json:"commentCount"`
	BookmarkCount  int       `json:"bookmarkCount"`
	LastLoginAt    time.Time `json:"lastLoginAt"`
	RegisteredAt   time.Time `json:"registeredAt"`
	ModerationNote string    `json:"moderationNote"`
}

type UserStats struct {
	Total         int `json:"total"`
	EmailVerified int `json:"emailVerified"`
	Authors       int `json:"authors"`
	Muted         int `json:"muted"`
	Banned        int `json:"banned"`
}

type UserListResult struct {
	Items []ManagedUser `json:"items"`
	Total int           `json:"total"`
	Stats UserStats     `json:"stats"`
}

type PasswordResetResult struct {
	OK         bool        `json:"ok"`
	User       ManagedUser `json:"user"`
	ResetToken string      `json:"resetToken,omitempty"`
	Delivery   string      `json:"delivery"`
}

type InvitationResult struct {
	OK         bool        `json:"ok"`
	User       ManagedUser `json:"user"`
	ResetToken string      `json:"resetToken,omitempty"`
	Delivery   string      `json:"delivery"`
}

type StatusRequest struct {
	Status string `json:"status"`
}

type AccountSettings struct {
	DisplayName              string    `json:"displayName"`
	Username                 string    `json:"username"`
	Email                    string    `json:"email"`
	AvatarText               string    `json:"avatarText"`
	Bio                      string    `json:"bio"`
	TwoFactor                bool      `json:"twoFactor"`
	LoginAlert               bool      `json:"loginAlert"`
	NotifyReview             bool      `json:"notifyReview"`
	NotifyComment            bool      `json:"notifyComment"`
	NotifyAnnouncement       bool      `json:"notifyAnnouncement"`
	EmailNotification        bool      `json:"emailNotification"`
	PublicProfile            bool      `json:"publicProfile"`
	PublicBookmarks          bool      `json:"publicBookmarks"`
	ProfileURL               string    `json:"profileUrl"`
	Timezone                 string    `json:"timezone"`
	SecurityLevel            string    `json:"securityLevel"`
	LoginDeviceCount         int       `json:"loginDeviceCount"`
	PublicPostCount          int       `json:"publicPostCount"`
	ProfileCompleteness      int       `json:"profileCompleteness"`
	CurrentDeviceDescription string    `json:"currentDeviceDescription"`
	LastDeviceDescription    string    `json:"lastDeviceDescription"`
	UpdatedAt                time.Time `json:"updatedAt"`
}
