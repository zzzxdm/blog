package operations

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"net/http"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

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
		slog.Error("failed to load audit logs", "error", err, "action", ctx.Query("action"), "resourceType", ctx.Query("resourceType"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load audit logs"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
