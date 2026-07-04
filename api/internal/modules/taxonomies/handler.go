package taxonomies

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
		"items": categories,
		"total": len(categories),
	})
}

func (handler *Handler) ListTags(ctx *gin.Context) {
	tags, err := handler.repo.ListTags(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load tags"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items": tags,
		"total": len(tags),
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
