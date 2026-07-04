package submissions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/messages"
	"blog/api/internal/modules/operations"
	"blog/api/internal/modules/posts"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo      Repository
	messages  messages.Repository
	publisher posts.Publisher
	settings  settingsReader
}

type settingsReader interface {
	GetSettings(ctx context.Context) (operations.Settings, error)
}

func NewHandler(repo Repository, messageRepo messages.Repository, publisher posts.Publisher, settings settingsReader) *Handler {
	return &Handler{
		repo:      repo,
		messages:  messageRepo,
		publisher: publisher,
		settings:  settings,
	}
}

func RegisterRoutes(router gin.IRouter, repo Repository, messageRepo messages.Repository, publisher posts.Publisher, settings settingsReader) {
	handler := NewHandler(repo, messageRepo, publisher, settings)

	router.GET("/submissions", handler.ListMine)
	router.POST("/submissions", handler.Create)
	router.PUT("/submissions/:id", handler.Update)
	router.POST("/submissions/:id/submit", handler.Submit)

	router.GET("/admin/submissions", handler.AdminList)
	router.PUT("/admin/submissions/:id", handler.AdminUpdate)
	router.POST("/admin/submissions/:id/review", handler.Review)
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load submissions"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) Create(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}
	if !canSubmit(user) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "user is not allowed to submit posts"})
		return
	}
	settings, ok := handler.requireSubmissionSettings(ctx)
	if !ok {
		return
	}

	var request SaveRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission payload"})
		return
	}
	if saveRequestContainsBlockedWord(request, settings.BlockedWords) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "submission contains blocked word"})
		return
	}
	if request.Submit && !handler.requireSubmissionSlot(ctx, user.ID, settings, "") {
		return
	}

	submission, err := handler.repo.Create(ctx.Request.Context(), request, user)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, submission)
}

func (handler *Handler) Update(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}
	if !canSubmit(user) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "user is not allowed to update submissions"})
		return
	}
	settings, ok := handler.requireSubmissionSettings(ctx)
	if !ok {
		return
	}

	var request SaveRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission payload"})
		return
	}
	if saveRequestContainsBlockedWord(request, settings.BlockedWords) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "submission contains blocked word"})
		return
	}
	if request.Submit && !handler.requireSubmissionSlot(ctx, user.ID, settings, ctx.Param("id")) {
		return
	}

	submission, err := handler.repo.Update(ctx.Request.Context(), ctx.Param("id"), user.ID, request)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func (handler *Handler) Submit(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}
	if !canSubmit(user) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "user is not allowed to submit posts"})
		return
	}
	settings, ok := handler.requireSubmissionSettings(ctx)
	if !ok {
		return
	}

	current, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}
	if current.AuthorID != user.ID {
		handler.writeSubmissionError(ctx, ErrForbidden)
		return
	}
	if submissionContainsBlockedWord(current, settings.BlockedWords) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "submission contains blocked word"})
		return
	}
	if !handler.requireSubmissionSlot(ctx, user.ID, settings, current.ID) {
		return
	}

	submission, err := handler.repo.Submit(ctx.Request.Context(), ctx.Param("id"), user.ID)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func (handler *Handler) requireSubmissionSettings(ctx *gin.Context) (operations.Settings, bool) {
	if handler.settings == nil {
		return operations.Settings{SubmissionsEnabled: true}, true
	}

	settings, err := handler.settings.GetSettings(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load submission settings"})
		return operations.Settings{}, false
	}
	if !settings.SubmissionsEnabled {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "submissions are disabled"})
		return operations.Settings{}, false
	}

	return settings, true
}

func (handler *Handler) requireSubmissionSlot(ctx *gin.Context, userID string, settings operations.Settings, excludeID string) bool {
	since, limit, ok := submissionLimitWindow(settings.SubmissionLimit, time.Now())
	if !ok {
		return true
	}

	total, err := handler.repo.CountSubmittedSince(ctx.Request.Context(), userID, since, excludeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load submission limit"})
		return false
	}
	if total >= limit {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "submission limit exceeded"})
		return false
	}

	return true
}

func (handler *Handler) AdminList(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.AdminList(ctx.Request.Context(), ListQuery{
		Status: ctx.Query("status"),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load submissions"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) AdminUpdate(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request SaveRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission payload"})
		return
	}

	submission, err := handler.repo.AdminUpdate(ctx.Request.Context(), ctx.Param("id"), request)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func (handler *Handler) Review(ctx *gin.Context) {
	reviewer, ok := auth.RequireAdmin(ctx)
	if !ok {
		return
	}

	var request ReviewRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid review payload"})
		return
	}

	publishedPostSlug := ""
	if strings.ToLower(strings.TrimSpace(request.Action)) == ActionApprove {
		submission, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
		if err != nil {
			handler.writeSubmissionError(ctx, err)
			return
		}
		if !canReviewSubmission(submission.Status) {
			handler.writeSubmissionError(ctx, ErrInvalidReview)
			return
		}

		if handler.publisher == nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "post publisher is unavailable"})
			return
		}

		post, err := handler.publisher.Publish(ctx.Request.Context(), posts.PublishInput{
			Slug:       defaultString(strings.TrimSpace(request.Slug), submission.Slug),
			Title:      submission.Title,
			Summary:    submission.Summary,
			Content:    submission.Content,
			Category:   defaultString(strings.TrimSpace(request.Category), submission.Category),
			Tags:       submission.Tags,
			CoverImage: submission.CoverImage,
			AuthorName: submission.AuthorName,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish submission"})
			return
		}
		publishedPostSlug = post.Slug
	}

	submission, err := handler.repo.Review(ctx.Request.Context(), ctx.Param("id"), reviewer, request, publishedPostSlug)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}

	if _, err := handler.messages.Create(ctx.Request.Context(), reviewMessage(submission, request.Action), reviewer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review message"})
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func (handler *Handler) writeSubmissionError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSubmissionNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
	case errors.Is(err, ErrForbidden):
		ctx.JSON(http.StatusForbidden, gin.H{"error": "submission forbidden"})
	case errors.Is(err, ErrInvalidSubmission):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "title and content are required before submit"})
	case errors.Is(err, ErrInvalidReview):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid review action"})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update submission"})
	}
}

func reviewMessage(submission Submission, action string) messages.CreateRequest {
	action = strings.ToLower(strings.TrimSpace(action))
	title := "你的投稿审核结果已更新"
	body := fmt.Sprintf("《%s》的审核结果已更新。", submission.Title)

	switch action {
	case ActionApprove:
		title = "你的投稿已通过并发布"
		body = fmt.Sprintf("《%s》已通过审核并发布到站点。", submission.Title)
	case ActionReturn:
		title = "你的投稿已退回修改"
		body = fmt.Sprintf("《%s》暂未通过审核，请根据审核意见修改后重新提交。", submission.Title)
	case ActionReject:
		title = "你的投稿未通过审核"
		body = fmt.Sprintf("《%s》未通过本次审核。", submission.Title)
	}

	if strings.TrimSpace(submission.ReviewNote) != "" {
		body = body + " 审核意见：" + strings.TrimSpace(submission.ReviewNote)
	}

	return messages.CreateRequest{
		RecipientID:   submission.AuthorID,
		RecipientName: submission.AuthorName,
		Type:          messages.TypeReview,
		Priority:      "normal",
		Title:         title,
		Body:          body,
		TargetType:    "submission",
		TargetID:      submission.ID,
		TargetTitle:   submission.Title,
	}
}

func canSubmit(user auth.User) bool {
	return user.Status == "" || user.Status == "active"
}

func canReviewSubmission(status string) bool {
	return status == StatusSubmitted || status == StatusReturned
}

func saveRequestContainsBlockedWord(request SaveRequest, blockedWords []string) bool {
	return textContainsBlockedWord(strings.Join([]string{
		request.Title,
		request.Summary,
		request.Content,
		strings.Join(request.Tags, " "),
	}, " "), blockedWords)
}

func submissionContainsBlockedWord(submission Submission, blockedWords []string) bool {
	return textContainsBlockedWord(strings.Join([]string{
		submission.Title,
		submission.Summary,
		submission.Content,
		strings.Join(submission.Tags, " "),
	}, " "), blockedWords)
}

func textContainsBlockedWord(value string, blockedWords []string) bool {
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

func submissionLimitWindow(value string, now time.Time) (time.Time, int, bool) {
	limit := firstPositiveInt(value)
	if limit <= 0 {
		return time.Time{}, 0, false
	}

	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if strings.Contains(value, "\u5468") {
		offset := (int(now.Weekday()) + 6) % 7
		start = start.AddDate(0, 0, -offset)
	}

	return start, limit, true
}

func firstPositiveInt(value string) int {
	number := 0
	found := false
	for _, item := range value {
		if item < '0' || item > '9' {
			if found {
				break
			}
			continue
		}

		found = true
		number = number*10 + int(item-'0')
	}

	if !found {
		return 0
	}

	return number
}
