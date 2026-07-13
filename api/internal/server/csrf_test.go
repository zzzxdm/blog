package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCSRFProtectionRejectsWriteWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(csrfProtection("http://localhost:5173", "http://localhost:5173", false))
	router.POST("/api/auth/login", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/api/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})

	getRecorder := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	router.ServeHTTP(getRecorder, getRequest)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", getRecorder.Code)
	}

	csrfCookie := ""
	for _, cookie := range getRecorder.Result().Cookies() {
		if cookie.Name == csrfCookieName {
			csrfCookie = cookie.Value
		}
	}
	if csrfCookie == "" {
		t.Fatal("expected csrf cookie on GET")
	}

	missingToken := httptest.NewRecorder()
	missingRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{}`))
	missingRequest.Header.Set("Content-Type", "application/json")
	missingRequest.Header.Set("Origin", "http://localhost:5173")
	missingRequest.AddCookie(&http.Cookie{Name: csrfCookieName, Value: csrfCookie})
	router.ServeHTTP(missingToken, missingRequest)
	if missingToken.Code != http.StatusForbidden {
		t.Fatalf("missing token status = %d, want 403 body %q", missingToken.Code, missingToken.Body.String())
	}

	okRecorder := httptest.NewRecorder()
	okRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{}`))
	okRequest.Header.Set("Content-Type", "application/json")
	okRequest.Header.Set("Origin", "http://localhost:5173")
	okRequest.Header.Set(csrfHeaderName, csrfCookie)
	okRequest.AddCookie(&http.Cookie{Name: csrfCookieName, Value: csrfCookie})
	router.ServeHTTP(okRecorder, okRequest)
	if okRecorder.Code != http.StatusOK {
		t.Fatalf("valid token status = %d, want 200 body %q", okRecorder.Code, okRecorder.Body.String())
	}
}

func TestCSRFProtectionRejectsBadOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(csrfProtection("http://localhost:5173", "http://localhost:5173", false))
	router.POST("/api/auth/login", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Seed cookie.
	seed := httptest.NewRecorder()
	seedReq := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	// Use a GET route via middleware only path under /api
	router.GET("/api/seed", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	seedReq = httptest.NewRequest(http.MethodGet, "/api/seed", nil)
	router.ServeHTTP(seed, seedReq)
	csrfCookie := ""
	for _, cookie := range seed.Result().Cookies() {
		if cookie.Name == csrfCookieName {
			csrfCookie = cookie.Value
		}
	}
	if csrfCookie == "" {
		t.Fatal("expected csrf cookie")
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://evil.example")
	request.Header.Set(csrfHeaderName, csrfCookie)
	request.AddCookie(&http.Cookie{Name: csrfCookieName, Value: csrfCookie})
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("bad origin status = %d, want 403", recorder.Code)
	}
}
