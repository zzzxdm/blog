package server

import (
	"net/http"
	"net/url"
	"strings"

	"blog/api/internal/modules/operations"

	"github.com/gin-gonic/gin"
)

func navigationRedirects(repo operations.Repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if repo == nil || !redirectableRequest(ctx.Request) {
			ctx.Next()
			return
		}

		navigation, err := repo.GetNavigation(ctx.Request.Context())
		if err != nil {
			ctx.Next()
			return
		}

		requestPath := normalizedRedirectPath(ctx.Request.URL.Path)
		if requestPath == "" {
			ctx.Next()
			return
		}

		for _, rule := range navigation.Redirects {
			if normalizedRedirectPath(rule.From) != requestPath {
				continue
			}

			target := normalizedRedirectTarget(rule.To, ctx.Request.URL.RawQuery)
			if target == "" || sameRedirectTarget(target, ctx.Request.URL) {
				ctx.Next()
				return
			}

			ctx.Redirect(normalizedRedirectCode(rule.Code), target)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func redirectableRequest(request *http.Request) bool {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		return false
	}

	path := request.URL.Path
	return path != "/api" &&
		path != "/uploads" &&
		!strings.HasPrefix(path, "/api/") &&
		!strings.HasPrefix(path, "/uploads/")
}

func normalizedRedirectPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "/") {
		return ""
	}

	parsed, err := url.Parse(value)
	if err == nil && parsed.Path != "" {
		value = parsed.Path
	}

	if value != "/" {
		value = strings.TrimRight(value, "/")
	}

	if value == "" {
		return "/"
	}

	return value
}

func normalizedRedirectTarget(value string, rawQuery string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		value = "/" + value
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	if parsed.Path == "" && parsed.Host == "" {
		return ""
	}
	if parsed.RawQuery == "" {
		parsed.RawQuery = rawQuery
	}

	return parsed.String()
}

func sameRedirectTarget(target string, requestURL *url.URL) bool {
	parsed, err := url.Parse(target)
	if err != nil || parsed.IsAbs() {
		return false
	}

	return normalizedRedirectPath(parsed.Path) == normalizedRedirectPath(requestURL.Path) &&
		parsed.RawQuery == requestURL.RawQuery
}

func normalizedRedirectCode(code int) int {
	switch code {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return code
	default:
		return http.StatusMovedPermanently
	}
}
