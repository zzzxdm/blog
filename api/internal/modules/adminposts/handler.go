package adminposts

import (
	"errors"
	"net/http"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/posts"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo      Repository
	publisher posts.Publisher
}

func NewHandler(repo Repository, publisher posts.Publisher) *Handler {
	return &Handler{repo: repo, publisher: publisher}
}

func RegisterRoutes(router gin.IRouter, repo Repository, publisher posts.Publisher) {
	handler := NewHandler(repo, publisher)

	router.GET("/admin/posts", handler.List)
	router.GET("/admin/posts/:id", handler.Get)
	router.POST("/admin/posts", handler.Create)
	router.PUT("/admin/posts/:id", handler.Update)
	router.POST("/admin/posts/:id/publish", handler.Publish)
}

func (handler *Handler) List(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load admin posts"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) Get(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	post, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, post)
}

func (handler *Handler) Create(ctx *gin.Context) {
	handler.save(ctx, "")
}

func (handler *Handler) Update(ctx *gin.Context) {
	handler.save(ctx, ctx.Param("id"))
}

func (handler *Handler) save(ctx *gin.Context, id string) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request SaveRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post payload"})
		return
	}

	post, err := handler.repo.Save(ctx.Request.Context(), id, request)
	if err != nil {
		handler.writeError(ctx, err)
		return
	}

	if id == "" {
		ctx.JSON(http.StatusCreated, post)
		return
	}
	ctx.JSON(http.StatusOK, post)
}

func (handler *Handler) Publish(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	post, err := handler.repo.Publish(ctx.Request.Context(), ctx.Param("id"), handler.publisher)
	if err != nil {
		handler.writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, post)
}

func (handler *Handler) writeError(ctx *gin.Context, err error) {
	if errors.Is(err, ErrPostNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "admin post not found"})
		return
	}
	if errors.Is(err, ErrInvalidPost) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "title and content are required"})
		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update admin post"})
}
