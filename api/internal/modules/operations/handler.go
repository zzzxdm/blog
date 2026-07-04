package operations

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

const maxMediaUploadBytes = 10 << 20

type Handler struct {
	repo      Repository
	uploadDir string
}

func NewHandler(repo Repository, uploadDir string) *Handler {
	if uploadDir == "" {
		uploadDir = "uploads"
	}

	return &Handler{repo: repo, uploadDir: uploadDir}
}

func RegisterRoutes(router gin.IRouter, repo Repository, uploadDir string) {
	handler := NewHandler(repo, uploadDir)

	router.GET("/admin/settings", handler.GetSettings)
	router.PUT("/admin/settings", handler.UpdateSettings)
	router.POST("/admin/settings/test-mail", handler.SendTestMail)
	router.POST("/admin/backups", handler.RunBackup)
	router.GET("/admin/navigation", handler.GetNavigation)
	router.PUT("/admin/navigation", handler.UpdateNavigation)
	router.GET("/admin/media", handler.ListMedia)
	router.POST("/admin/media", handler.UploadMedia)
	router.GET("/admin/media/:id", handler.GetMedia)
	router.PATCH("/admin/media/:id", handler.UpdateMedia)
	router.DELETE("/admin/media/:id", handler.DeleteMedia)
	router.GET("/admin/stats", handler.GetStats)
	router.GET("/admin/stats/export", handler.ExportStats)
	router.GET("/admin/audit-logs", handler.ListAuditLogs)
}

func (handler *Handler) GetSettings(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	settings, err := handler.repo.GetSettings(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (handler *Handler) UpdateSettings(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request Settings
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid settings payload"})
		return
	}

	settings, err := handler.repo.UpdateSettings(ctx.Request.Context(), request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (handler *Handler) SendTestMail(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.SendTestMail(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test mail"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) RunBackup(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.RunBackup(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run backup"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) GetNavigation(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	navigation, err := handler.repo.GetNavigation(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load navigation"})
		return
	}

	ctx.JSON(http.StatusOK, navigation)
}

func (handler *Handler) UpdateNavigation(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request Navigation
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid navigation payload"})
		return
	}

	navigation, err := handler.repo.UpdateNavigation(ctx.Request.Context(), request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update navigation"})
		return
	}

	ctx.JSON(http.StatusOK, navigation)
}

func (handler *Handler) ListMedia(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.ListMedia(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load media"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) GetMedia(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	asset, err := handler.repo.GetMedia(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeMediaError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, asset)
}

func (handler *Handler) UploadMedia(ctx *gin.Context) {
	user, ok := auth.RequireAdmin(ctx)
	if !ok {
		return
	}

	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxMediaUploadBytes+(1<<20))

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxMediaUploadBytes {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file size must be between 1 byte and 10 MB"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read upload"})
		return
	}
	defer file.Close()

	contentType, extension, err := detectMediaType(file, fileHeader.Filename)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	width, height := mediaDimensions(file, contentType)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read upload"})
		return
	}

	now := time.Now()
	relativeDir := now.Format("2006/01")
	originalName := safeOriginalName(fileHeader.Filename)
	storedName := uniqueStoredName(originalName, extension, now)
	targetDir := filepath.Join(handler.uploadDir, relativeDir)
	targetPath := filepath.Join(targetDir, storedName)

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare upload directory"})
		return
	}

	destination, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save upload"})
		return
	}
	if _, err := io.Copy(destination, file); err != nil {
		_ = destination.Close()
		_ = os.Remove(targetPath)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save upload"})
		return
	}
	if err := destination.Close(); err != nil {
		_ = os.Remove(targetPath)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save upload"})
		return
	}

	alt := strings.TrimSpace(ctx.PostForm("alt"))
	if alt == "" {
		alt = strings.TrimSuffix(originalName, filepath.Ext(originalName))
	}
	category := strings.TrimSpace(ctx.PostForm("category"))
	if category == "" {
		category = "上传"
	}

	asset := MediaAsset{
		ID:         fmt.Sprintf("media_%d", now.UnixNano()),
		FileName:   originalName,
		URL:        "/uploads/" + filepath.ToSlash(filepath.Join(relativeDir, storedName)),
		Alt:        alt,
		Type:       mediaKind(contentType),
		Category:   category,
		SizeLabel:  formatBytes(fileHeader.Size),
		Width:      width,
		Height:     height,
		UsageCount: 0,
		UploadedBy: user.DisplayName,
		UploadedAt: now,
	}

	created, err := handler.repo.CreateMedia(ctx.Request.Context(), asset)
	if err != nil {
		_ = os.Remove(targetPath)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record media"})
		return
	}

	ctx.JSON(http.StatusCreated, created)
}

func (handler *Handler) UpdateMedia(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request MediaUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid media metadata payload"})
		return
	}

	request.Alt = strings.TrimSpace(request.Alt)
	request.Category = strings.TrimSpace(request.Category)
	if request.Category == "" {
		request.Category = "未分类"
	}

	asset, err := handler.repo.UpdateMedia(ctx.Request.Context(), ctx.Param("id"), request)
	if err != nil {
		handler.writeMediaError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, asset)
}

func (handler *Handler) DeleteMedia(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	asset, err := handler.repo.DeleteMedia(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeMediaError(ctx, err)
		return
	}

	_ = handler.removeLocalUpload(asset)
	ctx.JSON(http.StatusOK, gin.H{"ok": true, "asset": asset})
}

func (handler *Handler) GetStats(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	stats, err := handler.repo.GetStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

func (handler *Handler) ExportStats(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	stats, err := handler.repo.GetStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export stats"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"scope":      "stats",
		"exportedAt": time.Now(),
		"stats":      stats,
	})
}

func (handler *Handler) ListAuditLogs(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.ListAuditLogs(ctx.Request.Context(), AuditLogQuery{
		Action:       ctx.Query("action"),
		ResourceType: ctx.Query("resourceType"),
		Page:         parsePositiveInt(ctx.Query("page")),
		PageSize:     parsePositiveInt(ctx.Query("pageSize")),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load audit logs"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) writeMediaError(ctx *gin.Context, err error) {
	if errors.Is(err, ErrMediaNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "media asset not found"})
		return
	}
	if errors.Is(err, ErrMediaInUse) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "media asset is in use"})
		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update media"})
}

func (handler *Handler) removeLocalUpload(asset MediaAsset) error {
	if !strings.HasPrefix(asset.URL, "/uploads/") {
		return nil
	}

	relativePath := strings.TrimPrefix(asset.URL, "/uploads/")
	targetPath := filepath.Join(handler.uploadDir, filepath.FromSlash(relativePath))
	root, err := filepath.Abs(handler.uploadDir)
	if err != nil {
		return err
	}
	target, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	if target == root || !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return nil
	}

	if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func detectMediaType(file multipart.File, fileName string) (string, string, error) {
	header := make([]byte, 512)
	size, err := file.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", "", errors.New("failed to inspect upload")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", errors.New("failed to inspect upload")
	}

	contentType := http.DetectContentType(header[:size])
	if extension, ok := mediaContentTypes()[contentType]; ok {
		return contentType, extension, nil
	}

	extension := strings.ToLower(filepath.Ext(fileName))
	if contentType, ok := mediaExtensions()[extension]; ok {
		return contentType, normalizedMediaExtension(extension), nil
	}

	return "", "", errors.New("unsupported media type")
}

func mediaDimensions(file multipart.File, contentType string) (int, int) {
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" {
		return 0, 0
	}

	config, _, err := image.DecodeConfig(file)
	_, _ = file.Seek(0, io.SeekStart)
	if err != nil {
		return 0, 0
	}

	return config.Width, config.Height
}

func mediaContentTypes() map[string]string {
	return map[string]string{
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"image/webp":      ".webp",
		"image/gif":       ".gif",
		"application/pdf": ".pdf",
	}
}

func mediaExtensions() map[string]string {
	return map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
	}
}

func normalizedMediaExtension(extension string) string {
	if extension == ".jpeg" {
		return ".jpg"
	}

	return extension
}

func mediaKind(contentType string) string {
	if contentType == "application/pdf" {
		return "document"
	}

	return "image"
}

func safeOriginalName(name string) string {
	fileName := strings.TrimSpace(filepath.Base(name))
	if fileName == "" || fileName == "." {
		return "upload"
	}

	return strings.NewReplacer("/", "-", "\\", "-", ":", "-").Replace(fileName)
}

func uniqueStoredName(originalName string, extension string, now time.Time) string {
	base := strings.TrimSuffix(filepath.Base(originalName), filepath.Ext(originalName))
	base = asciiSlug(base)
	if base == "" {
		base = "asset"
	}

	return fmt.Sprintf("%s-%d%s", base, now.UnixNano(), extension)
}

func asciiSlug(value string) string {
	var builder strings.Builder
	lastDash := false

	for _, r := range strings.ToLower(value) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case !lastDash:
			builder.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func formatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	kb := float64(size) / unit
	if kb < unit {
		return fmt.Sprintf("%.1f KB", kb)
	}

	return fmt.Sprintf("%.1f MB", kb/unit)
}

func parsePositiveInt(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return 0
	}

	return parsed
}
