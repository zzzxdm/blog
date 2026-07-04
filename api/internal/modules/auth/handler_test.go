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
	settings SecuritySettings
}

func (reader staticSettingsReader) SecuritySettings(context.Context) (SecuritySettings, error) {
	return reader.settings, nil
}

func TestLoginUsesConfiguredSessionDays(t *testing.T) {
	store := NewMemoryStore()
	router := gin.New()
	RegisterRoutesWithSettings(router, store, staticSettingsReader{
		settings: SecuritySettings{SessionDays: 14},
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

func TestLoginFailureLockBlocksRepeatedAttempts(t *testing.T) {
	store := NewMemoryStore()
	router := gin.New()
	RegisterRoutesWithSettings(router, store, staticSettingsReader{
		settings: SecuritySettings{
			SessionDays:      7,
			LoginFailureLock: true,
		},
	})

	for index := 0; index < loginFailureLimit-1; index++ {
		recorder := performLogin(router, "linyi@example.com", "wrong-password")
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want 401 with body %q", index+1, recorder.Code, recorder.Body.String())
		}
	}

	locked := performLogin(router, "linyi@example.com", "wrong-password")
	if locked.Code != http.StatusTooManyRequests {
		t.Fatalf("locked attempt status = %d, want 429 with body %q", locked.Code, locked.Body.String())
	}

	valid := performLogin(router, "linyi@example.com", "password")
	if valid.Code != http.StatusTooManyRequests {
		t.Fatalf("valid login while locked status = %d, want 429 with body %q", valid.Code, valid.Body.String())
	}
}

func performLogin(router *gin.Engine, email string, password string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{
		"email":"`+email+`",
		"password":"`+password+`"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func sessionCookie(cookies []*http.Cookie) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == SessionCookieName {
			return cookie
		}
	}

	return nil
}
