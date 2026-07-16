package operations

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const maxMediaUploadBytes = 10 << 20

func (handler *Handler) storeJob(job AdminJob) {
	job.UpdatedAt = time.Now()
	handler.jobsMu.Lock()
	handler.jobs[job.ID] = job
	handler.jobsMu.Unlock()
}

type Handler struct {
	repo    Repository
	storage MediaStorage
	jobsMu  sync.RWMutex
	jobs    map[string]AdminJob
}

func RegisterRoutes(router gin.IRouter, repo Repository, storage MediaStorage) {
	handler := NewHandler(repo, storage)

	router.GET("/settings", handler.GetPublicSettings)
	router.GET("/navigation", handler.GetPublicNavigation)
	router.POST("/media/uploads", handler.UploadUserMedia)
	router.GET("/media/:id/file", handler.ResolveMediaFile)
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
	router.GET("/admin/media/:id/references", handler.ListMediaReferences)
	router.PATCH("/admin/media/:id", handler.UpdateMedia)
	router.PUT("/admin/media/:id/file", handler.ReplaceMediaFile)
	router.DELETE("/admin/media/:id", handler.DeleteMedia)
	router.GET("/admin/stats", handler.GetStats)
	router.GET("/admin/statistics", handler.GetStats)
	router.GET("/admin/stats/export", handler.ExportStats)
	router.POST("/admin/import", handler.CreateImportJob)
	router.POST("/admin/export", handler.CreateExportJob)
	router.GET("/admin/jobs/:id", handler.GetJob)
	router.GET("/admin/audit-logs", handler.ListAuditLogs)
}

func NewHandler(repo Repository, storage MediaStorage) *Handler {
	if storage == nil {
		storage = NewLocalMediaStorage("uploads")
	}

	return &Handler{repo: repo, storage: storage, jobs: map[string]AdminJob{}}
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
