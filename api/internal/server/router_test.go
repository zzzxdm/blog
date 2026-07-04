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

func contains(value string, part string) bool {
	return strings.Contains(value, part)
}
