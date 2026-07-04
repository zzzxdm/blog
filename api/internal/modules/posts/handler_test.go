package posts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestListPosts(t *testing.T) {
	router := gin.New()
	RegisterPublicRoutes(router, NewMemoryRepository())

	req := httptest.NewRequest(http.MethodGet, "/posts?page=1&pageSize=2", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"total":3`) {
		t.Fatalf("expected total count in body, got %q", body)
	}
}

func TestGetPostBySlug(t *testing.T) {
	router := gin.New()
	RegisterPublicRoutes(router, NewMemoryRepository())

	req := httptest.NewRequest(http.MethodGet, "/posts/blog-system-design", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"slug":"blog-system-design"`) {
		t.Fatalf("expected post slug in body, got %q", rec.Body.String())
	}
}

func TestGetPostBySlugRecordsView(t *testing.T) {
	repo := NewMemoryRepository()
	before, err := repo.GetBySlug(context.Background(), "blog-system-design")
	if err != nil {
		t.Fatalf("GetBySlug before returned error: %v", err)
	}

	router := gin.New()
	RegisterPublicRoutes(router, repo)

	req := httptest.NewRequest(http.MethodGet, "/posts/blog-system-design", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var response Post
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ViewCount != before.ViewCount+1 {
		t.Fatalf("response viewCount = %d, want %d", response.ViewCount, before.ViewCount+1)
	}

	after, err := repo.GetBySlug(context.Background(), "blog-system-design")
	if err != nil {
		t.Fatalf("GetBySlug after returned error: %v", err)
	}
	if after.ViewCount != before.ViewCount+1 {
		t.Fatalf("stored viewCount = %d, want %d", after.ViewCount, before.ViewCount+1)
	}
}

func TestSiteStats(t *testing.T) {
	router := gin.New()
	RegisterPublicRoutes(router, NewMemoryRepository())

	req := httptest.NewRequest(http.MethodGet, "/site-stats", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var response SiteStats
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.PostCount != 3 || response.ViewCount == 0 || response.WordCount == 0 {
		t.Fatalf("unexpected site stats: %+v", response)
	}
}

func TestSearchRequiresKeyword(t *testing.T) {
	router := gin.New()
	RegisterPublicRoutes(router, NewMemoryRepository())

	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"total":0`) {
		t.Fatalf("expected empty search result, got %q", rec.Body.String())
	}
}
