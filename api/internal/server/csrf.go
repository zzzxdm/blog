package server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	csrfCookieName = "blog_csrf"
	csrfHeaderName = "X-CSRF-Token"
	csrfTokenBytes = 32
)

func csrfProtection(webOrigin string, publicURL string, cookieSecure bool) gin.HandlerFunc {
	allowedOrigins := map[string]bool{}
	addAllowedOrigin(allowedOrigins, webOrigin)
	addAllowedOrigin(allowedOrigins, publicURL)

	return func(ctx *gin.Context) {
		if !strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			ctx.Next()
			return
		}

		token := ensureCSRFCookie(ctx, cookieSecure)

		if !isWriteMethod(ctx.Request.Method) {
			ctx.Next()
			return
		}

		if !csrfOriginAllowed(ctx, allowedOrigins) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf origin rejected"})
			return
		}

		headerToken := strings.TrimSpace(ctx.GetHeader(csrfHeaderName))
		if headerToken == "" || token == "" || !secureTokenEqual(headerToken, token) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf token rejected"})
			return
		}

		ctx.Next()
	}
}

func ensureCSRFCookie(ctx *gin.Context, cookieSecure bool) string {
	if token, err := ctx.Cookie(csrfCookieName); err == nil {
		token = strings.TrimSpace(token)
		if token != "" {
			return token
		}
	}

	token, err := newCSRFToken()
	if err != nil {
		return ""
	}

	setCSRFCookie(ctx, token, cookieSecure)
	return token
}

func setCSRFCookie(ctx *gin.Context, token string, cookieSecure bool) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	// Readable by JS for double-submit CSRF header.
	ctx.SetCookie(csrfCookieName, token, 7*24*60*60, "/", "", cookieSecure, false)
}

func newCSRFToken() (string, error) {
	buf := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func secureTokenEqual(left string, right string) bool {
	if len(left) != len(right) {
		return false
	}

	var diff byte
	for i := 0; i < len(left); i++ {
		diff |= left[i] ^ right[i]
	}
	return diff == 0
}

func csrfOriginAllowed(ctx *gin.Context, allowedOrigins map[string]bool) bool {
	if rawOrigin := strings.TrimSpace(ctx.GetHeader("Origin")); rawOrigin != "" {
		origin := normalizedHeaderOrigin(rawOrigin)
		return origin != "" && originAllowed(ctx, allowedOrigins, origin)
	}

	if rawReferer := strings.TrimSpace(ctx.GetHeader("Referer")); rawReferer != "" {
		referer := normalizedRefererOrigin(rawReferer)
		return referer != "" && originAllowed(ctx, allowedOrigins, referer)
	}

	// Same-origin browser writes usually send Origin/Referer.
	// Allow only when the request host itself is an allowed origin (direct API call tooling).
	requestOriginValue := requestOrigin(ctx)
	return requestOriginValue != "" && allowedOrigins[requestOriginValue]
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
