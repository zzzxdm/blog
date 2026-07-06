package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

type fakeUserEmailSender struct {
	token           string
	initialPassword string
	err             error
}

func (sender *fakeUserEmailSender) SendEmailVerification(context.Context, auth.User, string) error {
	return nil
}

func (sender *fakeUserEmailSender) SendPasswordSetup(_ context.Context, _ auth.User, token string) error {
	sender.token = token
	return sender.err
}

func (sender *fakeUserEmailSender) SendInvitation(_ context.Context, _ auth.User, initialPassword string, token string) error {
	sender.initialPassword = initialPassword
	sender.token = token
	return sender.err
}

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

func TestDeleteSyncsAuthStoreAndManagedUser(t *testing.T) {
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

	request := httptest.NewRequest(http.MethodDelete, "/admin/users/user_linyi", nil)
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
	if updated.Status != "deleted" {
		t.Fatalf("response status = %q, want deleted", updated.Status)
	}

	managed, err := repo.Get(request.Context(), "user_linyi")
	if err != nil {
		t.Fatalf("repo.Get returned error: %v", err)
	}
	if managed.Status != "deleted" {
		t.Fatalf("managed status = %q, want deleted", managed.Status)
	}

	_, _, err = authStore.Authenticate("linyi@example.com", "password")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("Authenticate deleted user error = %v, want ErrInvalidCredentials", err)
	}

	_, err = authStore.UserBySession(userToken)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("UserBySession deleted user error = %v, want ErrInvalidSession", err)
	}
}

func TestDeleteRejectsCurrentAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authStore := auth.NewMemoryStore()
	repo := NewMemoryRepository()
	_, token, err := authStore.Authenticate("admin@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate admin returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(authStore))
	RegisterRoutes(router, repo, authStore)

	request := httptest.NewRequest(http.MethodDelete, "/admin/users/user_admin", nil)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	user, err := authStore.UserBySession(token)
	if err != nil {
		t.Fatalf("admin session should remain valid: %v", err)
	}
	if user.Status != "active" {
		t.Fatalf("admin status = %q, want active", user.Status)
	}
}

func TestListUsersPaginatesAndFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authStore := auth.NewMemoryStore()
	repo := NewMemoryRepository()
	_, token, err := authStore.Authenticate("admin@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate admin returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(authStore))
	RegisterRoutes(router, repo, authStore)

	request := httptest.NewRequest(http.MethodGet, "/admin/users?page=1&pageSize=1&status=active&role=user", nil)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	var response UserListResult
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Page != 1 {
		t.Fatalf("page = %d, want 1", response.Page)
	}
	if response.PageSize != 1 {
		t.Fatalf("pageSize = %d, want 1", response.PageSize)
	}
	if response.Total != 2 {
		t.Fatalf("total = %d, want 2", response.Total)
	}
	if len(response.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(response.Items))
	}
	if response.Items[0].Role != "user" || response.Items[0].Status != "active" {
		t.Fatalf("item role/status = %q/%q, want user/active", response.Items[0].Role, response.Items[0].Status)
	}
	if response.Stats.Total != 5 {
		t.Fatalf("stats.total = %d, want global total 5", response.Stats.Total)
	}
}

func TestInviteSendsPasswordSetupEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authStore := auth.NewMemoryStore()
	repo := NewMemoryRepository()
	emailSender := &fakeUserEmailSender{}
	_, token, err := authStore.Authenticate("admin@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate admin returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(authStore))
	RegisterRoutesWithEmailSender(router, repo, authStore, emailSender)

	request := httptest.NewRequest(http.MethodPost, "/admin/users/invitations", bytes.NewBufferString(`{
		"email":"invite-mail@example.com",
		"displayName":"Invite Mail",
		"role":"author"
	}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	var response InvitationResult
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Delivery != "email" {
		t.Fatalf("delivery = %q, want email", response.Delivery)
	}
	if response.ResetToken != "" {
		t.Fatal("did not expect reset token when invitation email is sent")
	}
	if response.InitialPassword != "" {
		t.Fatal("did not expect initial password when invitation email is sent")
	}
	if emailSender.token == "" {
		t.Fatal("expected password reset token to be emailed")
	}
	if emailSender.initialPassword == "" {
		t.Fatal("expected initial password to be emailed")
	}
}

func TestInviteKeepsTokenWhenPasswordSetupEmailFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authStore := auth.NewMemoryStore()
	repo := NewMemoryRepository()
	_, token, err := authStore.Authenticate("admin@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate admin returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(authStore))
	RegisterRoutesWithEmailSender(router, repo, authStore, &fakeUserEmailSender{err: errors.New("smtp unavailable")})

	request := httptest.NewRequest(http.MethodPost, "/admin/users/invitations", bytes.NewBufferString(`{
		"email":"invite-failed@example.com",
		"displayName":"Invite Failed",
		"role":"author"
	}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	var response InvitationResult
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Delivery != "email-failed" {
		t.Fatalf("delivery = %q, want email-failed", response.Delivery)
	}
	if response.ResetToken == "" {
		t.Fatal("expected reset token when invitation email fails")
	}
	if response.InitialPassword == "" {
		t.Fatal("expected initial password when invitation email fails")
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
