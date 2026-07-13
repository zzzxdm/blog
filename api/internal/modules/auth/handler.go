package auth

import (
	"blog/api/internal/httpx"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var cookieSecureFlag bool

// ConfigureCookieSecurity controls the Secure flag for session cookies.
func ConfigureCookieSecurity(secure bool) {
	cookieSecureFlag = secure
}

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
	store             Store
	settings          SecuritySettingsReader
	emailSender       EmailSender
	turnstileVerifier TurnstileVerifier
	loginFailures     map[string]loginFailure
	loginMu           sync.Mutex
}

type SecuritySettings struct {
	SessionDays         int
	LoginFailureLock    bool
	TurnstileEnabled    bool
	TurnstileSiteKey    string
	TurnstileSecretKey  string
	TurnstileRegister   bool
	TurnstileLogin      bool
	TurnstileSubmission bool
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
	return NewHandlerWithSettingsAndEmailSender(store, settings, nil)
}

func NewHandlerWithSettingsAndEmailSender(store Store, settings SecuritySettingsReader, emailSender EmailSender) *Handler {
	return NewHandlerWithDependencies(store, settings, emailSender, nil)
}

func NewHandlerWithDependencies(store Store, settings SecuritySettingsReader, emailSender EmailSender, turnstileVerifier TurnstileVerifier) *Handler {
	if turnstileVerifier == nil {
		turnstileVerifier = NewHTTPTurnstileVerifier()
	}

	return &Handler{
		store:             store,
		settings:          settings,
		emailSender:       emailSender,
		turnstileVerifier: turnstileVerifier,
		loginFailures:     map[string]loginFailure{},
	}
}

func RegisterRoutes(router gin.IRouter, store Store) {
	RegisterRoutesWithSettings(router, store, nil)
}

func RegisterRoutesWithSettings(router gin.IRouter, store Store, settings SecuritySettingsReader) {
	RegisterRoutesWithSettingsAndEmailSender(router, store, settings, nil)
}

func RegisterRoutesWithSettingsAndEmailSender(router gin.IRouter, store Store, settings SecuritySettingsReader, emailSender EmailSender) {
	RegisterRoutesWithDependencies(router, store, settings, emailSender, nil)
}

func RegisterRoutesWithDependencies(router gin.IRouter, store Store, settings SecuritySettingsReader, emailSender EmailSender, turnstileVerifier TurnstileVerifier) {
	handler := NewHandlerWithDependencies(store, settings, emailSender, turnstileVerifier)

	router.POST("/auth/login", handler.Login)
	router.POST("/auth/register", handler.Register)
	router.POST("/auth/logout", handler.Logout)
	router.POST("/auth/email-verification", handler.RequestEmailVerification)
	router.POST("/auth/verify-email", handler.VerifyEmail)
	router.POST("/auth/forgot-password", handler.ForgotPassword)
	router.POST("/auth/reset-password", handler.ResetPassword)
	router.GET("/auth/me", handler.Me)
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
	if !httpx.BindJSON(ctx, &request, "invalid login payload") {
		return
	}

	security := handler.configuredSecuritySettings(ctx)
	if !handler.verifyTurnstile(ctx, security, request.TurnstileToken, security.TurnstileLogin) {
		return
	}
	if security.LoginFailureLock && handler.isLoginLocked(request.Email, time.Now()) {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "account temporarily locked"})
		return
	}

	user, token, err := handler.store.Authenticate(request.Email, request.Password)
	if err != nil {
		if errors.Is(err, ErrAccountDeleted) {
			ctx.JSON(http.StatusGone, gin.H{"error": "account has been deleted"})
			return
		}
		if errors.Is(err, ErrAccountBanned) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "account has been banned"})
			return
		}
		if errors.Is(err, ErrInvalidCredentials) {
			if security.LoginFailureLock && handler.recordLoginFailure(request.Email, time.Now()) {
				ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "account temporarily locked"})
				return
			}

			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		slog.Error("login failed", "error", err, "email", normalizeEmail(request.Email))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	handler.clearLoginFailure(request.Email)
	sessionDays := clampSessionDays(security.SessionDays)
	if err := handler.store.SetSessionExpiry(token, sessionExpiry(sessionDays)); err != nil {
		slog.Error("failed to update login session expiry", "error", err, "userID", user.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session expiry"})
		return
	}

	setSessionCookie(ctx, token, sessionDays)
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (handler *Handler) Register(ctx *gin.Context) {
	var request RegisterRequest
	if !httpx.BindJSON(ctx, &request, "invalid register payload") {
		return
	}

	if strings.TrimSpace(request.Email) == "" || len(request.Password) < 6 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email and at least 6-character password are required"})
		return
	}

	security := handler.configuredSecuritySettings(ctx)
	if !handler.verifyTurnstile(ctx, security, request.TurnstileToken, security.TurnstileRegister) {
		return
	}

	user, token, err := handler.store.Register(request)
	if err != nil {
		if errors.Is(err, ErrAccountDeleted) {
			ctx.JSON(http.StatusGone, gin.H{"error": "account has been deleted"})
			return
		}
		if errors.Is(err, ErrEmailExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}

		slog.Error("register failed", "error", err, "email", normalizeEmail(request.Email))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "register failed"})
		return
	}

	sessionDays := clampSessionDays(security.SessionDays)
	if err := handler.store.SetSessionExpiry(token, sessionExpiry(sessionDays)); err != nil {
		slog.Error("failed to update register session expiry", "error", err, "userID", user.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session expiry"})
		return
	}

	setSessionCookie(ctx, token, sessionDays)

	verificationToken, err := handler.store.RequestEmailVerification(user.ID)
	if err != nil {
		slog.Error("failed to create email verification after register", "error", err, "userID", user.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create email verification"})
		return
	}

	response, ok := handler.deliverEmailVerification(ctx, user, verificationToken, true)
	if !ok {
		return
	}
	response["user"] = user
	ctx.JSON(http.StatusCreated, response)
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
		slog.Error("failed to load sessions", "error", err, "userID", user.ID)
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

		slog.Error("failed to delete session", "error", err, "userID", user.ID, "sessionID", sessionID)
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
		slog.Error("failed to export account data", "error", err, "userID", user.ID)
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
		slog.Error("failed to delete account", "error", err, "userID", user.ID)
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
	if !httpx.BindJSON(ctx, &request, "invalid password payload") {
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

		slog.Error("failed to change password", "error", err, "userID", user.ID)
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
		slog.Error("failed to create email verification", "error", err, "userID", user.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create email verification"})
		return
	}

	response, ok := handler.deliverEmailVerification(ctx, user, token, false)
	if !ok {
		return
	}
	response["ok"] = true
	ctx.JSON(http.StatusOK, response)
}

func (handler *Handler) VerifyEmail(ctx *gin.Context) {
	var request TokenRequest
	if !httpx.BindJSON(ctx, &request, "invalid verification payload") {
		return
	}

	user, err := handler.store.VerifyEmail(request.Token)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "verification token is invalid or expired"})
			return
		}

		slog.Error("failed to verify email", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify email"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (handler *Handler) deliverEmailVerification(ctx *gin.Context, user User, token string, failSoft bool) (gin.H, bool) {
	if handler.emailSender == nil {
		return gin.H{
			"verificationToken": token,
			"delivery":          "dev-response",
		}, true
	}

	if err := handler.emailSender.SendEmailVerification(ctx.Request.Context(), user, token); err != nil {
		slog.Error("failed to send email verification", "error", err, "userID", user.ID, "email", normalizeEmail(user.Email), "failSoft", failSoft)
		if failSoft {
			return gin.H{
				"delivery": "email-failed",
				"warning":  "failed to send email verification",
			}, true
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email verification"})
		return nil, false
	}

	return gin.H{"delivery": "email"}, true
}

func (handler *Handler) verifyTurnstile(ctx *gin.Context, settings SecuritySettings, token string, required bool) bool {
	if !settings.TurnstileEnabled || !required {
		return true
	}
	if strings.TrimSpace(settings.TurnstileSecretKey) == "" {
		slog.Warn("turnstile verification is enabled but secret key is missing", "path", ctx.FullPath(), "ip", ctx.ClientIP())
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "turnstile is not configured"})
		return false
	}
	if strings.TrimSpace(token) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "turnstile token is required"})
		return false
	}

	ok, err := handler.turnstileVerifier.Verify(ctx.Request.Context(), settings.TurnstileSecretKey, token, ctx.ClientIP())
	if err != nil {
		slog.Error("turnstile verification failed", "error", err, "ip", ctx.ClientIP())
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "turnstile verification failed"})
		return false
	}
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "turnstile token is invalid"})
		return false
	}

	return true
}

func (handler *Handler) ForgotPassword(ctx *gin.Context) {
	var request ForgotPasswordRequest
	if !httpx.BindJSON(ctx, &request, "invalid forgot password payload") {
		return
	}

	user, token, err := handler.store.RequestPasswordReset(request.Email)
	if err != nil {
		slog.Error("failed to create password reset", "error", err, "email", normalizeEmail(request.Email))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create password reset"})
		return
	}
	if token == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "email is not registered"})
		return
	}

	response, ok := handler.deliverPasswordReset(ctx, user, token)
	if !ok {
		return
	}
	response["ok"] = true
	ctx.JSON(http.StatusOK, response)
}

func (handler *Handler) deliverPasswordReset(ctx *gin.Context, user User, token string) (gin.H, bool) {
	if handler.emailSender == nil {
		response := gin.H{"delivery": "dev-response"}
		if token != "" {
			response["resetToken"] = token
		}
		return response, true
	}

	if token == "" {
		return gin.H{"delivery": "email"}, true
	}

	if err := handler.emailSender.SendPasswordSetup(ctx.Request.Context(), user, token); err != nil {
		slog.Error("failed to send password reset email", "error", err, "userID", user.ID, "email", normalizeEmail(user.Email))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send password reset email"})
		return nil, false
	}

	return gin.H{"delivery": "email"}, true
}

func (handler *Handler) ResetPassword(ctx *gin.Context) {
	var request ResetPasswordRequest
	if !httpx.BindJSON(ctx, &request, "invalid reset password payload") {
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

		slog.Error("failed to reset password", "error", err)
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
		slog.Warn("failed to load auth security settings", "error", err)
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
	ctx.SetCookie(SessionCookieName, token, clampSessionDays(days)*24*60*60, "/", "", cookieSecureFlag, true)
}

func clearSessionCookie(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(SessionCookieName, "", -1, "/", "", cookieSecureFlag, true)
}

func currentSessionToken(ctx *gin.Context) string {
	token, err := ctx.Cookie(SessionCookieName)
	if err != nil {
		return ""
	}

	return token
}

