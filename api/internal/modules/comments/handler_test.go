package comments

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/operations"

	"github.com/gin-gonic/gin"
)

func TestCreateCommentRejectsBlockedWords(t *testing.T) {
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
	settings.BlockedWords = []string{"推广"}
	settings.CommentsEnabled = true
	if _, err := settingsRepo.UpdateSettings(context.Background(), settings); err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}

	commentRepo := NewMemoryRepository()
	before, err := commentRepo.List(context.Background(), "blog-system-design", "user_linyi")
	if err != nil {
		t.Fatalf("List before returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(store))
	RegisterRoutes(router, commentRepo, settingsRepo)

	req := httptest.NewRequest(http.MethodPost, "/posts/blog-system-design/comments", bytes.NewBufferString(`{"body":"这是一条推广评论"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d with body %q", rec.Code, rec.Body.String())
	}

	after, err := commentRepo.List(context.Background(), "blog-system-design", "user_linyi")
	if err != nil {
		t.Fatalf("List after returned error: %v", err)
	}
	if after.Total != before.Total {
		t.Fatalf("comment total = %d, want unchanged %d", after.Total, before.Total)
	}
}

func TestCreateReplyAliasUsesParentPost(t *testing.T) {
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
	settings.CommentsEnabled = true
	if _, err := settingsRepo.UpdateSettings(context.Background(), settings); err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}

	commentRepo := NewMemoryRepository()
	router := gin.New()
	router.Use(auth.Middleware(store))
	RegisterRoutes(router, commentRepo, settingsRepo)

	req := httptest.NewRequest(http.MethodPost, "/comments/comment_001/replies", bytes.NewBufferString(`{"body":"同意这个观点"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d with body %q", rec.Code, rec.Body.String())
	}

	result, err := commentRepo.List(context.Background(), "blog-system-design", "user_linyi")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	found := false
	for _, comment := range result.Items {
		if comment.ParentID == "comment_001" && comment.Body == "同意这个观点" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected reply under parent comment_001")
	}
}

func TestDeleteMineMarksOwnCommentDeleted(t *testing.T) {
	store := auth.NewMemoryStore()
	_, token, err := store.Authenticate("linyi@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	commentRepo := NewMemoryRepository()
	router := gin.New()
	router.Use(auth.Middleware(store))
	RegisterRoutes(router, commentRepo, operations.NewMemoryRepository())

	req := httptest.NewRequest(http.MethodDelete, "/comments/comment_003", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %q", rec.Code, rec.Body.String())
	}

	result, err := commentRepo.ListByAuthor(context.Background(), "user_linyi", ListQuery{Status: "deleted"})
	if err != nil {
		t.Fatalf("ListByAuthor returned error: %v", err)
	}
	if result.Total != 1 || result.Items[0].ID != "comment_003" {
		t.Fatalf("deleted comments = %+v, want comment_003", result.Items)
	}
}

func TestContainsBlockedWord(t *testing.T) {
	if !containsBlockedWord("This has SPAM content", []string{"spam"}) {
		t.Fatal("expected case-insensitive blocked word match")
	}
	if containsBlockedWord("正常评论", []string{"推广"}) {
		t.Fatal("did not expect blocked word match")
	}
}
