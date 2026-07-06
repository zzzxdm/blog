package auth

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	AvatarText    string `json:"avatarText"`
	EmailVerified bool   `json:"emailVerified"`
}

type Credentials struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	TurnstileToken string `json:"turnstileToken"`
}

type RegisterRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	DisplayName    string `json:"displayName"`
	TurnstileToken string `json:"turnstileToken"`
}

type InviteUserRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type PasswordChangeRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type TokenRequest struct {
	Token string `json:"token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type SessionInfo struct {
	ID        string `json:"id"`
	Device    string `json:"device"`
	Current   bool   `json:"current"`
	CreatedAt string `json:"createdAt"`
	ExpiresAt string `json:"expiresAt"`
}

type ExportData struct {
	User          User          `json:"user"`
	Sessions      []SessionInfo `json:"sessions"`
	CommentCount  int           `json:"commentCount"`
	BookmarkCount int           `json:"bookmarkCount"`
	ExportedAt    string        `json:"exportedAt"`
}
