package messages

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

	router.GET("/messages", handler.List)
	router.PUT("/messages/read-all", handler.MarkAllRead)
	router.PUT("/messages/:id/read", handler.MarkRead)
	router.PUT("/messages/:id/archive", handler.Archive)
	router.GET("/me/notifications", handler.List)
	router.GET("/me/messages", handler.List)
	router.GET("/me/messages/:id", handler.GetMine)
	router.POST("/me/messages/read-all", handler.MarkAllRead)
	router.POST("/me/messages/:id/read", handler.MarkRead)
	router.POST("/me/messages/:id/archive", handler.Archive)
	router.DELETE("/me/messages/:id", handler.Archive)

	router.GET("/admin/messages", handler.AdminList)
	router.GET("/admin/messages/export", handler.AdminExport)
	router.POST("/admin/messages", handler.AdminCreate)
	router.POST("/admin/messages/broadcast", handler.AdminBroadcast)
	router.POST("/admin/messages/:id/revoke", handler.AdminRevoke)
	router.GET("/admin/messages/:id/statistics", handler.AdminStatistics)
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

func (handler *Handler) AdminExport(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.AdminList(ctx.Request.Context(), ListQuery{
		Status: ctx.Query("status"),
		Type:   ctx.Query("type"),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export messages"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"scope":      "messages",
		"exportedAt": time.Now(),
		"items":      result.Items,
		"total":      result.Total,
		"stats":      result.Stats,
	})
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

func (handler *Handler) AdminBroadcast(ctx *gin.Context) {
	user, ok := auth.RequireAdmin(ctx)
	if !ok {
		return
	}

	var request BroadcastRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid broadcast payload"})
		return
	}
	if len(request.Recipients) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "broadcast recipients are required"})
		return
	}

	items := make([]Message, 0, len(request.Recipients))
	for _, recipient := range request.Recipients {
		message, err := handler.repo.Create(ctx.Request.Context(), CreateRequest{
			RecipientID:   recipient.ID,
			RecipientName: recipient.Name,
			Type:          request.Type,
			Priority:      request.Priority,
			Title:         request.Title,
			Body:          request.Body,
			TargetType:    request.TargetType,
			TargetID:      request.TargetID,
			TargetTitle:   request.TargetTitle,
			ScheduledAt:   request.ScheduledAt,
		}, user)
		if err != nil {
			if errors.Is(err, ErrInvalidMessage) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "recipients, title and body are required"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast message"})
			return
		}
		items = append(items, message)
	}

	ctx.JSON(http.StatusCreated, gin.H{"items": items, "total": len(items)})
}

func (handler *Handler) AdminRevoke(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	message, ok := handler.adminMessageByID(ctx, ctx.Param("id"))
	if !ok {
		return
	}

	archived, err := handler.repo.Archive(ctx.Request.Context(), message.RecipientID, message.ID)
	if err != nil {
		handler.writeMessageError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true, "message": archived})
}

func (handler *Handler) AdminStatistics(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	message, ok := handler.adminMessageByID(ctx, ctx.Param("id"))
	if !ok {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":        message.ID,
		"status":    message.Status,
		"read":      message.ReadAt != nil,
		"archived":  message.ArchivedAt != nil,
		"scheduled": message.ScheduledAt != nil && message.ScheduledAt.After(time.Now()),
		"recipient": gin.H{"id": message.RecipientID, "name": message.RecipientName},
	})
}

func (handler *Handler) adminMessageByID(ctx *gin.Context, id string) (Message, bool) {
	result, err := handler.repo.AdminList(ctx.Request.Context(), ListQuery{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load message"})
		return Message{}, false
	}
	for _, message := range result.Items {
		if message.ID == id {
			return message, true
		}
	}

	ctx.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
	return Message{}, false
}

func (handler *Handler) GetMine(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	result, err := handler.repo.List(ctx.Request.Context(), user.ID, ListQuery{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load message"})
		return
	}
	for _, message := range result.Items {
		if message.ID == ctx.Param("id") {
			ctx.JSON(http.StatusOK, message)
			return
		}
	}

	ctx.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
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
