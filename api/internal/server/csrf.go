package server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func csrfProtection(webOrigin string, publicURL string) gin.HandlerFunc {
	allowedOrigins := map[string]bool{}
	addAllowedOrigin(allowedOrigins, webOrigin)
	addAllowedOrigin(allowedOrigins, publicURL)

	return func(ctx *gin.Context) {
		if !strings.HasPrefix(ctx.Request.URL.Path, "/api/") || !isWriteMethod(ctx.Request.Method) {
			ctx.Next()
			return
		}

		if rawOrigin := strings.TrimSpace(ctx.GetHeader("Origin")); rawOrigin != "" {
			origin := normalizedHeaderOrigin(rawOrigin)
			if origin == "" || !originAllowed(ctx, allowedOrigins, origin) {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf origin rejected"})
				return
			}

			ctx.Next()
			return
		}

		if rawReferer := strings.TrimSpace(ctx.GetHeader("Referer")); rawReferer != "" {
			referer := normalizedRefererOrigin(rawReferer)
			if referer == "" || !originAllowed(ctx, allowedOrigins, referer) {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf referer rejected"})
				return
			}
		}

		ctx.Next()
	}
}

func addAllowedOrigin(allowedOrigins map[string]bool, value string) {
	if origin := normalizeOrigin(value); origin != "" {
		allowedOrigins[origin] = true
	}
}

func originAllowed(ctx *gin.Context, allowedOrigins map[string]bool, origin string) bool {
	if allowedOrigins[origin] {
		return true
	}

	return origin == requestOrigin(ctx)
}

func normalizedHeaderOrigin(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "null") {
		return "null"
	}

	return normalizeOrigin(value)
}

func normalizedRefererOrigin(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}

	return normalizeOrigin(value)
}

func normalizeOrigin(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return ""
	}

	return scheme + "://" + strings.ToLower(parsed.Host)
}

func requestOrigin(ctx *gin.Context) string {
	scheme := strings.ToLower(strings.TrimSpace(ctx.GetHeader("X-Forwarded-Proto")))
	if scheme == "" {
		scheme = "http"
		if ctx.Request.TLS != nil {
			scheme = "https"
		}
	}

	host := strings.ToLower(strings.TrimSpace(ctx.Request.Host))
	if host == "" {
		return ""
	}

	return scheme + "://" + host
}
