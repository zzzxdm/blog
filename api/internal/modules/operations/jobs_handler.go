package operations

import (
	"blog/api/internal/httpx"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"net/http"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) CreateExportJob(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request AdminJobRequest
	if !httpx.BindJSON(ctx, &request, "invalid export job payload") {
		return
	}
	job := handler.createJob("export", request.Scope, request.FileName)
	job.Status = "completed"
	job.Progress = 100
	job.Message = "导出任务已完成。"
	job.DownloadURL = "/api/admin/jobs/" + job.ID
	handler.storeJob(job)

	ctx.JSON(http.StatusAccepted, job)
}

func (handler *Handler) CreateImportJob(ctx *gin.Context) {
	if _, ok := auth.RequireAdmin(ctx); !ok {
		return
	}

	var request AdminJobRequest
	if !httpx.BindJSON(ctx, &request, "invalid import job payload") {
		return
	}
	slog.Warn("admin import job is not implemented", "scope", request.Scope, "fileName", request.FileName)
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "批量导入任务暂未开放，请使用文章管理中的 Markdown 导入。"})
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
