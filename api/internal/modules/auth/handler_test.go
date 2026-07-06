package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type fakeEmailSender struct {
	token              string
	passwordResetToken string
	passwordResetUser  User
	err                error
}

func (sender *fakeEmailSender) SendEmailVerification(_ context.Context, _ User, token string) error {
	sender.token = token
	return sender.err
}

func (sender *fakeEmailSender) SendPasswordSetup(_ context.Context, user User, token string) error {
	sender.passwordResetUser = user
	sender.passwordResetToken = token
	return sender.err
}

func (sender *fakeEmailSender) SendInvitation(context.Context, User, string, string) error {
	return nil
}

type fakeTurnstileVerifier struct {
	secret   string
	token    string
	remoteIP string
	ok       bool
	err      error
}

func (verifier *fakeTurnstileVerifier) Verify(_ context.Context, secret string, token string, remoteIP string) (bool, error) {
	verifier.secret = secret
	verifier.token = token
	verifier.remoteIP = remoteIP
	return verifier.ok, verifier.err
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

func TestLoginReportsDeletedAccount(t *testing.T) {
	store := NewMemoryStore()
	if _, err := store.UpdateStatus("user_linyi", "deleted"); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	router := gin.New()
	RegisterRoutes(router, store)

	recorder := performLogin(router, "linyi@example.com", "password")

	if recorder.Code != http.StatusGone {
		t.Fatalf("deleted login status = %d, want 410 with body %q", recorder.Code, recorder.Body.String())
	}
}

func TestRegisterReportsDeletedAccountEmail(t *testing.T) {
	store := NewMemoryStore()
	if _, err := store.UpdateStatus("user_linyi", "deleted"); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	router := gin.New()
	RegisterRoutes(router, store)

	request := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{
		"email":"linyi@example.com",
		"password":"password",
		"displayName":"林一新号"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusGone {
		t.Fatalf("deleted account email register status = %d, want 410 with body %q", recorder.Code, recorder.Body.String())
	}
}

func TestLoginRequiresTurnstileTokenWhenEnabled(t *testing.T) {
	store := NewMemoryStore()
	router := gin.New()
	RegisterRoutesWithDependencies(router, store, staticSettingsReader{
		settings: SecuritySettings{
			SessionDays:        7,
			TurnstileEnabled:   true,
			TurnstileSecretKey: "secret-key",
			TurnstileLogin:     true,
		},
	}, nil, &fakeTurnstileVerifier{ok: true})

	recorder := performLogin(router, "linyi@example.com", "password")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected login status 400, got %d with body %q", recorder.Code, recorder.Body.String())
	}
}

func TestLoginVerifiesTurnstileToken(t *testing.T) {
	store := NewMemoryStore()
	verifier := &fakeTurnstileVerifier{ok: true}
	router := gin.New()
	RegisterRoutesWithDependencies(router, store, staticSettingsReader{
		settings: SecuritySettings{
			SessionDays:        7,
			TurnstileEnabled:   true,
			TurnstileSecretKey: "secret-key",
			TurnstileLogin:     true,
		},
	}, nil, verifier)

	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{
		"email":"linyi@example.com",
		"password":"password",
		"turnstileToken":"login-token"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d with body %q", recorder.Code, recorder.Body.String())
	}
	if verifier.secret != "secret-key" || verifier.token != "login-token" {
		t.Fatalf("turnstile verifier got secret/token %q/%q", verifier.secret, verifier.token)
	}
}

func TestRegisterReturnsEmailVerificationToken(t *testing.T) {
	store := NewMemoryStore()
	router := gin.New()
	RegisterRoutes(router, store)

	request := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{
		"email":"verify-new@example.com",
		"password":"password",
		"displayName":"Verify New"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	var registered struct {
		User              User   `json:"user"`
		VerificationToken string `json:"verificationToken"`
		Delivery          string `json:"delivery"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &registered); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if registered.User.EmailVerified {
		t.Fatal("expected registered user to start unverified")
	}
	if registered.VerificationToken == "" {
		t.Fatal("expected verification token")
	}
	if registered.Delivery != "dev-response" {
		t.Fatalf("delivery = %q, want dev-response", registered.Delivery)
	}

	verifyRequest := httptest.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewBufferString(`{"token":"`+registered.VerificationToken+`"}`))
	verifyRequest.Header.Set("Content-Type", "application/json")
	verifyRecorder := httptest.NewRecorder()

	router.ServeHTTP(verifyRecorder, verifyRequest)

	if verifyRecorder.Code != http.StatusOK {
		t.Fatalf("expected verify status 200, got %d with body %q", verifyRecorder.Code, verifyRecorder.Body.String())
	}

	var verified struct {
		User User `json:"user"`
	}
	if err := json.Unmarshal(verifyRecorder.Body.Bytes(), &verified); err != nil {
		t.Fatalf("decode verify response: %v", err)
	}
	if !verified.User.EmailVerified {
		t.Fatal("expected verified user")
	}
}

func TestRegisterSendsEmailVerificationWithoutExposingToken(t *testing.T) {
	store := NewMemoryStore()
	emailSender := &fakeEmailSender{}
	router := gin.New()
	RegisterRoutesWithSettingsAndEmailSender(router, store, nil, emailSender)

	request := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{
		"email":"mail-new@example.com",
		"password":"password",
		"displayName":"Mail New"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	var registered struct {
		User              User   `json:"user"`
		VerificationToken string `json:"verificationToken"`
		Delivery          string `json:"delivery"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &registered); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if registered.VerificationToken != "" {
		t.Fatal("did not expect verification token in response when email sender is configured")
	}
	if emailSender.token == "" {
		t.Fatal("expected verification token to be sent by email sender")
	}
	if registered.Delivery != "email" {
		t.Fatalf("delivery = %q, want email", registered.Delivery)
	}

	verifiedUser, err := store.VerifyEmail(emailSender.token)
	if err != nil {
		t.Fatalf("VerifyEmail returned error: %v", err)
	}
	if !verifiedUser.EmailVerified {
		t.Fatal("expected emailed token to verify user")
	}
}

func TestRegisterSucceedsWhenEmailVerificationDeliveryFails(t *testing.T) {
	store := NewMemoryStore()
	router := gin.New()
	RegisterRoutesWithSettingsAndEmailSender(router, store, nil, &fakeEmailSender{err: errors.New("smtp unavailable")})

	request := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{
		"email":"mail-fail@example.com",
		"password":"password",
		"displayName":"Mail Fail"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	var registered struct {
		User     User   `json:"user"`
		Delivery string `json:"delivery"`
		Warning  string `json:"warning"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &registered); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if registered.User.Email != "mail-fail@example.com" {
		t.Fatalf("registered email = %q", registered.User.Email)
	}
	if registered.User.EmailVerified {
		t.Fatal("expected user to remain unverified when email delivery fails")
	}
	if registered.Delivery != "email-failed" {
		t.Fatalf("delivery = %q, want email-failed", registered.Delivery)
	}
	if registered.Warning == "" {
		t.Fatal("expected warning")
	}

	loginRecorder := performLogin(router, "mail-fail@example.com", "password")
	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("expected created account to login, got %d with body %q", loginRecorder.Code, loginRecorder.Body.String())
	}
}

func TestForgotPasswordSendsResetEmailWithoutExposingToken(t *testing.T) {
	store := NewMemoryStore()
	emailSender := &fakeEmailSender{}
	router := gin.New()
	RegisterRoutesWithSettingsAndEmailSender(router, store, nil, emailSender)

	request := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBufferString(`{
		"email":"linyi@example.com"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected forgot password status 200, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	var response struct {
		OK         bool   `json:"ok"`
		ResetToken string `json:"resetToken"`
		Delivery   string `json:"delivery"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode forgot password response: %v", err)
	}
	if !response.OK {
		t.Fatal("expected ok response")
	}
	if response.ResetToken != "" {
		t.Fatal("did not expect reset token in response when email sender is configured")
	}
	if response.Delivery != "email" {
		t.Fatalf("delivery = %q, want email", response.Delivery)
	}
	if emailSender.passwordResetToken == "" {
		t.Fatal("expected password reset token to be sent by email sender")
	}
	if emailSender.passwordResetUser.Email != "linyi@example.com" {
		t.Fatalf("password reset email user = %q, want linyi@example.com", emailSender.passwordResetUser.Email)
	}

	if err := store.ResetPassword(emailSender.passwordResetToken, "new-password"); err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}
	if _, _, err := store.Authenticate("linyi@example.com", "new-password"); err != nil {
		t.Fatalf("Authenticate with reset password returned error: %v", err)
	}
}

func TestForgotPasswordRejectsUnknownEmail(t *testing.T) {
	store := NewMemoryStore()
	emailSender := &fakeEmailSender{}
	router := gin.New()
	RegisterRoutesWithSettingsAndEmailSender(router, store, nil, emailSender)

	request := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBufferString(`{
		"email":"missing@example.com"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected forgot password status 404, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode forgot password response: %v", err)
	}
	if response.Error != "email is not registered" {
		t.Fatalf("error = %q, want email is not registered", response.Error)
	}
	if emailSender.passwordResetToken != "" {
		t.Fatal("did not expect email sender to receive token for unknown email")
	}
}

func TestRegisterRequiresTurnstileTokenWhenEnabled(t *testing.T) {
	store := NewMemoryStore()
	router := gin.New()
	RegisterRoutesWithDependencies(router, store, staticSettingsReader{
		settings: SecuritySettings{
			SessionDays:        7,
			TurnstileEnabled:   true,
			TurnstileSecretKey: "secret-key",
			TurnstileRegister:  true,
		},
	}, nil, &fakeTurnstileVerifier{ok: true})

	request := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{
		"email":"missing-turnstile@example.com",
		"password":"password",
		"displayName":"Missing Turnstile"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected register status 400, got %d with body %q", recorder.Code, recorder.Body.String())
	}
}

func TestRegisterVerifiesTurnstileToken(t *testing.T) {
	store := NewMemoryStore()
	verifier := &fakeTurnstileVerifier{ok: true}
	router := gin.New()
	RegisterRoutesWithDependencies(router, store, staticSettingsReader{
		settings: SecuritySettings{
			SessionDays:        7,
			TurnstileEnabled:   true,
			TurnstileSecretKey: "secret-key",
			TurnstileRegister:  true,
		},
	}, nil, verifier)

	request := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{
		"email":"valid-turnstile@example.com",
		"password":"password",
		"displayName":"Valid Turnstile",
		"turnstileToken":"client-token"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d with body %q", recorder.Code, recorder.Body.String())
	}
	if verifier.secret != "secret-key" || verifier.token != "client-token" {
		t.Fatalf("turnstile verifier got secret/token %q/%q", verifier.secret, verifier.token)
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
