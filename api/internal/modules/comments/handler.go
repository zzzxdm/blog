package comments

import (
	"blog/api/internal/httpx"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/operations"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo     Repository
	settings settingsReader
}

type settingsReader interface {
	GetSettings(ctx context.Context) (operations.Settings, error)
}

func NewHandler(repo Repository, settings settingsReader) *Handler {
	return &Handler{repo: repo, settings: settings}
}

func RegisterRoutes(router gin.IRouter, repo Repository, settings settingsReader) {
	handler := NewHandler(repo, settings)

	router.GET("/posts/:slug/comments", handler.List)
	router.POST("/posts/:slug/comments", handler.Create)
	router.POST("/comments/:id/replies", handler.CreateReply)
	router.DELETE("/comments/:id", handler.DeleteMine)
	router.POST("/comments/:id/like", handler.ToggleLike)
	router.PUT("/comments/:id/like", handler.ToggleLike)
	router.POST("/comments/:id/report", handler.Report)
	router.GET("/comments/mine", handler.ListMine)
	router.GET("/me/comments", handler.ListMine)
	router.GET("/admin/comments", handler.AdminList)
	router.GET("/admin/comments/export", handler.AdminExport)
	router.PUT("/admin/comments/:id/status", handler.UpdateStatus)
	router.DELETE("/admin/comments/:id", handler.DeleteAdmin)
	router.GET("/admin/comment-reports", handler.ListReports)
	router.PUT("/admin/comment-reports/:id/status", handler.UpdateReportStatus)
}

func (handler *Handler) List(ctx *gin.Context) {
	viewer, _ := auth.CurrentUser(ctx)

	result, err := handler.repo.List(ctx.Request.Context(), ctx.Param("slug"), viewer.ID)
	if err != nil {
		slog.Error("failed to load public comments", "error", err, "slug", ctx.Param("slug"), "viewerID", viewer.ID)
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
	if !canInteract(user) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "user is not allowed to comment"})
		return
	}
	settings, ok := handler.requireCommentSettings(ctx)
	if !ok {
		return
	}

	var request CreateRequest
	if !httpx.BindJSON(ctx, &request, "invalid comment payload") {
		return
	}
	if containsBlockedWord(request.Body, settings.BlockedWords) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "comment contains blocked word"})
		return
	}

	comment, err := handler.repo.Create(ctx.Request.Context(), ctx.Param("slug"), request, user)
	if err != nil {
		if errors.Is(err, ErrEmptyBody) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "comment body is required"})
			return
		}

		slog.Error("failed to create comment", "error", err, "slug", ctx.Param("slug"), "userID", user.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}
	if settings.AutoApproveComments {
		approved, err := handler.repo.UpdateStatus(ctx.Request.Context(), comment.ID, "approved")
		if err == nil {
			comment = approved
		}
	}

	ctx.JSON(http.StatusCreated, comment)
}

func (handler *Handler) CreateReply(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}
	if !canInteract(user) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "user is not allowed to reply"})
		return
	}
	settings, ok := handler.requireCommentSettings(ctx)
	if !ok {
		return
	}

	var request CreateRequest
	if !httpx.BindJSON(ctx, &request, "invalid reply payload") {
		return
	}
	if containsBlockedWord(request.Body, settings.BlockedWords) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "reply contains blocked word"})
		return
	}

	comment, err := handler.repo.CreateReply(ctx.Request.Context(), ctx.Param("id"), request, user)
	if err != nil {
		if errors.Is(err, ErrEmptyBody) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "reply body is required"})
			return
		}
		if errors.Is(err, ErrCommentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}

		slog.Error("failed to create reply", "error", err, "parentID", ctx.Param("id"), "userID", user.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reply"})
		return
	}
	if settings.AutoApproveComments {
		approved, err := handler.repo.UpdateStatus(ctx.Request.Context(), comment.ID, "approved")
		if err == nil {
			comment = approved
		}
	}

	ctx.JSON(http.StatusCreated, comment)
}

func (handler *Handler) requireCommentSettings(ctx *gin.Context) (operations.Settings, bool) {
	if handler.settings == nil {
		return operations.Settings{CommentsEnabled: true}, true
	}

	settings, err := handler.settings.GetSettings(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to load comment settings", "error", err, "path", ctx.FullPath())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load comment settings"})
		return operations.Settings{}, false
	}
	if !settings.CommentsEnabled {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "comments are disabled"})
		return operations.Settings{}, false
	}

	return settings, true
}

func (handler *Handler) ToggleLike(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}
	if !canInteract(user) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "user is not allowed to like comments"})
		return
	}

	comment, err := handler.repo.ToggleLike(ctx.Request.Context(), ctx.Param("id"), user.ID)
	if err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}

		slog.Error("failed to update comment like", "error", err, "commentID", ctx.Param("id"), "userID", user.ID)
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
	if !canInteract(user) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "user is not allowed to report comments"})
		return
	}

	var request ReportRequest
	if !httpx.BindJSON(ctx, &request, "invalid report payload") {
		return
	}

	if err := handler.repo.Report(ctx.Request.Context(), ctx.Param("id"), user, request); err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}

		slog.Error("failed to report comment", "error", err, "commentID", ctx.Param("id"), "userID", user.ID)
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
		Status:   ctx.Query("status"),
		Keyword:  ctx.Query("q"),
		Sort:     ctx.Query("sort"),
		Page:     parsePositiveInt(ctx.Query("page")),
		PageSize: parsePositiveInt(ctx.Query("pageSize")),
	})
	if err != nil {
		slog.Error("failed to load user comments", "error", err, "userID", user.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load comments"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) DeleteMine(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	comment, err := handler.repo.DeleteByAuthor(ctx.Request.Context(), ctx.Param("id"), user.ID)
	if err != nil {
		handler.writeCommentError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true, "comment": comment})
}

func (handler *Handler) AdminList(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.AdminList(ctx.Request.Context(), ListQuery{
		Status:   ctx.Query("status"),
		Keyword:  ctx.Query("q"),
		Sort:     ctx.Query("sort"),
		Page:     parsePositiveInt(ctx.Query("page")),
		PageSize: parsePositiveInt(ctx.Query("pageSize")),
	})
	if err != nil {
		slog.Error("failed to load admin comments", "error", err, "status", ctx.Query("status"), "keyword", ctx.Query("q"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load comments"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) AdminExport(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.AdminList(ctx.Request.Context(), ListQuery{
		Status: ctx.Query("status"),
		All:    true,
	})
	if err != nil {
		slog.Error("failed to export comments", "error", err, "status", ctx.Query("status"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export comments"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"scope":      "comments",
		"exportedAt": time.Now(),
		"items":      result.Items,
		"total":      result.Total,
		"stats":      result.Stats,
	})
}

func (handler *Handler) UpdateStatus(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request StatusRequest
	if !httpx.BindJSON(ctx, &request, "invalid comment status payload") {
		return
	}

	comment, err := handler.repo.UpdateStatus(ctx.Request.Context(), ctx.Param("id"), request.Status)
	if err != nil {
		handler.writeCommentError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, comment)
}

func (handler *Handler) DeleteAdmin(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	comment, err := handler.repo.UpdateStatus(ctx.Request.Context(), ctx.Param("id"), "deleted")
	if err != nil {
		handler.writeCommentError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true, "comment": comment})
}

func (handler *Handler) ListReports(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.ListReports(ctx.Request.Context(), ctx.Query("status"))
	if err != nil {
		slog.Error("failed to load comment reports", "error", err, "status", ctx.Query("status"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load comment reports"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) UpdateReportStatus(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request StatusRequest
	if !httpx.BindJSON(ctx, &request, "invalid report status payload") {
		return
	}

	report, err := handler.repo.UpdateReportStatus(ctx.Request.Context(), ctx.Param("id"), request.Status)
	if err != nil {
		handler.writeCommentError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, report)
}

func (handler *Handler) writeCommentError(ctx *gin.Context, err error) {
	if errors.Is(err, ErrCommentNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	if errors.Is(err, ErrInvalidStatus) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment status"})
		return
	}
	if errors.Is(err, ErrForbidden) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "comment forbidden"})
		return
	}

	slog.Error("failed to update comment", "error", err, "path", ctx.FullPath(), "commentID", ctx.Param("id"))
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update comment"})
}

func canInteract(user auth.User) bool {
	return (user.Status == "" || user.Status == "active") && user.EmailVerified
}

func containsBlockedWord(value string, blockedWords []string) bool {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	if normalizedValue == "" {
		return false
	}

	for _, word := range blockedWords {
		normalizedWord := strings.ToLower(strings.TrimSpace(word))
		if normalizedWord != "" && strings.Contains(normalizedValue, normalizedWord) {
			return true
		}
	}

	return false
}

func parsePositiveInt(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return 0
	}

	return parsed
}
