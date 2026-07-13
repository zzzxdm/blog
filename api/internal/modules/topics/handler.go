package topics

import (
	"blog/api/internal/httpx"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/posts"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo     Repository
	postRepo posts.Repository
}

func NewHandler(repo Repository, postRepo posts.Repository) *Handler {
	return &Handler{repo: repo, postRepo: postRepo}
}

func RegisterRoutes(router gin.IRouter, repo Repository, postRepo posts.Repository) {
	handler := NewHandler(repo, postRepo)

	router.GET("/topics", handler.ListPublic)
	router.GET("/topics/:slug/posts", handler.ListTopicPosts)
	router.GET("/topics/:slug", handler.GetPublic)
	router.GET("/admin/topics", handler.ListAdmin)
	router.POST("/admin/topics", handler.Create)
	router.PUT("/admin/topics/:id", handler.Update)
	router.DELETE("/admin/topics/:id", handler.Delete)
}

func (handler *Handler) ListPublic(ctx *gin.Context) {
	handler.list(ctx, false)
}

func (handler *Handler) ListAdmin(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	handler.list(ctx, true)
}

func (handler *Handler) list(ctx *gin.Context, admin bool) {
	query := ListQuery{
		Keyword:  strings.TrimSpace(ctx.Query("q")),
		Status:   strings.TrimSpace(ctx.Query("status")),
		Featured: boolQuery(ctx.Query("featured")),
		All:      admin || boolQuery(ctx.Query("all")),
		Page:     intQuery(ctx.Query("page"), 1),
		PageSize: intQuery(ctx.Query("pageSize"), 10),
	}
	if !admin {
		query.All = false
	}

	result, err := handler.repo.List(ctx.Request.Context(), query)
	if err != nil {
		slog.Error("failed to load topics", "error", err, "admin", admin, "status", query.Status, "keyword", query.Keyword, "featured", query.Featured)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load topics"})
		return
	}

	handler.enrichTopics(ctx.Request.Context(), result.Items)
	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) GetPublic(ctx *gin.Context) {
	topic, err := handler.repo.GetBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		handler.writeError(ctx, err)
		return
	}
	if topic.Status != "active" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "topic not found"})
		return
	}

	handler.enrichTopic(ctx.Request.Context(), &topic)
	ctx.JSON(http.StatusOK, topic)
}

func (handler *Handler) ListTopicPosts(ctx *gin.Context) {
	topic, err := handler.repo.GetBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		handler.writeError(ctx, err)
		return
	}
	if topic.Status != "active" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "topic not found"})
		return
	}

	page := normalizePage(intQuery(ctx.Query("page"), 1))
	pageSize := normalizePageSize(intQuery(ctx.Query("pageSize"), 10))
	allPosts, err := handler.loadAllPublishedPosts(ctx.Request.Context())
	if err != nil {
		slog.Error("failed to load topic posts", "error", err, "topicSlug", ctx.Param("slug"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load topic posts"})
		return
	}

	filtered := filterTopicPosts(allPosts, topic)
	ctx.JSON(http.StatusOK, posts.ListResult{
		Items:    pageItems(filtered, page, pageSize),
		Page:     page,
		PageSize: pageSize,
		Total:    len(filtered),
	})
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
	if !httpx.BindJSON(ctx, &request, "invalid topic payload") {
		return
	}

	topic, err := handler.repo.Save(ctx.Request.Context(), id, request)
	if err != nil {
		handler.writeError(ctx, err)
		return
	}
	handler.enrichTopic(ctx.Request.Context(), &topic)

	if id == "" {
		ctx.JSON(http.StatusCreated, topic)
		return
	}
	ctx.JSON(http.StatusOK, topic)
}

func (handler *Handler) Delete(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	if err := handler.repo.Delete(ctx.Request.Context(), ctx.Param("id")); err != nil {
		handler.writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (handler *Handler) enrichTopics(ctx context.Context, items []Topic) {
	if len(items) == 0 {
		return
	}

	allPosts, err := handler.loadAllPublishedPosts(ctx)
	if err != nil {
		slog.Warn("failed to enrich topics with posts", "error", err)
		return
	}

	for index := range items {
		enrichTopicWithPosts(&items[index], allPosts)
	}
}

func (handler *Handler) enrichTopic(ctx context.Context, topic *Topic) {
	allPosts, err := handler.loadAllPublishedPosts(ctx)
	if err != nil {
		slog.Warn("failed to enrich topic with posts", "error", err, "topicSlug", topic.Slug)
		return
	}

	enrichTopicWithPosts(topic, allPosts)
}

func (handler *Handler) loadAllPublishedPosts(ctx context.Context) ([]posts.Post, error) {
	if handler.postRepo == nil {
		return []posts.Post{}, nil
	}

	allPosts := []posts.Post{}
	for page := 1; ; page++ {
		result, err := handler.postRepo.List(ctx, posts.ListQuery{Page: page, PageSize: 50})
		if err != nil {
			return nil, err
		}
		allPosts = append(allPosts, result.Items...)
		if len(allPosts) >= result.Total || len(result.Items) == 0 {
			break
		}
	}

	sort.SliceStable(allPosts, func(i, j int) bool {
		return allPosts[i].PublishedAt.After(allPosts[j].PublishedAt)
	})
	return allPosts, nil
}

func enrichTopicWithPosts(topic *Topic, allPosts []posts.Post) {
	filtered := filterTopicPosts(allPosts, *topic)
	topic.PostCount = len(filtered)
	topic.LatestPostAt = nil
	if len(filtered) > 0 {
		latest := filtered[0].PublishedAt
		topic.LatestPostAt = &latest
	}
}

func filterTopicPosts(allPosts []posts.Post, topic Topic) []posts.Post {
	filtered := make([]posts.Post, 0)
	for _, post := range allPosts {
		if matchesTopic(post, topic) {
			filtered = append(filtered, post)
		}
	}
	return filtered
}

func matchesTopic(post posts.Post, topic Topic) bool {
	for _, category := range topic.Categories {
		if strings.EqualFold(post.Category, category) {
			return true
		}
	}

	for _, topicTag := range topic.Tags {
		if hasPostTag(post.Tags, topicTag) {
			return true
		}
		normalizedTag := strings.ToLower(strings.TrimSpace(topicTag))
		if normalizedTag != "" && (strings.Contains(strings.ToLower(post.Title), normalizedTag) || strings.Contains(strings.ToLower(post.Summary), normalizedTag)) {
			return true
		}
	}

	return false
}

func hasPostTag(tags []string, target string) bool {
	for _, tag := range tags {
		if strings.EqualFold(tag, target) {
			return true
		}
	}
	return false
}

func (handler *Handler) writeError(ctx *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "topic not found"})
		return
	}
	if errors.Is(err, ErrInvalid) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "topic title is required"})
		return
	}
	if errors.Is(err, ErrDuplicate) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "topic slug or title already exists"})
		return
	}

	slog.Error("failed to update topic", "error", err, "id", ctx.Param("id"), "slug", ctx.Param("slug"), "path", ctx.FullPath())
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update topic"})
}

func intQuery(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func boolQuery(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
