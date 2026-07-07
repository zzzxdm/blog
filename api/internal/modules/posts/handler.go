package posts

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

const maxSearchKeywordRunes = 80

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func RegisterPublicRoutes(router gin.IRouter, repo Repository) {
	handler := NewHandler(repo)

	router.GET("/posts", handler.List)
	router.GET("/site-stats", handler.Stats)
	router.GET("/posts/:slug", handler.GetBySlug)
	router.GET("/categories/:slug/posts", handler.ListCategoryPosts)
	router.GET("/tags/:slug/posts", handler.ListTagPosts)
	router.GET("/search", handler.Search)
	router.GET("/me/private-posts", handler.ListPrivate)
	router.POST("/admin/published-posts/:slug/archive", handler.Archive)
}

func (handler *Handler) List(ctx *gin.Context) {
	handler.list(ctx, false)
}

func (handler *Handler) Search(ctx *gin.Context) {
	handler.list(ctx, true)
}

func (handler *Handler) ListPrivate(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}
	lister, ok := handler.repo.(PrivateLister)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "private post listing is unavailable"})
		return
	}

	result, err := lister.ListPrivate(ctx.Request.Context(), Viewer{ID: user.ID, Role: user.Role}, ListQuery{
		Keyword:  strings.TrimSpace(ctx.Query("q")),
		Category: strings.TrimSpace(ctx.Query("category")),
		Tag:      strings.TrimSpace(ctx.Query("tag")),
		Author:   strings.TrimSpace(ctx.Query("author")),
		Sort:     strings.TrimSpace(ctx.Query("sort")),
		Page:     intQuery(ctx, "page", 1),
		PageSize: intQuery(ctx, "pageSize", 10),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load private posts"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) Archive(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}
	archiver, ok := handler.repo.(Archiver)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "post archiving is unavailable"})
		return
	}
	if err := archiver.Archive(ctx.Request.Context(), ctx.Param("slug")); err != nil {
		if errors.Is(err, ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to archive post"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (handler *Handler) ListCategoryPosts(ctx *gin.Context) {
	ctx.Request.URL.RawQuery = mergeQuery(ctx.Request.URL.Query(), "category", ctx.Param("slug"))
	handler.list(ctx, false)
}

func (handler *Handler) ListTagPosts(ctx *gin.Context) {
	ctx.Request.URL.RawQuery = mergeQuery(ctx.Request.URL.Query(), "tag", ctx.Param("slug"))
	handler.list(ctx, false)
}

func (handler *Handler) GetBySlug(ctx *gin.Context) {
	post, err := handler.getPostForPublicView(ctx)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load post"})
		return
	}

	ctx.JSON(http.StatusOK, post)
}

func mergeQuery(values url.Values, key string, value string) string {
	values[key] = []string{value}
	return values.Encode()
}

func (handler *Handler) Stats(ctx *gin.Context) {
	stats, err := handler.repo.Stats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load site stats"})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

func (handler *Handler) getPostForPublicView(ctx *gin.Context) (Post, error) {
	slug := ctx.Param("slug")
	user, _ := auth.CurrentUser(ctx)
	viewer := Viewer{ID: user.ID, Role: user.Role}
	if recorder, ok := handler.repo.(RestrictedViewRecorder); ok {
		return recorder.RecordRestrictedView(ctx.Request.Context(), slug, viewer)
	}
	if getter, ok := handler.repo.(RestrictedGetter); ok {
		return getter.GetBySlugForViewer(ctx.Request.Context(), slug, viewer)
	}
	if recorder, ok := handler.repo.(ViewRecorder); ok {
		return recorder.RecordView(ctx.Request.Context(), slug)
	}

	return handler.repo.GetBySlug(ctx.Request.Context(), slug)
}

func (handler *Handler) list(ctx *gin.Context, forceKeyword bool) {
	query := ListQuery{
		Keyword:  strings.TrimSpace(ctx.Query("q")),
		Category: strings.TrimSpace(ctx.Query("category")),
		Tag:      strings.TrimSpace(ctx.Query("tag")),
		Author:   strings.TrimSpace(ctx.Query("author")),
		Sort:     strings.TrimSpace(ctx.Query("sort")),
		Page:     intQuery(ctx, "page", 1),
		PageSize: intQuery(ctx, "pageSize", 10),
	}

	if len([]rune(query.Keyword)) > maxSearchKeywordRunes {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "search keyword is too long"})
		return
	}

	if forceKeyword && query.Keyword == "" {
		ctx.JSON(http.StatusOK, ListResult{
			Items:    []Post{},
			Page:     normalizePage(query.Page),
			PageSize: normalizePageSize(query.PageSize),
			Total:    0,
		})
		return
	}

	result, err := handler.repo.List(ctx.Request.Context(), query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load posts"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func intQuery(ctx *gin.Context, key string, fallback int) int {
	value := ctx.Query(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
