package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName  = "blog_session"
	currentUserKey     = "currentUser"
	defaultSessionDays = 7
	minSessionDays     = 1
	maxSessionDays     = 90
	loginFailureLimit  = 5
	loginLockDuration  = 15 * time.Minute
)

type Handler struct {
	store         Store
	settings      SecuritySettingsReader
	loginFailures map[string]loginFailure
	loginMu       sync.Mutex
}

type SecuritySettings struct {
	SessionDays      int
	LoginFailureLock bool
}

type SecuritySettingsReader interface {
	SecuritySettings(ctx context.Context) (SecuritySettings, error)
}

type loginFailure struct {
	Count       int
	LockedUntil time.Time
}

func NewHandler(store Store) *Handler {
	return NewHandlerWithSettings(store, nil)
}

func NewHandlerWithSettings(store Store, settings SecuritySettingsReader) *Handler {
	return &Handler{
		store:         store,
		settings:      settings,
		loginFailures: map[string]loginFailure{},
	}
}

func RegisterRoutes(router gin.IRouter, store Store) {
	RegisterRoutesWithSettings(router, store, nil)
}

func RegisterRoutesWithSettings(router gin.IRouter, store Store, settings SecuritySettingsReader) {
	handler := NewHandlerWithSettings(store, settings)

	router.POST("/auth/login", handler.Login)
	router.POST("/auth/register", handler.Register)
	router.POST("/auth/logout", handler.Logout)
	router.POST("/auth/email-verification", handler.RequestEmailVerification)
	router.POST("/auth/verify-email", handler.VerifyEmail)
	router.POST("/auth/forgot-password", handler.ForgotPassword)
	router.POST("/auth/reset-password", handler.ResetPassword)
	router.GET("/me", handler.Me)
	router.GET("/me/sessions", handler.Sessions)
	router.DELETE("/me/sessions/:id", handler.DeleteSession)
	router.POST("/me/export", handler.ExportMe)
	router.DELETE("/me", handler.DeleteMe)
	router.PUT("/me/password", handler.ChangePassword)
}

func Middleware(store Store) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie(SessionCookieName)
		if err != nil || token == "" {
			ctx.Next()
			return
		}

		user, err := store.UserBySession(token)
		if err == nil {
			ctx.Set(currentUserKey, user)
		}

		ctx.Next()
	}
}

func CurrentUser(ctx *gin.Context) (User, bool) {
	value, ok := ctx.Get(currentUserKey)
	if !ok {
		return User{}, false
	}

	user, ok := value.(User)
	return user, ok
}

func RequireUser(ctx *gin.Context) (User, bool) {
	user, ok := CurrentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "login required"})
		return User{}, false
	}

	return user, true
}

func RequireAdmin(ctx *gin.Context) (User, bool) {
	user, ok := RequireUser(ctx)
	if !ok {
		return User{}, false
	}

	if user.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin required"})
		return User{}, false
	}

	return user, true
}

func (handler *Handler) Login(ctx *gin.Context) {
	var request Credentials
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid login payload"})
		return
	}

	security := handler.configuredSecuritySettings(ctx)
	if security.LoginFailureLock && handler.isLoginLocked(request.Email, time.Now()) {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "account temporarily locked"})
		return
	}

	user, token, err := handler.store.Authenticate(request.Email, request.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			if security.LoginFailureLock && handler.recordLoginFailure(request.Email, time.Now()) {
				ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "account temporarily locked"})
				return
			}

			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	handler.clearLoginFailure(request.Email)
	sessionDays := clampSessionDays(security.SessionDays)
	if err := handler.store.SetSessionExpiry(token, sessionExpiry(sessionDays)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session expiry"})
		return
	}

	setSessionCookie(ctx, token, sessionDays)
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (handler *Handler) Register(ctx *gin.Context) {
	var request RegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid register payload"})
		return
	}

	if strings.TrimSpace(request.Email) == "" || len(request.Password) < 6 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email and at least 6-character password are required"})
		return
	}

	user, token, err := handler.store.Register(request)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "register failed"})
		return
	}

	sessionDays := clampSessionDays(handler.configuredSecuritySettings(ctx).SessionDays)
	if err := handler.store.SetSessionExpiry(token, sessionExpiry(sessionDays)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session expiry"})
		return
	}

	setSessionCookie(ctx, token, sessionDays)

	verificationToken, err := handler.store.RequestEmailVerification(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create email verification"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"user":              user,
		"verificationToken": verificationToken,
		"delivery":          "dev-response",
	})
}

func (handler *Handler) Logout(ctx *gin.Context) {
	token, err := ctx.Cookie(SessionCookieName)
	if err == nil && token != "" {
		handler.store.DeleteSession(token)
	}

	clearSessionCookie(ctx)
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (handler *Handler) Me(ctx *gin.Context) {
	user, ok := CurrentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "login required"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (handler *Handler) Sessions(ctx *gin.Context) {
	user, ok := RequireUser(ctx)
	if !ok {
		return
	}

	sessions, err := handler.store.ListSessions(user.ID, currentSessionToken(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load sessions"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items": sessions,
		"total": len(sessions),
	})
}

func (handler *Handler) DeleteSession(ctx *gin.Context) {
	user, ok := RequireUser(ctx)
	if !ok {
		return
	}

	sessionID := ctx.Param("id")
	if err := handler.store.DeleteUserSession(user.ID, sessionID); err != nil {
		if errors.Is(err, ErrInvalidSession) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete session"})
		return
	}

	if sessionID == currentSessionToken(ctx) {
		clearSessionCookie(ctx)
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (handler *Handler) ExportMe(ctx *gin.Context) {
	user, ok := RequireUser(ctx)
	if !ok {
		return
	}

	data, err := handler.store.ExportUserData(user.ID, currentSessionToken(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export account data"})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func (handler *Handler) DeleteMe(ctx *gin.Context) {
	user, ok := RequireUser(ctx)
	if !ok {
		return
	}

	if err := handler.store.DeleteUser(user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account"})
		return
	}

	clearSessionCookie(ctx)
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (handler *Handler) ChangePassword(ctx *gin.Context) {
	user, ok := RequireUser(ctx)
	if !ok {
		return
	}

	var request PasswordChangeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid password payload"})
		return
	}

	if len(request.NewPassword) < 6 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 6 characters"})
		return
	}

	if err := handler.store.ChangePassword(user.ID, request.CurrentPassword, request.NewPassword); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to change password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (handler *Handler) RequestEmailVerification(ctx *gin.Context) {
	user, ok := RequireUser(ctx)
	if !ok {
		return
	}

	token, err := handler.store.RequestEmailVerification(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create email verification"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"ok":                true,
		"verificationToken": token,
		"delivery":          "dev-response",
	})
}

func (handler *Handler) VerifyEmail(ctx *gin.Context) {
	var request TokenRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid verification payload"})
		return
	}

	user, err := handler.store.VerifyEmail(request.Token)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "verification token is invalid or expired"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify email"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (handler *Handler) ForgotPassword(ctx *gin.Context) {
	var request ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid forgot password payload"})
		return
	}

	token, err := handler.store.RequestPasswordReset(request.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create password reset"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"resetToken": token,
		"delivery":   "dev-response",
	})
}

func (handler *Handler) ResetPassword(ctx *gin.Context) {
	var request ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid reset password payload"})
		return
	}

	if len(request.NewPassword) < 6 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 6 characters"})
		return
	}

	if err := handler.store.ResetPassword(request.Token, request.NewPassword); err != nil {
		if errors.Is(err, ErrInvalidToken) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "reset token is invalid or expired"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (handler *Handler) configuredSecuritySettings(ctx *gin.Context) SecuritySettings {
	if handler.settings == nil {
		return SecuritySettings{SessionDays: defaultSessionDays}
	}

	settings, err := handler.settings.SecuritySettings(ctx.Request.Context())
	if err != nil {
		return SecuritySettings{SessionDays: defaultSessionDays}
	}

	settings.SessionDays = clampSessionDays(settings.SessionDays)
	return settings
}

func (handler *Handler) isLoginLocked(email string, now time.Time) bool {
	key := normalizeEmail(email)
	if key == "" {
		return false
	}

	handler.loginMu.Lock()
	defer handler.loginMu.Unlock()

	failure := handler.loginFailures[key]
	if failure.LockedUntil.IsZero() {
		return false
	}
	if failure.LockedUntil.After(now) {
		return true
	}

	delete(handler.loginFailures, key)
	return false
}

func (handler *Handler) recordLoginFailure(email string, now time.Time) bool {
	key := normalizeEmail(email)
	if key == "" {
		return false
	}

	handler.loginMu.Lock()
	defer handler.loginMu.Unlock()

	failure := handler.loginFailures[key]
	failure.Count++
	if failure.Count >= loginFailureLimit {
		failure.LockedUntil = now.Add(loginLockDuration)
	}
	handler.loginFailures[key] = failure

	return !failure.LockedUntil.IsZero() && failure.LockedUntil.After(now)
}

func (handler *Handler) clearLoginFailure(email string) {
	key := normalizeEmail(email)
	if key == "" {
		return
	}

	handler.loginMu.Lock()
	delete(handler.loginFailures, key)
	handler.loginMu.Unlock()
}

func clampSessionDays(days int) int {
	if days < minSessionDays {
		return defaultSessionDays
	}
	if days > maxSessionDays {
		return maxSessionDays
	}

	return days
}

func sessionExpiry(days int) time.Time {
	return time.Now().Add(time.Duration(clampSessionDays(days)) * 24 * time.Hour)
}

func setSessionCookie(ctx *gin.Context, token string, days int) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(SessionCookieName, token, clampSessionDays(days)*24*60*60, "/", "", false, true)
}

func clearSessionCookie(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
}

func currentSessionToken(ctx *gin.Context) string {
	token, err := ctx.Cookie(SessionCookieName)
	if err != nil {
		return ""
	}

	return token
}
