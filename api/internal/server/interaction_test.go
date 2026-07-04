package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blog/api/internal/config"
)

func TestAuthCommentAndReactionFlow(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
	})

	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"linyi@example.com","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d body=%q", loginRec.Code, loginRec.Body.String())
	}

	cookies := loginRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie after login")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	for _, cookie := range cookies {
		meReq.AddCookie(cookie)
	}
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)

	if meRec.Code != http.StatusOK || !strings.Contains(meRec.Body.String(), `"displayName":"林一"`) {
		t.Fatalf("expected current user response, got status=%d body=%q", meRec.Code, meRec.Body.String())
	}

	anonCommentReq := httptest.NewRequest(http.MethodPost, "/api/posts/blog-system-design/comments", bytes.NewBufferString(`{"body":"未登录评论"}`))
	anonCommentReq.Header.Set("Content-Type", "application/json")
	anonCommentRec := httptest.NewRecorder()
	router.ServeHTTP(anonCommentRec, anonCommentReq)

	if anonCommentRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected anonymous comment status 401, got %d", anonCommentRec.Code)
	}

	commentReq := httptest.NewRequest(http.MethodPost, "/api/posts/blog-system-design/comments", bytes.NewBufferString(`{"body":"审核结果是否会同步到站内信？"}`))
	commentReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		commentReq.AddCookie(cookie)
	}
	commentRec := httptest.NewRecorder()
	router.ServeHTTP(commentRec, commentReq)

	if commentRec.Code != http.StatusCreated {
		t.Fatalf("expected created comment, got status=%d body=%q", commentRec.Code, commentRec.Body.String())
	}
	if !strings.Contains(commentRec.Body.String(), `"status":"pending"`) {
		t.Fatalf("expected pending comment status, got %q", commentRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/posts/blog-system-design/comments", nil)
	for _, cookie := range cookies {
		listReq.AddCookie(cookie)
	}
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK || !strings.Contains(listRec.Body.String(), "审核结果是否会同步到站内信？") {
		t.Fatalf("expected own pending comment in list, got status=%d body=%q", listRec.Code, listRec.Body.String())
	}

	reactionReq := httptest.NewRequest(http.MethodPut, "/api/posts/blog-system-design/reaction", bytes.NewBufferString(`{"type":"dislike"}`))
	reactionReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		reactionReq.AddCookie(cookie)
	}
	reactionRec := httptest.NewRecorder()
	router.ServeHTTP(reactionRec, reactionReq)

	if reactionRec.Code != http.StatusOK {
		t.Fatalf("expected reaction status 200, got %d body=%q", reactionRec.Code, reactionRec.Body.String())
	}
	if !strings.Contains(reactionRec.Body.String(), `"myReaction":"dislike"`) {
		t.Fatalf("expected dislike reaction, got %q", reactionRec.Body.String())
	}
}
