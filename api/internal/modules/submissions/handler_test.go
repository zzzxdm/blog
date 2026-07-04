package submissions

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
