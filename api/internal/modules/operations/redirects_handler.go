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

func (handler *Handler) CreateRedirect(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request RedirectRule
	if !httpx.BindJSON(ctx, &request, "invalid redirect payload") {
		return
	}

	navigation, err := handler.repo.GetNavigation(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to load redirects before create", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load redirects"})
		return
	}

	updatedRedirects := normalizeRedirects(append(navigation.Redirects, request))
	if len(updatedRedirects) == len(navigation.Redirects) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "redirect from and to are required"})
		return
	}

	navigation.Redirects = updatedRedirects
	navigation, err = handler.repo.UpdateNavigation(ctx.Request.Context(), navigation)
	if err != nil {
		slog.Error("failed to save redirect", "error", err, "from", request.From, "to", request.To)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save redirect"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"item":  navigation.Redirects[len(navigation.Redirects)-1],
		"items": navigation.Redirects,
		"total": len(navigation.Redirects),
	})
}

func (handler *Handler) ReplaceRedirects(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request struct {
		Items []RedirectRule `json:"items"`
	}
	if !httpx.BindJSON(ctx, &request, "invalid redirects payload") {
		return
	}

	navigation, err := handler.repo.GetNavigation(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to load redirects before replace", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load redirects"})
		return
	}

	navigation.Redirects = normalizeRedirects(request.Items)
	navigation, err = handler.repo.UpdateNavigation(ctx.Request.Context(), navigation)
	if err != nil {
		slog.Error("failed to save redirects", "error", err, "count", len(request.Items))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save redirects"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items": navigation.Redirects,
		"total": len(navigation.Redirects),
	})
}

func (handler *Handler) ListRedirects(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	navigation, err := handler.repo.GetNavigation(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to load redirects", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load redirects"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items": navigation.Redirects,
		"total": len(navigation.Redirects),
	})
}
