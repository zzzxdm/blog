package users

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

func TestUpdateStatusSyncsAuthStoreAndManagedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authStore := auth.NewMemoryStore()
	repo := NewMemoryRepository()
	_, token, err := authStore.Authenticate("admin@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate admin returned error: %v", err)
	}
	_, userToken, err := authStore.Authenticate("linyi@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate user returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(authStore))
	RegisterRoutes(router, repo, authStore)

	request := httptest.NewRequest(http.MethodPut, "/admin/users/user_linyi/status", bytes.NewBufferString(`{"status":"banned"}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	var updated ManagedUser
	if err := json.NewDecoder(recorder.Body).Decode(&updated); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if updated.Status != "banned" {
		t.Fatalf("response status = %q, want banned", updated.Status)
	}
	if updated.ModerationNote == "" {
		t.Fatal("expected moderation note in response")
	}

	managed, err := repo.Get(request.Context(), "user_linyi")
	if err != nil {
		t.Fatalf("repo.Get returned error: %v", err)
	}
	if managed.Status != "banned" {
		t.Fatalf("managed status = %q, want banned", managed.Status)
	}

	_, _, err = authStore.Authenticate("linyi@example.com", "password")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("Authenticate banned user error = %v, want ErrInvalidCredentials", err)
	}

	_, err = authStore.UserBySession(userToken)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("UserBySession banned user error = %v, want ErrInvalidSession", err)
	}
}

func TestUpdateAvatarSyncsAuthStore(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authStore := auth.NewMemoryStore()
	repo := NewMemoryRepository()
	_, token, err := authStore.Authenticate("linyi@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate user returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(authStore))
	RegisterRoutes(router, repo, authStore)

	request := httptest.NewRequest(http.MethodPost, "/me/avatar", bytes.NewBufferString(`{"avatarText":"新"}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	user, err := authStore.UserBySession(token)
	if err != nil {
		t.Fatalf("UserBySession returned error: %v", err)
	}
	if user.AvatarText != "新" {
		t.Fatalf("avatar text = %q, want 新", user.AvatarText)
	}
}
