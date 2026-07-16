package operations

import (
	"blog/api/internal/httpx"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"blog/api/internal/idgen"
	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

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

func safeOriginalName(name string) string {
	fileName := strings.TrimSpace(filepath.Base(name))
	if fileName == "" || fileName == "." {
		return "upload"
	}

	return strings.NewReplacer("/", "-", "\\", "-", ":", "-").Replace(fileName)
}

func (handler *Handler) uploadMedia(ctx *gin.Context, user auth.User, defaultCategory string) {
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
	relativeDir := now.Format("2006/01/02")
	originalName := safeOriginalName(fileHeader.Filename)
	storedName := uniqueStoredName(originalName, extension, now)
	objectKey := filepath.ToSlash(filepath.Join(relativeDir, storedName))
	mediaURL, err := handler.storage.Save(ctx.Request.Context(), objectKey, file, fileHeader.Size, contentType)
	if err != nil {
		slog.Error("failed to save upload", "error", err, "userID", user.ID, "fileName", fileHeader.Filename, "objectKey", objectKey, "contentType", contentType, "size", fileHeader.Size)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save upload"})
		return
	}

	alt := strings.TrimSpace(ctx.PostForm("alt"))
	if alt == "" {
		alt = strings.TrimSuffix(originalName, filepath.Ext(originalName))
	}
	category := strings.TrimSpace(ctx.PostForm("category"))
	if category == "" {
		category = defaultCategory
	}

	asset := MediaAsset{
		ID:         idgen.NextString(),
		FileName:   originalName,
		URL:        mediaURL,
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
		if deleteErr := handler.storage.Delete(ctx.Request.Context(), mediaURL); deleteErr != nil {
			slog.Warn("failed to delete uploaded media after record failure", "error", deleteErr, "url", mediaURL, "userID", user.ID)
		}
		slog.Error("failed to record media", "error", err, "userID", user.ID, "fileName", originalName, "url", mediaURL)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record media"})
		return
	}

	created.URL = MediaAssetURL(created.ID)
	ctx.JSON(http.StatusCreated, created)
}

func (handler *Handler) ListMedia(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.ListMedia(ctx.Request.Context(), MediaListQuery{
		Keyword:  ctx.Query("q"),
		Type:     ctx.Query("type"),
		Sort:     ctx.Query("sort"),
		Page:     parsePositiveInt(ctx.Query("page")),
		PageSize: parsePositiveInt(ctx.Query("pageSize")),
		All:      boolQuery(ctx.Query("all")),
	})
	if err != nil {
		slog.Error("failed to load media", "error", err, "keyword", ctx.Query("q"), "type", ctx.Query("type"), "sort", ctx.Query("sort"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load media"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) UploadUserMedia(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	handler.uploadMedia(ctx, user, "写作插图")
}

func parsePositiveInt(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return 0
	}

	return parsed
}

func isWebP(header []byte) bool {
	return len(header) >= 12 &&
		string(header[:4]) == "RIFF" &&
		string(header[8:12]) == "WEBP"
}

func uniqueStoredName(originalName string, extension string, now time.Time) string {
	base := strings.TrimSuffix(filepath.Base(originalName), filepath.Ext(originalName))
	base = asciiSlug(base)
	if base == "" {
		base = "asset"
	}

	return fmt.Sprintf("%s-%d%s", base, now.UnixNano(), extension)
}

func (handler *Handler) ResolveMediaFile(ctx *gin.Context) {
	asset, err := handler.repo.GetMediaFile(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeMediaError(ctx, err)
		return
	}

	ctx.Redirect(http.StatusFound, asset.URL)
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

func (handler *Handler) writeMediaError(ctx *gin.Context, err error) {
	if errors.Is(err, ErrMediaNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "media asset not found"})
		return
	}
	if errors.Is(err, ErrMediaInUse) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "media asset is in use"})
		return
	}

	slog.Error("failed to update media", "error", err, "path", ctx.FullPath(), "mediaID", ctx.Param("id"))
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update media"})
}

func detectMediaType(file multipart.File, _ string) (string, string, error) {
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

	if isWebP(header[:size]) {
		return "image/webp", ".webp", nil
	}

	return "", "", errors.New("unsupported media type")
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

	if err := handler.storage.Delete(ctx.Request.Context(), asset.URL); err != nil {
		slog.Warn("failed to delete media file", "error", err, "mediaID", asset.ID, "url", asset.URL)
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true, "asset": asset})
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

func boolQuery(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "1" || value == "true" || value == "yes"
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

func (handler *Handler) ListMediaReferences(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	result, err := handler.repo.ListMediaReferences(ctx.Request.Context(), ctx.Param("id"), parsePositiveInt(ctx.Query("page")), parsePositiveInt(ctx.Query("pageSize")))
	if err != nil {
		handler.writeMediaError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) UploadMedia(ctx *gin.Context) {
	user, ok := auth.RequireAdmin(ctx)
	if !ok {
		return
	}

	handler.uploadMedia(ctx, user, "上传")
}

func (handler *Handler) ReplaceMediaFile(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	current, err := handler.repo.GetMediaFile(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handler.writeMediaError(ctx, err)
		return
	}
	if current.Type != "image" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "only image assets can be replaced"})
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
	if mediaKind(contentType) != "image" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "replacement file must be an image"})
		return
	}
	width, height := mediaDimensions(file, contentType)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read upload"})
		return
	}

	now := time.Now()
	relativeDir := now.Format("2006/01/02")
	originalName := safeOriginalName(fileHeader.Filename)
	storedName := uniqueStoredName(originalName, extension, now)
	objectKey := filepath.ToSlash(filepath.Join(relativeDir, storedName))
	mediaURL, err := handler.storage.Save(ctx.Request.Context(), objectKey, file, fileHeader.Size, contentType)
	if err != nil {
		slog.Error("failed to save replacement media file", "error", err, "mediaID", current.ID, "fileName", fileHeader.Filename, "objectKey", objectKey)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replace media file"})
		return
	}

	updated, err := handler.repo.UpdateMediaFile(ctx.Request.Context(), current.ID, MediaFileUpdateRequest{
		FileName:  originalName,
		URL:       mediaURL,
		Type:      mediaKind(contentType),
		SizeLabel: formatBytes(fileHeader.Size),
		Width:     width,
		Height:    height,
	})
	if err != nil {
		if deleteErr := handler.storage.Delete(ctx.Request.Context(), mediaURL); deleteErr != nil {
			slog.Warn("failed to delete replacement media after record failure", "error", deleteErr, "url", mediaURL, "mediaID", current.ID)
		}
		handler.writeMediaError(ctx, err)
		return
	}
	if err := handler.storage.Delete(ctx.Request.Context(), current.URL); err != nil {
		slog.Warn("failed to delete previous media file after replacement", "error", err, "mediaID", current.ID, "url", current.URL)
	}

	ctx.JSON(http.StatusOK, updated)
}

func (handler *Handler) UpdateMedia(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request MediaUpdateRequest
	if !httpx.BindJSON(ctx, &request, "invalid media metadata payload") {
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

func mediaKind(contentType string) string {
	if contentType == "application/pdf" {
		return "document"
	}

	return "image"
}
