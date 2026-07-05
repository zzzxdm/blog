package adminposts

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/posts"

	"github.com/gin-gonic/gin"
)

const previewTokenTTL = 30 * time.Minute
const previewTokenSecret = "blog-admin-post-preview-v1"

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
	router.GET("/admin/posts/:id/revisions", handler.ListRevisions)
	router.GET("/preview/:token", handler.GetPreview)
	router.POST("/admin/posts", handler.Create)
	router.PUT("/admin/posts/:id", handler.Update)
	router.DELETE("/admin/posts/:id", handler.Delete)
	router.POST("/admin/posts/:id/preview", handler.CreatePreview)
	router.POST("/admin/posts/:id/publish", handler.Publish)
	router.POST("/admin/posts/:id/revisions/:revisionId/restore", handler.RestoreRevision)
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

func (handler *Handler) CreatePreview(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	post, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeError(ctx, err)
		return
	}

	expiresAt := time.Now().Add(previewTokenTTL)
	token := previewToken(post.ID, expiresAt)
	ctx.JSON(http.StatusOK, PreviewResult{
		PreviewURL: "/preview/" + token,
		Token:      token,
		ExpiresAt:  expiresAt,
	})
}

func (handler *Handler) GetPreview(ctx *gin.Context) {
	postID, ok := parsePreviewToken(ctx.Param("token"), time.Now())
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "preview not found or expired"})
		return
	}

	post, err := handler.repo.Get(ctx.Request.Context(), postID)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "preview not found or expired"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load preview"})
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

func (handler *Handler) Delete(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	post, err := handler.repo.Delete(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true, "post": post})
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

func (handler *Handler) ListRevisions(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.ListRevisions(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) RestoreRevision(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	post, err := handler.repo.RestoreRevision(ctx.Request.Context(), ctx.Param("id"), ctx.Param("revisionId"))
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
	if errors.Is(err, ErrRevisionNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "admin post revision not found"})
		return
	}
	if errors.Is(err, ErrInvalidPost) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "title and content are required"})
		return
	}
	if errors.Is(err, ErrPostNotPublic) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "非公开文章暂不支持发布到公开站点"})
		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update admin post"})
}

func previewToken(postID string, expiresAt time.Time) string {
	payload := fmt.Sprintf("%s|%d", postID, expiresAt.Unix())
	signature := previewSignature(payload)
	return base64.RawURLEncoding.EncodeToString([]byte(payload + "|" + signature))
}

func parsePreviewToken(token string, now time.Time) (string, bool) {
	data, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return "", false
	}

	parts := strings.Split(string(data), "|")
	if len(parts) != 3 {
		return "", false
	}

	payload := parts[0] + "|" + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(previewSignature(payload))) {
		return "", false
	}

	expiresUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || !time.Unix(expiresUnix, 0).After(now) {
		return "", false
	}

	return parts[0], true
}

func previewSignature(payload string) string {
	mac := hmac.New(sha256.New, []byte(previewTokenSecret))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
