package operations

import (
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

	router.GET("/admin/settings", handler.GetSettings)
	router.PUT("/admin/settings", handler.UpdateSettings)
	router.GET("/admin/navigation", handler.GetNavigation)
	router.PUT("/admin/navigation", handler.UpdateNavigation)
	router.GET("/admin/media", handler.ListMedia)
	router.GET("/admin/stats", handler.GetStats)
}

func (handler *Handler) GetSettings(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	settings, err := handler.repo.GetSettings(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (handler *Handler) UpdateSettings(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request Settings
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid settings payload"})
		return
	}

	settings, err := handler.repo.UpdateSettings(ctx.Request.Context(), request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (handler *Handler) GetNavigation(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	navigation, err := handler.repo.GetNavigation(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load navigation"})
		return
	}

	ctx.JSON(http.StatusOK, navigation)
}

func (handler *Handler) UpdateNavigation(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request Navigation
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid navigation payload"})
		return
	}

	navigation, err := handler.repo.UpdateNavigation(ctx.Request.Context(), request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update navigation"})
		return
	}

	ctx.JSON(http.StatusOK, navigation)
}

func (handler *Handler) ListMedia(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.ListMedia(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load media"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) GetStats(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	stats, err := handler.repo.GetStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
