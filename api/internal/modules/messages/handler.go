package messages

import (
	"errors"
	"net/http"

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

	router.GET("/messages", handler.List)
	router.PUT("/messages/read-all", handler.MarkAllRead)
	router.PUT("/messages/:id/read", handler.MarkRead)
	router.PUT("/messages/:id/archive", handler.Archive)

	router.GET("/admin/messages", handler.AdminList)
	router.POST("/admin/messages", handler.AdminCreate)
}

func (handler *Handler) List(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	result, err := handler.repo.List(ctx.Request.Context(), user.ID, ListQuery{
		Status: ctx.Query("status"),
		Type:   ctx.Query("type"),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load messages"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) AdminList(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.AdminList(ctx.Request.Context(), ListQuery{
		Status: ctx.Query("status"),
		Type:   ctx.Query("type"),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load messages"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) AdminCreate(ctx *gin.Context) {
	user, ok := auth.RequireAdmin(ctx)
	if !ok {
		return
	}

	var request CreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid message payload"})
		return
	}

	message, err := handler.repo.Create(ctx.Request.Context(), request, user)
	if err != nil {
		if errors.Is(err, ErrInvalidMessage) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "recipient, title and body are required"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message"})
		return
	}

	ctx.JSON(http.StatusCreated, message)
}

func (handler *Handler) MarkRead(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	message, err := handler.repo.MarkRead(ctx.Request.Context(), user.ID, ctx.Param("id"))
	if err != nil {
		handler.writeMessageError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, message)
}

func (handler *Handler) MarkAllRead(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	stats, err := handler.repo.MarkAllRead(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark messages read"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (handler *Handler) Archive(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	message, err := handler.repo.Archive(ctx.Request.Context(), user.ID, ctx.Param("id"))
	if err != nil {
		handler.writeMessageError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, message)
}

func (handler *Handler) writeMessageError(ctx *gin.Context, err error) {
	if errors.Is(err, ErrMessageNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update message"})
}
