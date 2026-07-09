package submissions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/messages"
	"blog/api/internal/modules/operations"
	"blog/api/internal/modules/posts"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo              Repository
	messages          messages.Repository
	publisher         posts.SubmissionPublisher
	archiver          posts.Archiver
	restorer          posts.Restorer
	settings          settingsReader
	turnstileVerifier auth.TurnstileVerifier
}

type settingsReader interface {
	GetSettings(ctx context.Context) (operations.Settings, error)
}

func NewHandler(repo Repository, messageRepo messages.Repository, publisher posts.SubmissionPublisher, archiver posts.Archiver, restorer posts.Restorer, settings settingsReader) *Handler {
	return NewHandlerWithTurnstile(repo, messageRepo, publisher, archiver, restorer, settings, nil)
}

func NewHandlerWithTurnstile(repo Repository, messageRepo messages.Repository, publisher posts.SubmissionPublisher, archiver posts.Archiver, restorer posts.Restorer, settings settingsReader, turnstileVerifier auth.TurnstileVerifier) *Handler {
	if turnstileVerifier == nil {
		turnstileVerifier = auth.NewHTTPTurnstileVerifier()
	}

	return &Handler{
		repo:              repo,
		messages:          messageRepo,
		publisher:         publisher,
		archiver:          archiver,
		restorer:          restorer,
		settings:          settings,
		turnstileVerifier: turnstileVerifier,
	}
}

func RegisterRoutes(router gin.IRouter, repo Repository, messageRepo messages.Repository, publisher posts.SubmissionPublisher, archiver posts.Archiver, restorer posts.Restorer, settings settingsReader) {
	RegisterRoutesWithTurnstile(router, repo, messageRepo, publisher, archiver, restorer, settings, nil)
}

func RegisterRoutesWithTurnstile(router gin.IRouter, repo Repository, messageRepo messages.Repository, publisher posts.SubmissionPublisher, archiver posts.Archiver, restorer posts.Restorer, settings settingsReader, turnstileVerifier auth.TurnstileVerifier) {
	handler := NewHandlerWithTurnstile(repo, messageRepo, publisher, archiver, restorer, settings, turnstileVerifier)

	router.GET("/submissions", handler.ListMine)
	router.GET("/me/submissions", handler.ListMine)
	router.POST("/submissions", handler.Create)
	router.PUT("/submissions/:id", handler.Update)
	router.POST("/submissions/:id/submit", handler.Submit)
	router.DELETE("/submissions/:id", handler.DeleteMine)

	router.GET("/admin/submissions", handler.AdminList)
	router.GET("/admin/submissions/:id", handler.AdminGet)
	router.PUT("/admin/submissions/:id", handler.AdminUpdate)
	router.POST("/admin/submissions/:id/review", handler.Review)
	router.POST("/admin/submissions/:id/approve", handler.Approve)
	router.POST("/admin/submissions/:id/reject", handler.Reject)
	router.POST("/admin/submissions/:id/archive", handler.ArchivePublished)
	router.POST("/admin/submissions/:id/restore", handler.RestorePublished)
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
		All:      boolQuery(ctx.Query("all")),
	})
	if err != nil {
		slog.Error("failed to load user submissions", "error", err, "userID", user.ID, "status", ctx.Query("status"), "keyword", ctx.Query("q"))
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
	if request.Submit && !handler.verifyTurnstile(ctx, settings, request.TurnstileToken) {
		return
	}
	if request.Submit && normalizeVisibility(request.Visibility) == VisibilityPublic && !handler.requireSubmissionSlot(ctx, user.ID, settings, "") {
		return
	}

	submission, err := handler.repo.Create(ctx.Request.Context(), request, user)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}
	if request.Submit && submission.Visibility == VisibilityPrivate {
		submission, err = handler.publishPrivateSubmission(ctx, user, submission)
		if err != nil {
			return
		}
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
	if request.Submit && !handler.verifyTurnstile(ctx, settings, request.TurnstileToken) {
		return
	}
	if request.Submit && normalizeVisibility(request.Visibility) == VisibilityPublic && !handler.requireSubmissionSlot(ctx, user.ID, settings, ctx.Param("id")) {
		return
	}

	submission, err := handler.repo.Update(ctx.Request.Context(), ctx.Param("id"), user.ID, request)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}
	if request.Submit && submission.Visibility == VisibilityPrivate {
		submission, err = handler.publishPrivateSubmission(ctx, user, submission)
		if err != nil {
			return
		}
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
	var request struct {
		TurnstileToken string `json:"turnstileToken"`
	}
	_ = ctx.ShouldBindJSON(&request)
	if !handler.verifyTurnstile(ctx, settings, request.TurnstileToken) {
		return
	}
	if current.Visibility == VisibilityPublic && !handler.requireSubmissionSlot(ctx, user.ID, settings, current.ID) {
		return
	}

	var submission Submission
	if current.Visibility == VisibilityPrivate {
		submission, err = handler.publishPrivateSubmission(ctx, user, current)
	} else {
		submission, err = handler.repo.Submit(ctx.Request.Context(), ctx.Param("id"), user.ID)
	}
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func (handler *Handler) DeleteMine(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	submission, err := handler.repo.DeleteByAuthor(ctx.Request.Context(), ctx.Param("id"), user.ID)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true, "submission": submission})
}

func (handler *Handler) requireSubmissionSettings(ctx *gin.Context) (operations.Settings, bool) {
	if handler.settings == nil {
		return operations.Settings{SubmissionsEnabled: true}, true
	}

	settings, err := handler.settings.GetSettings(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to load submission settings", "error", err, "path", ctx.FullPath())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load submission settings"})
		return operations.Settings{}, false
	}
	if !settings.SubmissionsEnabled {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "submissions are disabled"})
		return operations.Settings{}, false
	}

	return settings, true
}

func (handler *Handler) verifyTurnstile(ctx *gin.Context, settings operations.Settings, token string) bool {
	if !settings.TurnstileEnabled || !settings.TurnstileSubmission {
		return true
	}
	if strings.TrimSpace(settings.TurnstileSecretKey) == "" {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "turnstile is not configured"})
		return false
	}
	if strings.TrimSpace(token) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "turnstile token is required"})
		return false
	}

	ok, err := handler.turnstileVerifier.Verify(ctx.Request.Context(), settings.TurnstileSecretKey, token, ctx.ClientIP())
	if err != nil {
		slog.Error("turnstile verification failed", "error", err, "ip", ctx.ClientIP())
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "turnstile verification failed"})
		return false
	}
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "turnstile token is invalid"})
		return false
	}

	return true
}

func (handler *Handler) requireSubmissionSlot(ctx *gin.Context, userID string, settings operations.Settings, excludeID string) bool {
	since, limit, ok := submissionLimitWindow(settings.SubmissionLimit, time.Now())
	if !ok {
		return true
	}

	total, err := handler.repo.CountSubmittedSince(ctx.Request.Context(), userID, since, excludeID)
	if err != nil {
		slog.Error("failed to load submission limit", "error", err, "userID", userID, "since", since, "excludeID", excludeID)
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
		Status:   ctx.Query("status"),
		Keyword:  ctx.Query("q"),
		Sort:     ctx.Query("sort"),
		Page:     parsePositiveInt(ctx.Query("page")),
		PageSize: parsePositiveInt(ctx.Query("pageSize")),
		All:      boolQuery(ctx.Query("all")),
	})
	if err != nil {
		slog.Error("failed to load admin submissions", "error", err, "status", ctx.Query("status"), "keyword", ctx.Query("q"), "sort", ctx.Query("sort"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load submissions"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) AdminGet(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	submission, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, submission)
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

	handler.reviewWithRequest(ctx, reviewer, request)
}

func (handler *Handler) Approve(ctx *gin.Context) {
	reviewer, ok := auth.RequireAdmin(ctx)
	if !ok {
		return
	}

	var request ReviewRequest
	_ = ctx.ShouldBindJSON(&request)
	request.Action = ActionApprove
	handler.reviewWithRequest(ctx, reviewer, request)
}

func (handler *Handler) Reject(ctx *gin.Context) {
	reviewer, ok := auth.RequireAdmin(ctx)
	if !ok {
		return
	}

	var request ReviewRequest
	_ = ctx.ShouldBindJSON(&request)
	if strings.TrimSpace(request.Action) == "" {
		request.Action = ActionReject
	}
	handler.reviewWithRequest(ctx, reviewer, request)
}

func (handler *Handler) ArchivePublished(ctx *gin.Context) {
	reviewer, ok := auth.RequireAdmin(ctx)
	if !ok {
		return
	}
	if handler.archiver == nil {
		slog.Error("post archiver is unavailable", "submissionID", ctx.Param("id"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "post archiver is unavailable"})
		return
	}

	current, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}
	if current.Status != StatusPublished || strings.TrimSpace(current.PublishedPostSlug) == "" {
		handler.writeSubmissionError(ctx, ErrInvalidReview)
		return
	}
	if err := handler.archiver.Archive(ctx.Request.Context(), current.PublishedPostSlug); err != nil {
		if errors.Is(err, posts.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "published post not found"})
			return
		}
		slog.Error("failed to archive published post", "error", err, "submissionID", current.ID, "postSlug", current.PublishedPostSlug, "reviewerID", reviewer.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to archive published post"})
		return
	}

	submission, err := handler.repo.ArchivePublished(ctx.Request.Context(), ctx.Param("id"), reviewer)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}
	if handler.messages != nil {
		_, _ = handler.messages.Create(ctx.Request.Context(), messages.CreateRequest{
			RecipientID:   submission.AuthorID,
			RecipientName: submission.AuthorName,
			Type:          messages.TypeReview,
			Priority:      "normal",
			Title:         "你的文章已下架",
			Body:          fmt.Sprintf("《%s》已由管理员下架，不再公开展示。", submission.Title),
			TargetType:    "submission",
			TargetID:      submission.ID,
			TargetTitle:   submission.Title,
		}, reviewer)
	}

	ctx.JSON(http.StatusOK, submission)
}

func (handler *Handler) RestorePublished(ctx *gin.Context) {
	reviewer, ok := auth.RequireAdmin(ctx)
	if !ok {
		return
	}
	if handler.restorer == nil {
		slog.Error("post restorer is unavailable", "submissionID", ctx.Param("id"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "post restorer is unavailable"})
		return
	}

	current, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}
	if current.Status != StatusArchived || strings.TrimSpace(current.PublishedPostSlug) == "" {
		handler.writeSubmissionError(ctx, ErrInvalidReview)
		return
	}
	if err := handler.restorer.Restore(ctx.Request.Context(), current.PublishedPostSlug); err != nil {
		if errors.Is(err, posts.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "archived post not found"})
			return
		}
		slog.Error("failed to restore published post", "error", err, "submissionID", current.ID, "postSlug", current.PublishedPostSlug, "reviewerID", reviewer.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore published post"})
		return
	}

	submission, err := handler.repo.RestorePublished(ctx.Request.Context(), ctx.Param("id"), reviewer)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return
	}
	if handler.messages != nil {
		_, _ = handler.messages.Create(ctx.Request.Context(), messages.CreateRequest{
			RecipientID:   submission.AuthorID,
			RecipientName: submission.AuthorName,
			Type:          messages.TypeReview,
			Priority:      "normal",
			Title:         "你的文章已重新上架",
			Body:          fmt.Sprintf("《%s》已由管理员重新上架。", submission.Title),
			TargetType:    "submission",
			TargetID:      submission.ID,
			TargetTitle:   submission.Title,
		}, reviewer)
	}

	ctx.JSON(http.StatusOK, submission)
}

func (handler *Handler) reviewWithRequest(ctx *gin.Context, reviewer auth.User, request ReviewRequest) {
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
			slog.Error("post publisher is unavailable", "submissionID", submission.ID, "reviewerID", reviewer.ID)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "post publisher is unavailable"})
			return
		}

		post, err := handler.publisher.PublishSubmission(ctx.Request.Context(), posts.PublishInput{
			Slug:       defaultString(strings.TrimSpace(request.Slug), submission.Slug),
			Title:      submission.Title,
			Summary:    submission.Summary,
			Content:    submission.Content,
			Visibility: posts.VisibilityPublic,
			Category:   defaultString(strings.TrimSpace(request.Category), submission.Category),
			Tags:       submission.Tags,
			CoverImage: submission.CoverImage,
			AuthorID:   submission.AuthorID,
			AuthorName: submission.AuthorName,
		}, submission.PublishedPostSlug)
		if err != nil {
			slog.Error("failed to publish submission", "error", err, "submissionID", submission.ID, "reviewerID", reviewer.ID, "slug", defaultString(strings.TrimSpace(request.Slug), submission.Slug))
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
		slog.Error("failed to create review message", "error", err, "submissionID", submission.ID, "reviewerID", reviewer.ID, "recipientID", submission.AuthorID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review message"})
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func (handler *Handler) publishPrivateSubmission(ctx *gin.Context, user auth.User, submission Submission) (Submission, error) {
	if handler.publisher == nil {
		slog.Error("post publisher is unavailable for private submission", "submissionID", submission.ID, "userID", user.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "post publisher is unavailable"})
		return Submission{}, ErrInvalidSubmission
	}
	if err := validateSubmissionReady(submission); err != nil {
		handler.writeSubmissionError(ctx, err)
		return Submission{}, err
	}

	post, err := handler.publisher.PublishSubmission(ctx.Request.Context(), posts.PublishInput{
		Slug:       submission.Slug,
		Title:      submission.Title,
		Summary:    submission.Summary,
		Content:    submission.Content,
		Visibility: posts.VisibilityPrivate,
		Category:   submission.Category,
		Tags:       submission.Tags,
		CoverImage: submission.CoverImage,
		AuthorID:   submission.AuthorID,
		AuthorName: defaultString(strings.TrimSpace(submission.AuthorName), user.DisplayName),
	}, submission.PublishedPostSlug)
	if err != nil {
		slog.Error("failed to publish private article", "error", err, "submissionID", submission.ID, "userID", user.ID, "slug", submission.Slug)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish private article"})
		return Submission{}, err
	}

	published, err := handler.repo.MarkPublished(ctx.Request.Context(), submission.ID, user.ID, post.Slug)
	if err != nil {
		handler.writeSubmissionError(ctx, err)
		return Submission{}, err
	}

	return published, nil
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
		slog.Error("failed to update submission", "error", err, "path", ctx.FullPath(), "submissionID", ctx.Param("id"))
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
	return (user.Status == "" || user.Status == "active") && user.EmailVerified
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

func parsePositiveInt(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return 0
	}

	return parsed
}

func boolQuery(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "1" || value == "true" || value == "yes"
}
