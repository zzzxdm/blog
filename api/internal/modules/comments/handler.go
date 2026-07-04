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
	router.PUT("/comments/:id/like", handler.ToggleLike)
	router.POST("/comments/:id/report", handler.Report)
	router.GET("/comments/mine", handler.ListMine)
	router.GET("/admin/comments", handler.AdminList)
	router.PUT("/admin/comments/:id/status", handler.UpdateStatus)
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

func (handler *Handler) ToggleLike(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	comment, err := handler.repo.ToggleLike(ctx.Request.Context(), ctx.Param("id"), user.ID)
	if err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update comment like"})
		return
	}

	ctx.JSON(http.StatusOK, comment)
}

func (handler *Handler) Report(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	var request ReportRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid report payload"})
		return
	}

	if err := handler.repo.Report(ctx.Request.Context(), ctx.Param("id"), user, request); err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to report comment"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (handler *Handler) ListMine(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	result, err := handler.repo.ListByAuthor(ctx.Request.Context(), user.ID, ListQuery{
		Status: ctx.Query("status"),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load comments"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) AdminList(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.AdminList(ctx.Request.Context(), ListQuery{
		Status: ctx.Query("status"),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load comments"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) UpdateStatus(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request StatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment status payload"})
		return
	}

	comment, err := handler.repo.UpdateStatus(ctx.Request.Context(), ctx.Param("id"), request.Status)
	if err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}
		if errors.Is(err, ErrInvalidStatus) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment status"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update comment"})
		return
	}

	ctx.JSON(http.StatusOK, comment)
}
