package taxonomies

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

	router.GET("/categories", handler.ListCategories)
	router.GET("/tags", handler.ListTags)
	router.POST("/admin/categories", handler.CreateCategory)
	router.PUT("/admin/categories/:id", handler.UpdateCategory)
	router.DELETE("/admin/categories/:id", handler.DeleteCategory)
	router.POST("/admin/tags", handler.CreateTag)
	router.PUT("/admin/tags/:id", handler.UpdateTag)
	router.DELETE("/admin/tags/:id", handler.DeleteTag)
}

func (handler *Handler) ListCategories(ctx *gin.Context) {
	categories, err := handler.repo.ListCategories(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load categories"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items":    pageItems(categories, ctx),
		"page":     taxonomyPage(ctx),
		"pageSize": taxonomyPageSize(ctx, len(categories)),
		"total":    len(categories),
	})
}

func (handler *Handler) ListTags(ctx *gin.Context) {
	tags, err := handler.repo.ListTags(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load tags"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items":    pageItems(tags, ctx),
		"page":     taxonomyPage(ctx),
		"pageSize": taxonomyPageSize(ctx, len(tags)),
		"total":    len(tags),
	})
}

func (handler *Handler) CreateCategory(ctx *gin.Context) {
	handler.saveCategory(ctx, "")
}

func (handler *Handler) UpdateCategory(ctx *gin.Context) {
	handler.saveCategory(ctx, ctx.Param("id"))
}

func (handler *Handler) saveCategory(ctx *gin.Context, id string) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request SaveCategoryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid category payload"})
		return
	}

	category, err := handler.repo.SaveCategory(ctx.Request.Context(), id, request)
	if err != nil {
		handler.writeError(ctx, err, "category")
		return
	}

	if id == "" {
		ctx.JSON(http.StatusCreated, category)
		return
	}
	ctx.JSON(http.StatusOK, category)
}

func (handler *Handler) DeleteCategory(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	if err := handler.repo.DeleteCategory(ctx.Request.Context(), ctx.Param("id")); err != nil {
		handler.writeError(ctx, err, "category")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (handler *Handler) CreateTag(ctx *gin.Context) {
	handler.saveTag(ctx, "")
}

func (handler *Handler) UpdateTag(ctx *gin.Context) {
	handler.saveTag(ctx, ctx.Param("id"))
}

func (handler *Handler) saveTag(ctx *gin.Context, id string) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request SaveTagRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag payload"})
		return
	}

	tag, err := handler.repo.SaveTag(ctx.Request.Context(), id, request)
	if err != nil {
		handler.writeError(ctx, err, "tag")
		return
	}

	if id == "" {
		ctx.JSON(http.StatusCreated, tag)
		return
	}
	ctx.JSON(http.StatusOK, tag)
}

func (handler *Handler) DeleteTag(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	if err := handler.repo.DeleteTag(ctx.Request.Context(), ctx.Param("id")); err != nil {
		handler.writeError(ctx, err, "tag")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (handler *Handler) writeError(ctx *gin.Context, err error, resource string) {
	if errors.Is(err, ErrNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": resource + " not found"})
		return
	}
	if errors.Is(err, ErrInvalid) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": resource + " name is required"})
		return
	}
	if errors.Is(err, ErrDuplicate) {
		ctx.JSON(http.StatusConflict, gin.H{"error": resource + " slug or name already exists"})
		return
	}
	if errors.Is(err, ErrTaxonomyInUse) {
		ctx.JSON(http.StatusConflict, gin.H{"error": resource + " is still used by posts"})
		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update " + resource})
}

func pageItems[T any](items []T, ctx *gin.Context) []T {
	pageSize := parsePositiveInt(ctx.Query("pageSize"))
	if pageSize < 1 {
		return items
	}
	page := parsePositiveInt(ctx.Query("page"))
	if page < 1 {
		page = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func taxonomyPage(ctx *gin.Context) int {
	pageSize := parsePositiveInt(ctx.Query("pageSize"))
	if pageSize < 1 {
		return 1
	}
	page := parsePositiveInt(ctx.Query("page"))
	if page < 1 {
		return 1
	}
	return page
}

func taxonomyPageSize(ctx *gin.Context, total int) int {
	pageSize := parsePositiveInt(ctx.Query("pageSize"))
	if pageSize < 1 {
		return total
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

func parsePositiveInt(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return 0
	}
	return parsed
}
