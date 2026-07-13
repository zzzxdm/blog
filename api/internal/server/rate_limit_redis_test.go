package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestAuthSensitiveRateLimitWithRedisLimitsRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	router := gin.New()
	router.Use(authSensitiveRateLimitWithRedis(client))
	router.POST("/api/auth/register", func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	for i := 0; i < 5; i++ {
		recorder := performRateLimitedRegister(router)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("request %d status = %d, want 201 body=%s", i+1, recorder.Code, recorder.Body.String())
		}
	}

	recorder := performRateLimitedRegister(router)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("sixth request status = %d, want 429 body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestWriteRateLimitWithRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	router := gin.New()
	router.Use(rateLimitWithRedis(client, 2, time.Minute))
	router.POST("/api/demo", func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/demo", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want 204", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/demo", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("third request status = %d, want 429", rec.Code)
	}
}
