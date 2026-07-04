package posts

import (
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
