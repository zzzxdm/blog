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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"blog/api/internal/idgen"
	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

const maxMediaUploadBytes = 10 << 20

type Handler struct {
	repo    Repository
	storage MediaStorage
	jobsMu  sync.RWMutex
	jobs    map[string]AdminJob
}

func NewHandler(repo Repository, storage MediaStorage) *Handler {
	if storage == nil {
		storage = NewLocalMediaStorage("uploads")
	}

	return &Handler{repo: repo, storage: storage, jobs: map[string]AdminJob{}}
}

func RegisterRoutes(router gin.IRouter, repo Repository, storage MediaStorage) {
	handler := NewHandler(repo, storage)

	router.GET("/settings", handler.GetPublicSettings)
	router.GET("/navigation", handler.GetPublicNavigation)
	router.POST("/media/uploads", handler.UploadUserMedia)
	router.GET("/admin/settings", handler.GetSettings)
	router.PUT("/admin/settings", handler.UpdateSettings)
	router.POST("/admin/settings/test-mail", handler.SendTestMail)
	router.POST("/admin/backups", handler.RunBackup)
	router.GET("/admin/navigation", handler.GetNavigation)
	router.PUT("/admin/navigation", handler.UpdateNavigation)
	router.GET("/admin/redirects", handler.ListRedirects)
	router.POST("/admin/redirects", handler.CreateRedirect)
	router.PUT("/admin/redirects", handler.ReplaceRedirects)
	router.GET("/admin/media", handler.ListMedia)
	router.POST("/admin/media", handler.UploadMedia)
	router.GET("/admin/media/:id", handler.GetMedia)
	router.PATCH("/admin/media/:id", handler.UpdateMedia)
	router.DELETE("/admin/media/:id", handler.DeleteMedia)
	router.GET("/admin/stats", handler.GetStats)
	router.GET("/admin/statistics", handler.GetStats)
	router.GET("/admin/stats/export", handler.ExportStats)
	router.POST("/admin/import", handler.CreateImportJob)
	router.POST("/admin/export", handler.CreateExportJob)
	router.GET("/admin/jobs/:id", handler.GetJob)
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

func (handler *Handler) GetPublicSettings(ctx *gin.Context) {
	settings, err := handler.repo.GetSettings(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	ctx.JSON(http.StatusOK, publicSettings(settings))
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

	handler.writeNavigation(ctx)
}

func (handler *Handler) GetPublicNavigation(ctx *gin.Context) {
	handler.writeNavigation(ctx)
}

func (handler *Handler) writeNavigation(ctx *gin.Context) {
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

func (handler *Handler) ListRedirects(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	navigation, err := handler.repo.GetNavigation(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load redirects"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items": navigation.Redirects,
		"total": len(navigation.Redirects),
	})
}

func (handler *Handler) CreateRedirect(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request RedirectRule
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid redirect payload"})
		return
	}

	navigation, err := handler.repo.GetNavigation(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load redirects"})
		return
	}

	updatedRedirects := normalizeRedirects(append(navigation.Redirects, request))
	if len(updatedRedirects) == len(navigation.Redirects) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "redirect from and to are required"})
		return
	}

	navigation.Redirects = updatedRedirects
	navigation, err = handler.repo.UpdateNavigation(ctx.Request.Context(), navigation)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save redirect"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"item":  navigation.Redirects[len(navigation.Redirects)-1],
		"items": navigation.Redirects,
		"total": len(navigation.Redirects),
	})
}

func (handler *Handler) ReplaceRedirects(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request struct {
		Items []RedirectRule `json:"items"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid redirects payload"})
		return
	}

	navigation, err := handler.repo.GetNavigation(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load redirects"})
		return
	}

	navigation.Redirects = normalizeRedirects(request.Items)
	navigation, err = handler.repo.UpdateNavigation(ctx.Request.Context(), navigation)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save redirects"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items": navigation.Redirects,
		"total": len(navigation.Redirects),
	})
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

	handler.uploadMedia(ctx, user, "上传")
}

func (handler *Handler) UploadUserMedia(ctx *gin.Context) {
	user, ok := auth.RequireUser(ctx)
	if !ok {
		return
	}

	handler.uploadMedia(ctx, user, "写作插图")
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
		_ = handler.storage.Delete(ctx.Request.Context(), mediaURL)
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

	_ = handler.storage.Delete(ctx.Request.Context(), asset.URL)
	ctx.JSON(http.StatusOK, gin.H{"ok": true, "asset": asset})
}

func (handler *Handler) GetStats(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	stats, err := handler.repo.GetStats(ctx.Request.Context(), ctx.Query("range"))
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

	stats, err := handler.repo.GetStats(ctx.Request.Context(), ctx.Query("range"))
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

func (handler *Handler) CreateImportJob(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request AdminJobRequest
	_ = ctx.ShouldBindJSON(&request)
	job := handler.createJob("import", request.Scope, request.FileName)
	job.Status = "queued"
	job.Progress = 10
	job.Message = "导入任务已创建，等待离线校验和执行。"
	handler.storeJob(job)

	ctx.JSON(http.StatusAccepted, job)
}

func (handler *Handler) CreateExportJob(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request AdminJobRequest
	_ = ctx.ShouldBindJSON(&request)
	job := handler.createJob("export", request.Scope, request.FileName)
	job.Status = "completed"
	job.Progress = 100
	job.Message = "导出任务已完成。"
	job.DownloadURL = "/api/admin/jobs/" + job.ID
	handler.storeJob(job)

	ctx.JSON(http.StatusAccepted, job)
}

func (handler *Handler) GetJob(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	handler.jobsMu.RLock()
	job, ok := handler.jobs[ctx.Param("id")]
	handler.jobsMu.RUnlock()
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	ctx.JSON(http.StatusOK, job)
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

func (handler *Handler) createJob(kind string, scope string, fileName string) AdminJob {
	now := time.Now()
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = "site"
	}
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = fmt.Sprintf("%s-%s-%s.json", kind, scope, now.Format("20060102-150405"))
	}

	return AdminJob{
		ID:        fmt.Sprintf("job_%d", now.UnixNano()),
		Type:      kind,
		Scope:     scope,
		Status:    "queued",
		Progress:  0,
		Message:   "任务已创建。",
		FileName:  fileName,
		Result:    map[string]any{"scope": scope},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (handler *Handler) storeJob(job AdminJob) {
	job.UpdatedAt = time.Now()
	handler.jobsMu.Lock()
	handler.jobs[job.ID] = job
	handler.jobsMu.Unlock()
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

func isWebP(header []byte) bool {
	return len(header) >= 12 &&
		string(header[:4]) == "RIFF" &&
		string(header[8:12]) == "WEBP"
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

func boolQuery(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "1" || value == "true" || value == "yes"
}
