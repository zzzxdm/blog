package operations

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"net/http"
	"time"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) ExportStats(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	stats, err := handler.repo.GetStats(ctx.Request.Context(), ctx.Query("range"))
	if err != nil {
		slog.Error("failed to export stats", "error", err, "range", ctx.Query("range"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export stats"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"scope":      "stats",
		"exportedAt": time.Now(),
		"stats":      stats,
	})
}

func (handler *Handler) GetStats(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	stats, err := handler.repo.GetStats(ctx.Request.Context(), ctx.Query("range"))
	if err != nil {
		slog.Error("failed to load stats", "error", err, "range", ctx.Query("range"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
