package reactions

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

	router.GET("/posts/:slug/reaction", handler.Get)
	router.PUT("/posts/:slug/reaction", handler.SetReaction)
	router.PUT("/posts/:slug/bookmark", handler.SetBookmark)
}

func (handler *Handler) Get(ctx *gin.Context) {
	user, _ := auth.CurrentUser(ctx)

	summary, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("slug"), user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load reaction"})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

func (handler *Handler) SetReaction(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	var request ReactionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid reaction payload"})
		return
	}

	summary, err := handler.repo.SetReaction(ctx.Request.Context(), ctx.Param("slug"), user.ID, request.Type)
	if err != nil {
		if errors.Is(err, ErrInvalidReaction) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "reaction must be like, dislike, or empty"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update reaction"})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

func (handler *Handler) SetBookmark(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	var request BookmarkRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid bookmark payload"})
		return
	}

	summary, err := handler.repo.SetBookmark(ctx.Request.Context(), ctx.Param("slug"), user.ID, request.Bookmarked)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update bookmark"})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}
