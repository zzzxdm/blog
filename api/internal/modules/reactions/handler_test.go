package reactions

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/posts"

	"github.com/gin-gonic/gin"
)

func TestReactionEndpointsReturnNotFoundForMissingPost(t *testing.T) {
	store := auth.NewMemoryStore()
	_, token, err := store.Authenticate("linyi@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		auth   bool
	}{
		{name: "get reaction", method: http.MethodGet, path: "/posts/missing-post/reaction"},
		{name: "set reaction", method: http.MethodPut, path: "/posts/missing-post/reaction", body: `{"type":"like"}`, auth: true},
		{name: "set bookmark", method: http.MethodPut, path: "/posts/missing-post/bookmark", body: `{"bookmarked":true}`, auth: true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(auth.Middleware(store))
			RegisterRoutes(router, NewMemoryRepository(), posts.NewMemoryRepository())

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			if tt.auth {
				req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("expected status 404, got %d with body %q", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestMutedUserCannotReactOrBookmark(t *testing.T) {
	store := auth.NewMemoryStore()
	_, token, err := store.Authenticate("linyi@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if _, err := store.UpdateStatus("user_linyi", "muted"); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "set reaction", method: http.MethodPut, path: "/posts/blog-system-design/reaction", body: `{"type":"like"}`},
		{name: "clear reaction", method: http.MethodDelete, path: "/posts/blog-system-design/reaction"},
		{name: "set bookmark", method: http.MethodPut, path: "/posts/blog-system-design/bookmark", body: `{"bookmarked":true}`},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(auth.Middleware(store))
			RegisterRoutes(router, NewMemoryRepository(), posts.NewMemoryRepository())

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected status 403, got %d with body %q", rec.Code, rec.Body.String())
			}
		})
	}
}
