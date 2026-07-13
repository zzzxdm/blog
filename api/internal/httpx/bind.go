package httpx

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func BindJSON(ctx *gin.Context, target any, clientMessage string, attrs ...any) bool {
	if err := ctx.ShouldBindJSON(target); err != nil {
		logRequestError(ctx, slog.LevelWarn, "invalid request payload", err, attrs...)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": clientMessage})
		return false
	}

	return true
}

func BindOptionalJSON(ctx *gin.Context, target any, clientMessage string, attrs ...any) bool {
	if err := ctx.ShouldBindJSON(target); err != nil {
		if errors.Is(err, io.EOF) {
			return true
		}

		logRequestError(ctx, slog.LevelWarn, "invalid optional request payload", err, attrs...)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": clientMessage})
		return false
	}

	return true
}

func LogRequestError(ctx *gin.Context, message string, err error, attrs ...any) {
	logRequestError(ctx, slog.LevelError, message, err, attrs...)
}

func logRequestError(ctx *gin.Context, level slog.Level, message string, err error, attrs ...any) {
	args := []any{
		"error", err,
		"method", ctx.Request.Method,
		"route", ctx.FullPath(),
		"path", ctx.Request.URL.Path,
	}
	args = append(args, attrs...)
	slog.Log(ctx.Request.Context(), level, message, args...)
}
