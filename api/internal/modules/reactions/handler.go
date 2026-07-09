package reactions

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/posts"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo     Repository
	postRepo posts.Repository
}

func NewHandler(repo Repository, postRepo posts.Repository) *Handler {
	return &Handler{repo: repo, postRepo: postRepo}
}

func RegisterRoutes(router gin.IRouter, repo Repository, postRepo posts.Repository) {
	handler := NewHandler(repo, postRepo)

	router.GET("/posts/:slug/reaction", handler.Get)
	router.POST("/posts/:slug/bookmark", handler.SetBookmark)
	router.PUT("/posts/:slug/reaction", handler.SetReaction)
	router.DELETE("/posts/:slug/reaction", handler.ClearReaction)
	router.PUT("/posts/:slug/bookmark", handler.SetBookmark)
	router.GET("/bookmarks/mine", handler.ListBookmarks)
	router.GET("/me/bookmarks", handler.ListBookmarks)
}

func (handler *Handler) Get(ctx *gin.Context) {
	user, _ := auth.CurrentUser(ctx)
	if !handler.ensurePostExists(ctx) {
		return
	}

	summary, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("slug"), user.ID)
	if err != nil {
		slog.Error("failed to load reaction", "error", err, "slug", ctx.Param("slug"), "userID", user.ID)
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
	if !canInteract(user) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "user is not allowed to react to posts"})
		return
	}
	if !handler.ensurePostExists(ctx) {
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

		slog.Error("failed to update reaction", "error", err, "slug", ctx.Param("slug"), "userID", user.ID, "type", request.Type)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update reaction"})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

func (handler *Handler) ClearReaction(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}
	if !canInteract(user) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "user is not allowed to react to posts"})
		return
	}
	if !handler.ensurePostExists(ctx) {
		return
	}

	summary, err := handler.repo.SetReaction(ctx.Request.Context(), ctx.Param("slug"), user.ID, "")
	if err != nil {
		slog.Error("failed to clear reaction", "error", err, "slug", ctx.Param("slug"), "userID", user.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear reaction"})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

func (handler *Handler) SetBookmark(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}
	if !canInteract(user) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "user is not allowed to bookmark posts"})
		return
	}
	if !handler.ensurePostExists(ctx) {
		return
	}

	var request BookmarkRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid bookmark payload"})
		return
	}

	summary, err := handler.repo.SetBookmark(ctx.Request.Context(), ctx.Param("slug"), user.ID, request.Bookmarked)
	if err != nil {
		slog.Error("failed to update bookmark", "error", err, "slug", ctx.Param("slug"), "userID", user.ID, "bookmarked", request.Bookmarked)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update bookmark"})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

func (handler *Handler) ensurePostExists(ctx *gin.Context) bool {
	if handler.postRepo == nil {
		slog.Error("post repository is unavailable for reaction", "slug", ctx.Param("slug"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "post repository is unavailable"})
		return false
	}

	if _, err := handler.postRepo.GetBySlug(ctx.Request.Context(), ctx.Param("slug")); err != nil {
		if errors.Is(err, posts.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return false
		}

		slog.Error("failed to load post before reaction", "error", err, "slug", ctx.Param("slug"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load post"})
		return false
	}

	return true
}

type BookmarkItem struct {
	posts.Post
	BookmarkedAt string `json:"bookmarkedAt"`
}

type BookmarkListResult struct {
	Items    []BookmarkItem `json:"items"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
	Total    int            `json:"total"`
}

func (handler *Handler) ListBookmarks(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	bookmarks, err := handler.repo.ListBookmarks(ctx.Request.Context(), user.ID, BookmarkQuery{
		Keyword:  ctx.Query("q"),
		Category: ctx.Query("category"),
		Sort:     ctx.Query("sort"),
		Page:     parsePositiveInt(ctx.Query("page")),
		PageSize: parsePositiveInt(ctx.Query("pageSize")),
	})
	if err != nil {
		slog.Error("failed to load bookmarks", "error", err, "userID", user.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load bookmarks"})
		return
	}

	items := make([]BookmarkItem, 0, len(bookmarks.Items))
	for _, bookmark := range bookmarks.Items {
		post, err := handler.postRepo.GetBySlug(ctx.Request.Context(), bookmark.PostSlug)
		if err != nil {
			continue
		}

		items = append(items, BookmarkItem{
			Post:         post,
			BookmarkedAt: bookmark.BookmarkedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	ctx.JSON(http.StatusOK, BookmarkListResult{
		Items:    items,
		Page:     bookmarks.Page,
		PageSize: bookmarks.PageSize,
		Total:    bookmarks.Total,
	})
}

func canInteract(user auth.User) bool {
	return (user.Status == "" || user.Status == "active") && user.EmailVerified
}

func parsePositiveInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0
	}
	return parsed
}
