package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthSensitiveRateLimitLimitsRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(authSensitiveRateLimit())
	router.POST("/api/auth/register", func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	for i := 0; i < 30; i++ {
		recorder := performRateLimitedRegister(router)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("request %d status = %d, want 201", i+1, recorder.Code)
		}
	}

	recorder := performRateLimitedRegister(router)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("31st request status = %d, want 429 with body %q", recorder.Code, recorder.Body.String())
	}
}

func performRateLimitedRegister(router *gin.Engine) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
