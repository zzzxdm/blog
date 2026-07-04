package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/operations"

	"github.com/gin-gonic/gin"
)

func auditAdminWrites(repo operations.Repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if repo == nil || !isWriteMethod(ctx.Request.Method) || !strings.HasPrefix(ctx.Request.URL.Path, "/api/admin/") {
			ctx.Next()
			return
		}

		startedAt := time.Now()
		ctx.Next()

		user, _ := auth.CurrentUser(ctx)
		resourceType, resourceID, resourceTitle := auditResource(ctx.Request.URL.Path)
		statusCode := ctx.Writer.Status()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		_ = repo.RecordAuditLog(ctx.Request.Context(), operations.AuditLog{
			ActorID:       user.ID,
			ActorName:     user.DisplayName,
			Action:        auditAction(ctx.Request.Method, ctx.Request.URL.Path),
			ResourceType:  resourceType,
			ResourceID:    resourceID,
			ResourceTitle: resourceTitle,
			Status:        auditStatus(statusCode),
			IP:            ctx.ClientIP(),
			UserAgent:     ctx.Request.UserAgent(),
			Detail:        fmt.Sprintf("%s %s -> %d (%dms)", ctx.Request.Method, auditRoute(ctx), statusCode, time.Since(startedAt).Milliseconds()),
			CreatedAt:     time.Now(),
		})
	}
}

func auditAction(method string, path string) string {
	switch {
	case strings.Contains(path, "/admin/settings/test-mail"):
		return "settings.test_mail"
	case strings.Contains(path, "/admin/settings"):
		return "settings.update"
	case strings.Contains(path, "/admin/backups"):
		return "backup.create"
	case strings.Contains(path, "/admin/navigation"):
		return "navigation.update"
	case strings.Contains(path, "/admin/media") && method == http.MethodPost:
		return "media.create"
	case strings.Contains(path, "/admin/media") && method == http.MethodDelete:
		return "media.delete"
	case strings.Contains(path, "/admin/media"):
		return "media.update"
	case strings.Contains(path, "/admin/posts") && strings.Contains(path, "/publish"):
		return "post.publish"
	case strings.Contains(path, "/admin/posts") && strings.Contains(path, "/restore"):
		return "post.restore"
	case strings.Contains(path, "/admin/posts") && method == http.MethodPost:
		return "post.create"
	case strings.Contains(path, "/admin/posts"):
		return "post.update"
	case strings.Contains(path, "/admin/submissions") && strings.Contains(path, "/review"):
		return "submission.review"
	case strings.Contains(path, "/admin/comments"):
		return "comment.moderate"
	case strings.Contains(path, "/admin/users"):
		return "user.update"
	case strings.Contains(path, "/admin/messages"):
		return "message.send"
	case strings.Contains(path, "/admin/categories"), strings.Contains(path, "/admin/tags"):
		return "taxonomy.update"
	default:
		return "admin.write"
	}
}

func auditResource(path string) (string, string, string) {
	trimmed := strings.Trim(strings.TrimPrefix(path, "/api/admin/"), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "admin", "", "后台"
	}

	resourceType := strings.TrimSuffix(parts[0], "s")
	switch parts[0] {
	case "backups":
		return "backup", "", "数据备份"
	case "settings":
		return "settings", "", "系统设置"
	case "navigation":
		return "navigation", "", "导航设置"
	case "categories", "tags":
		resourceType = "taxonomy"
	case "submissions":
		resourceType = "submission"
	}

	resourceID := ""
	if len(parts) > 1 {
		resourceID = parts[1]
	}

	return resourceType, resourceID, strings.Join(parts, "/")
}

func auditStatus(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "error"
	case statusCode >= 400:
		return "blocked"
	default:
		return "success"
	}
}

func auditRoute(ctx *gin.Context) string {
	if route := ctx.FullPath(); route != "" {
		return route
	}

	return ctx.Request.URL.Path
}
