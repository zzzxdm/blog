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

func (handler *Handler) GetPublicNavigation(ctx *gin.Context) {
	handler.writeNavigation(ctx)
}

func (handler *Handler) GetNavigation(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	handler.writeNavigation(ctx)
}

func (handler *Handler) UpdateNavigation(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request Navigation
	if !httpx.BindJSON(ctx, &request, "invalid navigation payload") {
		return
	}

	navigation, err := handler.repo.UpdateNavigation(ctx.Request.Context(), request)
	if err != nil {
		slog.Error("failed to update navigation", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update navigation"})
		return
	}

	ctx.JSON(http.StatusOK, navigation)
}

func (handler *Handler) writeNavigation(ctx *gin.Context) {
	navigation, err := handler.repo.GetNavigation(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to load navigation", "error", err, "path", ctx.FullPath())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load navigation"})
		return
	}

	ctx.JSON(http.StatusOK, navigation)
}
