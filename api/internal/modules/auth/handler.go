package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "blog_session"
	currentUserKey    = "currentUser"
)

type Handler struct {
	store *MemoryStore
}

func NewHandler(store *MemoryStore) *Handler {
	return &Handler{store: store}
}

func RegisterRoutes(router gin.IRouter, store *MemoryStore) {
	handler := NewHandler(store)

	router.POST("/auth/login", handler.Login)
	router.POST("/auth/register", handler.Register)
	router.POST("/auth/logout", handler.Logout)
	router.GET("/me", handler.Me)
}

func Middleware(store *MemoryStore) gin.HandlerFunc {
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

func (handler *Handler) Login(ctx *gin.Context) {
	var request Credentials
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid login payload"})
		return
	}

	user, token, err := handler.store.Authenticate(request.Email, request.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	setSessionCookie(ctx, token)
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

	setSessionCookie(ctx, token)
	ctx.JSON(http.StatusCreated, gin.H{"user": user})
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

func setSessionCookie(ctx *gin.Context, token string) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(SessionCookieName, token, 7*24*60*60, "/", "", false, true)
}

func clearSessionCookie(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
}
