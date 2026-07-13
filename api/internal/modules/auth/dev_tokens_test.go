package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type forgotPasswordStore struct {
	token string
}

func (s forgotPasswordStore) Authenticate(string, string) (User, string, error) {
	return User{}, "", ErrInvalidCredentials
}
func (s forgotPasswordStore) Register(RegisterRequest) (User, string, error) {
	return User{}, "", ErrInvalidCredentials
}
func (s forgotPasswordStore) InviteUser(InviteUserRequest) (User, InvitationSecrets, error) {
	return User{}, InvitationSecrets{}, ErrInvalidCredentials
}
func (s forgotPasswordStore) UpdateRole(string, string) (User, error) {
	return User{}, ErrInvalidSession
}
func (s forgotPasswordStore) UpdateStatus(string, string) (User, error) {
	return User{}, ErrInvalidSession
}
func (s forgotPasswordStore) UpdateProfile(string, string, string) (User, error) {
	return User{}, ErrInvalidSession
}
func (s forgotPasswordStore) UserBySession(string) (User, error) {
	return User{}, ErrInvalidSession
}
func (s forgotPasswordStore) SetSessionExpiry(string, time.Time) error { return nil }
func (s forgotPasswordStore) ChangePassword(string, string, string) error {
	return ErrInvalidCredentials
}
func (s forgotPasswordStore) RequestEmailVerification(string) (string, error) {
	return "verify-token", nil
}
func (s forgotPasswordStore) VerifyEmail(string) (User, error) { return User{}, ErrInvalidToken }
func (s forgotPasswordStore) RequestPasswordReset(email string) (User, string, error) {
	if strings.TrimSpace(email) == "missing@example.com" {
		return User{}, "", nil
	}
	return User{ID: "1", Email: email, DisplayName: "User", Role: "user", Status: "active"}, s.token, nil
}
func (s forgotPasswordStore) ResetPassword(string, string) error { return ErrInvalidToken }
func (s forgotPasswordStore) ListSessions(string, string) ([]SessionInfo, error) {
	return nil, nil
}
func (s forgotPasswordStore) DeleteUserSession(string, string) error { return nil }
func (s forgotPasswordStore) ExportUserData(string, string) (ExportData, error) {
	return ExportData{}, nil
}
func (s forgotPasswordStore) DeleteUser(string) error  { return nil }
func (s forgotPasswordStore) DeleteSession(string)     {}

func TestProductionHidesPasswordResetToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ConfigureDevAuthTokens(false)
	t.Cleanup(func() { ConfigureDevAuthTokens(true) })

	handler := &Handler{store: forgotPasswordStore{token: "reset-token-secret"}}
	router := gin.New()
	router.POST("/auth/forgot-password", handler.ForgotPassword)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(`{"email":"user@example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body %q", recorder.Code, recorder.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := payload["resetToken"]; ok {
		t.Fatalf("resetToken must be omitted: %#v", payload)
	}
	if payload["ok"] != true {
		t.Fatalf("ok missing: %#v", payload)
	}
}

func TestForgotPasswordHidesMissingEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ConfigureDevAuthTokens(false)
	t.Cleanup(func() { ConfigureDevAuthTokens(true) })

	handler := &Handler{store: forgotPasswordStore{token: "reset-token-secret"}}
	router := gin.New()
	router.POST("/auth/forgot-password", handler.ForgotPassword)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(`{"email":"missing@example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body %q", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if strings.Contains(body, "not registered") || strings.Contains(body, "reset-token-secret") {
		t.Fatalf("leaked account/token info: %s", body)
	}
}

func TestDevelopmentCanExposePasswordResetToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ConfigureDevAuthTokens(true)
	t.Cleanup(func() { ConfigureDevAuthTokens(true) })

	handler := &Handler{store: forgotPasswordStore{token: "reset-token-secret"}}
	router := gin.New()
	router.POST("/auth/forgot-password", handler.ForgotPassword)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(`{"email":"user@example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body %q", recorder.Code, recorder.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["resetToken"] != "reset-token-secret" {
		t.Fatalf("dev mode should expose token: %#v", payload)
	}
}
