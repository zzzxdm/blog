package reactions

import (
	"errors"
	"net/http"

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
	router.PUT("/posts/:slug/reaction", handler.SetReaction)
	router.PUT("/posts/:slug/bookmark", handler.SetBookmark)
	router.GET("/bookmarks/mine", handler.ListBookmarks)
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

type BookmarkItem struct {
	posts.Post
	BookmarkedAt string `json:"bookmarkedAt"`
}

type BookmarkListResult struct {
	Items []BookmarkItem `json:"items"`
	Total int            `json:"total"`
}

func (handler *Handler) ListBookmarks(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	bookmarks, err := handler.repo.ListBookmarks(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load bookmarks"})
		return
	}

	items := make([]BookmarkItem, 0, len(bookmarks))
	for _, bookmark := range bookmarks {
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
		Items: items,
		Total: len(items),
	})
}
