package posts

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	router.GET("/posts/:slug", handler.GetBySlug)
	router.GET("/search", handler.Search)
}

func (handler *Handler) List(ctx *gin.Context) {
	handler.list(ctx, false)
}

func (handler *Handler) Search(ctx *gin.Context) {
	handler.list(ctx, true)
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

func (handler *Handler) getPostForPublicView(ctx *gin.Context) (Post, error) {
	slug := ctx.Param("slug")
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
