package comments

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

	router.GET("/posts/:slug/comments", handler.List)
	router.POST("/posts/:slug/comments", handler.Create)
}

func (handler *Handler) List(ctx *gin.Context) {
	viewer, _ := auth.CurrentUser(ctx)

	result, err := handler.repo.List(ctx.Request.Context(), ctx.Param("slug"), viewer.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load comments"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) Create(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	var request CreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment payload"})
		return
	}

	comment, err := handler.repo.Create(ctx.Request.Context(), ctx.Param("slug"), request, user)
	if err != nil {
		if errors.Is(err, ErrEmptyBody) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "comment body is required"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	ctx.JSON(http.StatusCreated, comment)
}
