package posts

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"blog/api/internal/cachex"
	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

const maxSearchKeywordRunes = 80

type Handler struct {
	repo  Repository
	views cachex.ViewDeduper
	cache cachex.TTLStore
}

func NewHandler(repo Repository) *Handler {
	return NewHandlerWithDeps(repo, nil, nil)
}

func NewHandlerWithDeps(repo Repository, views cachex.ViewDeduper, cache cachex.TTLStore) *Handler {
	if views == nil {
		views = cachex.NewViewDeduper(nil)
	}
	if cache == nil {
		cache = cachex.NewTTLStore(nil)
	}
	return &Handler{repo: repo, views: views, cache: cache}
}

func RegisterPublicRoutes(router gin.IRouter, repo Repository) {
	RegisterPublicRoutesWithDeps(router, repo, nil, nil)
}

func RegisterPublicRoutesWithDeps(router gin.IRouter, repo Repository, views cachex.ViewDeduper, cache cachex.TTLStore) {
	handler := NewHandlerWithDeps(repo, views, cache)

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
		slog.Error("private post listing is unavailable")
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
		slog.Error("failed to load private posts", "error", err, "userID", user.ID, "role", user.Role, "keyword", ctx.Query("q"))
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
		slog.Error("post archiving is unavailable", "slug", ctx.Param("slug"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "post archiving is unavailable"})
		return
	}
	if err := archiver.Archive(ctx.Request.Context(), ctx.Param("slug")); err != nil {
		if errors.Is(err, ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		slog.Error("failed to archive post", "error", err, "slug", ctx.Param("slug"))
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

		slog.Error("failed to load post", "error", err, "slug", ctx.Param("slug"))
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
	if raw, ok := handler.cache.Get(ctx.Request.Context(), cachex.CacheKeySiteStats); ok {
		ctx.Data(http.StatusOK, "application/json; charset=utf-8", []byte(raw))
		return
	}

	stats, err := handler.repo.Stats(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to load site stats", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load site stats"})
		return
	}
	if raw, err := json.Marshal(stats); err == nil {
		handler.cache.Set(ctx.Request.Context(), cachex.CacheKeySiteStats, string(raw), cachex.PublicCacheTTL)
	}

	ctx.JSON(http.StatusOK, stats)
}

func (handler *Handler) getPostForPublicView(ctx *gin.Context) (Post, error) {
	slug := ctx.Param("slug")
	user, _ := auth.CurrentUser(ctx)
	viewer := Viewer{ID: user.ID, Role: user.Role}

	sessionToken, _ := ctx.Cookie(auth.SessionCookieName)
	visitor := cachex.VisitorKey(sessionToken, ctx.ClientIP())
	shouldCount := handler.views == nil || handler.views.Allow(ctx.Request.Context(), visitor, slug)

	if shouldCount {
		if recorder, ok := handler.repo.(RestrictedViewRecorder); ok {
			return recorder.RecordRestrictedView(ctx.Request.Context(), slug, viewer)
		}
		if recorder, ok := handler.repo.(ViewRecorder); ok {
			return recorder.RecordView(ctx.Request.Context(), slug)
		}
	}

	if getter, ok := handler.repo.(RestrictedGetter); ok {
		return getter.GetBySlugForViewer(ctx.Request.Context(), slug, viewer)
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
		slog.Error("failed to load posts", "error", err, "keyword", query.Keyword, "category", query.Category, "tag", query.Tag, "author", query.Author, "sort", query.Sort, "page", query.Page, "pageSize", query.PageSize)
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
