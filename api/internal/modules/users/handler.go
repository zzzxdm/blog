package users

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo        Repository
	authStore   auth.Store
	emailSender auth.EmailSender
}

func NewHandler(repo Repository, authStore auth.Store) *Handler {
	return NewHandlerWithEmailSender(repo, authStore, nil)
}

func NewHandlerWithEmailSender(repo Repository, authStore auth.Store, emailSender auth.EmailSender) *Handler {
	return &Handler{repo: repo, authStore: authStore, emailSender: emailSender}
}

func RegisterRoutes(router gin.IRouter, repo Repository, authStore auth.Store) {
	RegisterRoutesWithEmailSender(router, repo, authStore, nil)
}

func RegisterRoutesWithEmailSender(router gin.IRouter, repo Repository, authStore auth.Store, emailSender auth.EmailSender) {
	handler := NewHandlerWithEmailSender(repo, authStore, emailSender)

	router.GET("/admin/users", handler.List)
	router.GET("/admin/users/export", handler.Export)
	router.GET("/admin/users/:id", handler.Get)
	router.GET("/admin/users/:id/sessions", handler.ListSessions)
	router.POST("/admin/users/invitations", handler.Invite)
	router.PUT("/admin/users/:id/role", handler.UpdateRole)
	router.PUT("/admin/users/:id/status", handler.UpdateStatus)
	router.DELETE("/admin/users/:id", handler.Delete)
	router.POST("/admin/users/:id/restore", handler.Restore)
	router.POST("/admin/users/:id/password-reset", handler.RequestPasswordReset)
	router.GET("/account/settings", handler.GetAccount)
	router.PUT("/account/settings", handler.UpdateAccount)
	router.PUT("/me/profile", handler.UpdateAccount)
	router.POST("/me/avatar", handler.UpdateAvatar)
}

func (handler *Handler) List(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.List(ctx.Request.Context(), ListQuery{
		Page:     intQuery(ctx, "page", 1),
		PageSize: intQuery(ctx, "pageSize", 10),
		Keyword:  strings.TrimSpace(ctx.Query("q")),
		Status:   strings.TrimSpace(ctx.Query("status")),
		Role:     strings.TrimSpace(ctx.Query("role")),
		All:      boolQuery(ctx, "all"),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) Get(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	user, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (handler *Handler) ListSessions(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}
	if handler.authStore == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "session management is unavailable"})
		return
	}

	sessions, err := handler.authStore.ListSessions(ctx.Param("id"), "")
	if err != nil {
		if errors.Is(err, auth.ErrInvalidSession) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user sessions"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items": sessions,
		"total": len(sessions),
	})
}

func (handler *Handler) Export(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.List(ctx.Request.Context(), ListQuery{All: true})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export users"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"scope":      "users",
		"exportedAt": time.Now(),
		"items":      result.Items,
		"total":      result.Total,
		"stats":      result.Stats,
	})
}

func (handler *Handler) Invite(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}
	if handler.authStore == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user invitation is unavailable"})
		return
	}

	var request auth.InviteUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation payload"})
		return
	}
	if strings.TrimSpace(request.Email) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	invited, secrets, err := handler.authStore.InviteUser(request)
	if err != nil {
		if errors.Is(err, auth.ErrEmailExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "该邮箱已存在，不能重复邀请。请在用户列表中搜索该邮箱，可直接调整角色或发送密码重置邮件。"})
			return
		}
		if errors.Is(err, auth.ErrAccountDeleted) {
			ctx.JSON(http.StatusGone, gin.H{"error": "该邮箱对应的账号已被删除，不能直接重新邀请。请确认是否需要恢复账号或换用其他邮箱。"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to invite user"})
		return
	}

	managed, err := handler.repo.EnsureFromAuth(ctx.Request.Context(), invited)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync invited user"})
		return
	}

	response := InvitationResult{
		OK:              true,
		User:            managed,
		InitialPassword: secrets.InitialPassword,
		ResetToken:      secrets.ResetToken,
		Delivery:        "dev-response",
	}
	if handler.emailSender != nil {
		if err := handler.emailSender.SendInvitation(ctx.Request.Context(), invited, secrets.InitialPassword, secrets.ResetToken); err == nil {
			response.InitialPassword = ""
			response.ResetToken = ""
			response.Delivery = "email"
		} else {
			response.Delivery = "email-failed"
		}
	}

	ctx.JSON(http.StatusCreated, response)
}

func (handler *Handler) UpdateStatus(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request StatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user status payload"})
		return
	}

	if handler.authStore != nil {
		updated, err := handler.authStore.UpdateStatus(ctx.Param("id"), request.Status)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidSession) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			if errors.Is(err, auth.ErrInvalidStatus) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user status"})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync user status"})
			return
		}

		managed, err := handler.repo.EnsureFromAuth(ctx.Request.Context(), updated)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync user status"})
			return
		}

		ctx.JSON(http.StatusOK, managed)
		return
	}

	user, err := handler.repo.UpdateStatus(ctx.Request.Context(), ctx.Param("id"), request.Status)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, ErrInvalidStatus) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user status"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (handler *Handler) UpdateRole(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}
	if handler.authStore == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "role update is unavailable"})
		return
	}

	var request RoleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user role payload"})
		return
	}

	updated, err := handler.authStore.UpdateRole(ctx.Param("id"), request.Role)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidSession) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, auth.ErrInvalidRole) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user role"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role"})
		return
	}

	managed, err := handler.repo.EnsureFromAuth(ctx.Request.Context(), updated)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync user role"})
		return
	}

	ctx.JSON(http.StatusOK, managed)
}

func (handler *Handler) Delete(ctx *gin.Context) {
	admin, ok := auth.RequireAdmin(ctx)
	if !ok {
		return
	}

	userID := ctx.Param("id")
	if userID == admin.ID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete current admin"})
		return
	}

	if handler.authStore != nil {
		if err := handler.authStore.DeleteUser(userID); err != nil {
			if errors.Is(err, auth.ErrInvalidSession) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
			return
		}
	}

	managed, err := handler.repo.UpdateStatus(ctx.Request.Context(), userID, "deleted")
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, ErrInvalidStatus) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user status"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync deleted user"})
		return
	}

	ctx.JSON(http.StatusOK, managed)
}

func (handler *Handler) Restore(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	userID := ctx.Param("id")
	if handler.authStore != nil {
		updated, err := handler.authStore.UpdateStatus(userID, "active")
		if err != nil {
			if errors.Is(err, auth.ErrInvalidSession) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			if errors.Is(err, auth.ErrInvalidStatus) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user status"})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore user"})
			return
		}

		managed, err := handler.repo.EnsureFromAuth(ctx.Request.Context(), updated)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync restored user"})
			return
		}

		ctx.JSON(http.StatusOK, managed)
		return
	}

	managed, err := handler.repo.UpdateStatus(ctx.Request.Context(), userID, "active")
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, ErrInvalidStatus) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user status"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore user"})
		return
	}

	ctx.JSON(http.StatusOK, managed)
}

func (handler *Handler) RequestPasswordReset(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}
	if handler.authStore == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "password reset is unavailable"})
		return
	}

	user, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}

	_, token, err := handler.authStore.RequestPasswordReset(user.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create password reset"})
		return
	}

	response := PasswordResetResult{
		OK:         true,
		User:       user,
		ResetToken: token,
		Delivery:   "dev-response",
	}
	if handler.emailSender != nil {
		authUser := auth.User{
			ID:            user.ID,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			Role:          user.Role,
			Status:        user.Status,
			AvatarText:    user.AvatarText,
			EmailVerified: user.EmailVerified,
		}
		if err := handler.emailSender.SendPasswordSetup(ctx.Request.Context(), authUser, token); err == nil {
			response.ResetToken = ""
			response.Delivery = "email"
		} else {
			response.Delivery = "email-failed"
		}
	}

	ctx.JSON(http.StatusOK, response)
}

func (handler *Handler) GetAccount(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	settings, err := handler.repo.GetAccount(ctx.Request.Context(), user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load account settings"})
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (handler *Handler) UpdateAccount(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	var request AccountSettings
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid account settings payload"})
		return
	}

	settings, err := handler.repo.UpdateAccount(ctx.Request.Context(), user, request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update account settings"})
		return
	}
	if !handler.syncAccountProfile(ctx, user.ID, settings) {
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (handler *Handler) UpdateAvatar(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	settings, err := handler.repo.GetAccount(ctx.Request.Context(), user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load account settings"})
		return
	}

	avatarText := strings.TrimSpace(ctx.PostForm("avatarText"))
	if avatarText == "" && strings.HasPrefix(ctx.ContentType(), "application/json") {
		var request AvatarRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid avatar payload"})
			return
		}
		avatarText = strings.TrimSpace(request.AvatarText)
	}
	if avatarText == "" {
		if file, err := ctx.FormFile("file"); err == nil {
			avatarText = firstRune(file.Filename)
		}
	}
	if avatarText == "" {
		avatarText = settings.AvatarText
	}

	settings.AvatarText = avatarText
	settings, err = handler.repo.UpdateAccount(ctx.Request.Context(), user, settings)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update avatar"})
		return
	}
	if !handler.syncAccountProfile(ctx, user.ID, settings) {
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func intQuery(ctx *gin.Context, key string, fallback int) int {
	value := ctx.Query(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func boolQuery(ctx *gin.Context, key string) bool {
	value := strings.ToLower(strings.TrimSpace(ctx.Query(key)))
	return value == "1" || value == "true" || value == "yes"
}

func (handler *Handler) syncAccountProfile(ctx *gin.Context, userID string, settings AccountSettings) bool {
	if handler.authStore == nil {
		return true
	}
	if _, err := handler.authStore.UpdateProfile(userID, settings.DisplayName, settings.AvatarText); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync account profile"})
		return false
	}

	return true
}
