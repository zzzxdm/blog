package operations

import (
	"blog/api/internal/httpx"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"net/http"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) GetPublicSettings(ctx *gin.Context) {
	settings, err := handler.repo.GetSettings(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to load public settings", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	ctx.JSON(http.StatusOK, publicSettings(settings))
}

func (handler *Handler) RunBackup(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.RunBackup(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to run backup", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run backup"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) GetSettings(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	settings, err := handler.repo.GetSettings(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to load admin settings", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (handler *Handler) SendTestMail(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.SendTestMail(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to send test mail", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test mail"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) UpdateSettings(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request Settings
	if !httpx.BindJSON(ctx, &request, "invalid settings payload") {
		return
	}

	settings, err := handler.repo.UpdateSettings(ctx.Request.Context(), request)
	if err != nil {
		slog.Error("failed to update settings", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}

	ctx.JSON(http.StatusOK, settings)
}
