package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type staticSettingsReader struct {
	settings SessionSettings
}

func (reader staticSettingsReader) SessionSettings(context.Context) (SessionSettings, error) {
	return reader.settings, nil
}

func TestLoginUsesConfiguredSessionDays(t *testing.T) {
	store := NewMemoryStore()
	router := gin.New()
	RegisterRoutesWithSettings(router, store, staticSettingsReader{
		settings: SessionSettings{SessionDays: 14},
	})

	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{
		"email":"linyi@example.com",
		"password":"password"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	cookie := sessionCookie(recorder.Result().Cookies())
	if cookie == nil {
		t.Fatal("expected session cookie")
	}
	if cookie.MaxAge != 14*24*60*60 {
		t.Fatalf("session cookie MaxAge = %d, want 14 days", cookie.MaxAge)
	}

	sessions, err := store.ListSessions("user_linyi", cookie.Value)
	if err != nil {
		t.Fatalf("ListSessions returned error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("session count = %d, want 1", len(sessions))
	}

	expiresAt, err := time.Parse(time.RFC3339, sessions[0].ExpiresAt)
	if err != nil {
		t.Fatalf("parse session expiry: %v", err)
	}
	remaining := time.Until(expiresAt)
	if remaining < 13*24*time.Hour || remaining > 15*24*time.Hour {
		t.Fatalf("session expires in %s, want about 14 days", remaining)
	}
}

func sessionCookie(cookies []*http.Cookie) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == SessionCookieName {
			return cookie
		}
	}

	return nil
}
