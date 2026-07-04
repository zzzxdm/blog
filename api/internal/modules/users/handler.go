package users

import (
	"errors"
	"net/http"
	"time"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func RegisterRoutes(router gin.IRouter, repo Repository) {
	handler := NewHandler(repo)

	router.GET("/admin/users", handler.List)
	router.GET("/admin/users/export", handler.Export)
	router.PUT("/admin/users/:id/status", handler.UpdateStatus)
	router.GET("/account/settings", handler.GetAccount)
	router.PUT("/account/settings", handler.UpdateAccount)
}

func (handler *Handler) List(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) Export(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.List(ctx.Request.Context())
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

func (handler *Handler) UpdateStatus(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request StatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user status payload"})
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

	ctx.JSON(http.StatusOK, settings)
}
