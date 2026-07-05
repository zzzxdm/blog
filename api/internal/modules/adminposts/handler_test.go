package adminposts

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

func TestPreviewTokenFlow(t *testing.T) {
	authStore := auth.NewMemoryStore()
	_, token, err := authStore.Authenticate("admin@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(authStore))
	RegisterRoutes(router, NewMemoryRepository(), nil)

	createReq := httptest.NewRequest(http.MethodPost, "/admin/posts/admin_post_001/preview", nil)
	createReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("expected preview creation status 200, got %d body=%q", createRec.Code, createRec.Body.String())
	}

	var preview PreviewResult
	if err := json.Unmarshal(createRec.Body.Bytes(), &preview); err != nil {
		t.Fatalf("decode preview result: %v", err)
	}
	if preview.Token == "" || !strings.HasPrefix(preview.PreviewURL, "/preview/") {
		t.Fatalf("unexpected preview result: %+v", preview)
	}

	previewReq := httptest.NewRequest(http.MethodGet, "/preview/"+preview.Token, nil)
	previewRec := httptest.NewRecorder()
	router.ServeHTTP(previewRec, previewReq)

	if previewRec.Code != http.StatusOK {
		t.Fatalf("expected preview status 200, got %d body=%q", previewRec.Code, previewRec.Body.String())
	}
	if !strings.Contains(previewRec.Body.String(), `"id":"admin_post_001"`) {
		t.Fatalf("expected draft post preview, got %q", previewRec.Body.String())
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/preview/not-a-token", nil)
	invalidRec := httptest.NewRecorder()
	router.ServeHTTP(invalidRec, invalidReq)
	if invalidRec.Code != http.StatusNotFound {
		t.Fatalf("expected invalid preview status 404, got %d body=%q", invalidRec.Code, invalidRec.Body.String())
	}
}
