package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blog/api/internal/config"
)

func TestHealth(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if body == "" || !contains(body, `"status":"ok"`) {
		t.Fatalf("expected ok health body, got %q", body)
	}
}

func TestFeedSitemapAndRobots(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
		PublicURL: "https://blog.example.com",
	})

	rssReq := httptest.NewRequest(http.MethodGet, "/rss.xml", nil)
	rssRec := httptest.NewRecorder()
	router.ServeHTTP(rssRec, rssReq)
	if rssRec.Code != http.StatusOK || !strings.Contains(rssRec.Body.String(), "<rss") || !strings.Contains(rssRec.Body.String(), "https://blog.example.com/posts/blog-system-design") {
		t.Fatalf("expected rss feed, got status=%d body=%q", rssRec.Code, rssRec.Body.String())
	}

	sitemapReq := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	sitemapRec := httptest.NewRecorder()
	router.ServeHTTP(sitemapRec, sitemapReq)
	if sitemapRec.Code != http.StatusOK || !strings.Contains(sitemapRec.Body.String(), "<urlset") || !strings.Contains(sitemapRec.Body.String(), "https://blog.example.com/archive") {
		t.Fatalf("expected sitemap, got status=%d body=%q", sitemapRec.Code, sitemapRec.Body.String())
	}

	robotsReq := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	robotsRec := httptest.NewRecorder()
	router.ServeHTTP(robotsRec, robotsReq)
	if robotsRec.Code != http.StatusOK || !strings.Contains(robotsRec.Body.String(), "Sitemap: https://blog.example.com/sitemap.xml") {
		t.Fatalf("expected robots txt, got status=%d body=%q", robotsRec.Code, robotsRec.Body.String())
	}
}

func TestArticleSEOHTML(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
		PublicURL: "https://blog.example.com",
	})

	req := httptest.NewRequest(http.MethodGet, "/posts/blog-system-design", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%q", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	expectedParts := []string{
		"<title>如何设计一个内容长期增长的博客系统 - 云间笔记</title>",
		`<link rel="canonical" href="https://blog.example.com/posts/blog-system-design">`,
		`<meta property="og:title" content="如何设计一个内容长期增长的博客系统">`,
		`<script type="application/ld+json">`,
		`"@type":"BlogPosting"`,
		`"url":"https://blog.example.com/posts/blog-system-design"`,
		`<link rel="stylesheet" href="/assets/index.css">`,
		`<script type="module" src="/assets/index.js"></script>`,
	}
	for _, part := range expectedParts {
		if !strings.Contains(body, part) {
			t.Fatalf("expected article html to contain %q, got %q", part, body)
		}
	}
}

func TestCSRFOriginProtection(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
		PublicURL: "https://blog.example.com",
	})

	evilReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	evilReq.Header.Set("Origin", "https://evil.example.com")
	evilRec := httptest.NewRecorder()
	router.ServeHTTP(evilRec, evilReq)
	if evilRec.Code != http.StatusForbidden {
		t.Fatalf("expected evil origin status 403, got %d body=%q", evilRec.Code, evilRec.Body.String())
	}

	webReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	webReq.Header.Set("Origin", "http://localhost:5173")
	webRec := httptest.NewRecorder()
	router.ServeHTTP(webRec, webReq)
	if webRec.Code != http.StatusOK {
		t.Fatalf("expected configured web origin accepted, got %d body=%q", webRec.Code, webRec.Body.String())
	}

	publicReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	publicReq.Header.Set("Origin", "https://blog.example.com")
	publicRec := httptest.NewRecorder()
	router.ServeHTTP(publicRec, publicReq)
	if publicRec.Code != http.StatusOK {
		t.Fatalf("expected public origin accepted, got %d body=%q", publicRec.Code, publicRec.Body.String())
	}

	refererReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	refererReq.Header.Set("Referer", "https://evil.example.com/logout")
	refererRec := httptest.NewRecorder()
	router.ServeHTTP(refererRec, refererReq)
	if refererRec.Code != http.StatusForbidden {
		t.Fatalf("expected evil referer status 403, got %d body=%q", refererRec.Code, refererRec.Body.String())
	}
}

func contains(value string, part string) bool {
	return strings.Contains(value, part)
}
