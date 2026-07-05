package submissions

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/messages"
	"blog/api/internal/modules/operations"

	"github.com/gin-gonic/gin"
)

func TestCanReviewSubmission(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{status: StatusSubmitted, want: true},
		{status: StatusReturned, want: true},
		{status: StatusDraft, want: false},
		{status: StatusRejected, want: false},
		{status: StatusPublished, want: false},
	}

	for _, tt := range cases {
		t.Run(tt.status, func(t *testing.T) {
			if got := canReviewSubmission(tt.status); got != tt.want {
				t.Fatalf("canReviewSubmission(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestCreateSubmissionRejectsBlockedWords(t *testing.T) {
	store := auth.NewMemoryStore()
	_, token, err := store.Authenticate("linyi@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	settingsRepo := operations.NewMemoryRepository()
	settings, err := settingsRepo.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	settings.SubmissionsEnabled = true
	settings.BlockedWords = []string{"推广"}
	if _, err := settingsRepo.UpdateSettings(context.Background(), settings); err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}

	submissionRepo := NewMemoryRepository()
	before, err := submissionRepo.ListByAuthor(context.Background(), "user_linyi", ListQuery{})
	if err != nil {
		t.Fatalf("ListByAuthor before returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(store))
	RegisterRoutes(router, submissionRepo, messages.NewMemoryRepository(), nil, settingsRepo)

	req := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(`{
		"title":"推广文案",
		"summary":"不应该进入投稿库",
		"content":"这篇投稿包含被屏蔽内容。",
		"category":"工程实践",
		"tags":["推广"],
		"coverImage":"",
		"slug":"blocked-submission",
		"submit":true
	}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d with body %q", rec.Code, rec.Body.String())
	}

	after, err := submissionRepo.ListByAuthor(context.Background(), "user_linyi", ListQuery{})
	if err != nil {
		t.Fatalf("ListByAuthor after returned error: %v", err)
	}
	if after.Total != before.Total {
		t.Fatalf("submission total = %d, want unchanged %d", after.Total, before.Total)
	}
}

func TestSubmissionLimitAllowsDraftButRejectsSubmit(t *testing.T) {
	store := auth.NewMemoryStore()
	user, token, err := store.Register(auth.RegisterRequest{
		Email:       "limit@example.com",
		Password:    "password",
		DisplayName: "Limit User",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	settingsRepo := operations.NewMemoryRepository()
	settings, err := settingsRepo.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	settings.SubmissionsEnabled = true
	settings.BlockedWords = nil
	settings.SubmissionLimit = "\u6bcf\u5929\u6700\u591a 1 \u7bc7"
	if _, err := settingsRepo.UpdateSettings(context.Background(), settings); err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}

	submissionRepo := NewMemoryRepository()
	before, err := submissionRepo.ListByAuthor(context.Background(), user.ID, ListQuery{})
	if err != nil {
		t.Fatalf("ListByAuthor before returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(store))
	RegisterRoutes(router, submissionRepo, messages.NewMemoryRepository(), nil, settingsRepo)

	draftReq := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(`{
		"title":"Draft allowed",
		"summary":"Drafts do not enter review",
		"content":"",
		"category":"Engineering",
		"tags":["draft"],
		"coverImage":"",
		"slug":"draft-allowed",
		"submit":false
	}`))
	draftReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	draftRec := httptest.NewRecorder()
	router.ServeHTTP(draftRec, draftReq)
	if draftRec.Code != http.StatusCreated {
		t.Fatalf("expected draft status 201, got %d with body %q", draftRec.Code, draftRec.Body.String())
	}

	firstSubmitReq := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(`{
		"title":"First submission",
		"summary":"This should use the daily slot",
		"content":"Ready for review.",
		"category":"Engineering",
		"tags":["limit"],
		"coverImage":"",
		"slug":"first-submission",
		"submit":true
	}`))
	firstSubmitReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	firstSubmitRec := httptest.NewRecorder()
	router.ServeHTTP(firstSubmitRec, firstSubmitReq)
	if firstSubmitRec.Code != http.StatusCreated {
		t.Fatalf("expected first submit status 201, got %d with body %q", firstSubmitRec.Code, firstSubmitRec.Body.String())
	}

	submitReq := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(`{
		"title":"Limited submission",
		"summary":"This should be blocked by the daily limit",
		"content":"Ready for review.",
		"category":"Engineering",
		"tags":["limit"],
		"coverImage":"",
		"slug":"limited-submission",
		"submit":true
	}`))
	submitReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	submitRec := httptest.NewRecorder()
	router.ServeHTTP(submitRec, submitReq)
	if submitRec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected submit status 429, got %d with body %q", submitRec.Code, submitRec.Body.String())
	}

	after, err := submissionRepo.ListByAuthor(context.Background(), user.ID, ListQuery{})
	if err != nil {
		t.Fatalf("ListByAuthor after returned error: %v", err)
	}
	if after.Total != before.Total+2 {
		t.Fatalf("submission total = %d, want draft and first submit total %d", after.Total, before.Total+2)
	}
}

func TestDeleteMineRemovesUnpublishedSubmission(t *testing.T) {
	store := auth.NewMemoryStore()
	_, token, err := store.Authenticate("linyi@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	submissionRepo := NewMemoryRepository()
	before, err := submissionRepo.ListByAuthor(context.Background(), "user_linyi", ListQuery{})
	if err != nil {
		t.Fatalf("ListByAuthor before returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(store))
	RegisterRoutes(router, submissionRepo, messages.NewMemoryRepository(), nil, operations.NewMemoryRepository())

	req := httptest.NewRequest(http.MethodDelete, "/submissions/submission_001", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %q", rec.Code, rec.Body.String())
	}

	after, err := submissionRepo.ListByAuthor(context.Background(), "user_linyi", ListQuery{})
	if err != nil {
		t.Fatalf("ListByAuthor after returned error: %v", err)
	}
	if after.Total != before.Total-1 {
		t.Fatalf("submission total = %d, want %d", after.Total, before.Total-1)
	}
	if _, err := submissionRepo.Get(context.Background(), "submission_001"); !errors.Is(err, ErrSubmissionNotFound) {
		t.Fatalf("Get deleted submission error = %v, want ErrSubmissionNotFound", err)
	}
}

func TestDeleteMineRejectsPublishedSubmission(t *testing.T) {
	store := auth.NewMemoryStore()
	_, token, err := store.Authenticate("linyi@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	submissionRepo := NewMemoryRepository()
	router := gin.New()
	router.Use(auth.Middleware(store))
	RegisterRoutes(router, submissionRepo, messages.NewMemoryRepository(), nil, operations.NewMemoryRepository())

	req := httptest.NewRequest(http.MethodDelete, "/submissions/submission_004", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d with body %q", rec.Code, rec.Body.String())
	}
}

func TestSubmissionLimitWindow(t *testing.T) {
	now := time.Date(2026, 7, 5, 15, 30, 0, 0, time.UTC)

	since, limit, ok := submissionLimitWindow("\u6bcf\u5929\u6700\u591a 3 \u7bc7", now)
	if !ok || limit != 3 || !since.Equal(time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("daily window = %v/%d/%v, want start of day limit 3", since, limit, ok)
	}

	since, limit, ok = submissionLimitWindow("\u6bcf\u5468\u6700\u591a 3 \u7bc7", now)
	if !ok || limit != 3 || !since.Equal(time.Date(2026, 6, 29, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("weekly window = %v/%d/%v, want monday start limit 3", since, limit, ok)
	}

	if _, _, ok := submissionLimitWindow("unlimited", now); ok {
		t.Fatal("expected non-numeric limit to be ignored")
	}
}

func TestSaveRequestContainsBlockedWord(t *testing.T) {
	request := SaveRequest{
		Title: "正常标题",
		Tags:  []string{"Workflow", "SPAM"},
	}
	if !saveRequestContainsBlockedWord(request, []string{"spam"}) {
		t.Fatal("expected blocked word match in tags")
	}
	if saveRequestContainsBlockedWord(request, []string{"推广"}) {
		t.Fatal("did not expect blocked word match")
	}
}
