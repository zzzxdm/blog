package server

import (
	"bytes"
	"encoding/json"
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

func TestSubmissionReviewPublishesPostAndCreatesMessage(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
	})

	userCookies := loginForTest(t, router, "linyi@example.com", "password")

	createReq := httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString(`{
		"title":"审核通过后公开的测试投稿",
		"summary":"这是一篇用于验证投稿审核闭环的文章。",
		"content":"用户提交文章后，管理员审核通过，文章应该进入公开文章列表，同时用户收到站内信。",
		"category":"工程实践",
		"tags":["投稿","审核"],
		"slug":"approved-submission-test",
		"submit":true
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range userCookies {
		createReq.AddCookie(cookie)
	}
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected submission created, got status=%d body=%q", createRec.Code, createRec.Body.String())
	}

	var created struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created submission: %v", err)
	}
	if created.ID == "" || created.Status != "submitted" {
		t.Fatalf("expected submitted submission, got %+v", created)
	}

	userAdminReq := httptest.NewRequest(http.MethodGet, "/api/admin/submissions", nil)
	for _, cookie := range userCookies {
		userAdminReq.AddCookie(cookie)
	}
	userAdminRec := httptest.NewRecorder()
	router.ServeHTTP(userAdminRec, userAdminReq)
	if userAdminRec.Code != http.StatusForbidden {
		t.Fatalf("expected non-admin status 403, got %d", userAdminRec.Code)
	}

	adminCookies := loginForTest(t, router, "admin@example.com", "password")

	reviewReq := httptest.NewRequest(http.MethodPost, "/api/admin/submissions/"+created.ID+"/review", bytes.NewBufferString(`{
		"action":"approve",
		"note":"内容结构清楚，可以发布。",
		"slug":"approved-submission-test",
		"category":"工程实践"
	}`))
	reviewReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		reviewReq.AddCookie(cookie)
	}
	reviewRec := httptest.NewRecorder()
	router.ServeHTTP(reviewRec, reviewReq)

	if reviewRec.Code != http.StatusOK {
		t.Fatalf("expected review status 200, got %d body=%q", reviewRec.Code, reviewRec.Body.String())
	}
	if !strings.Contains(reviewRec.Body.String(), `"status":"published"`) {
		t.Fatalf("expected published submission, got %q", reviewRec.Body.String())
	}

	postReq := httptest.NewRequest(http.MethodGet, "/api/posts/approved-submission-test", nil)
	postRec := httptest.NewRecorder()
	router.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusOK || !strings.Contains(postRec.Body.String(), "审核通过后公开的测试投稿") {
		t.Fatalf("expected published post, got status=%d body=%q", postRec.Code, postRec.Body.String())
	}

	messagesReq := httptest.NewRequest(http.MethodGet, "/api/messages", nil)
	for _, cookie := range userCookies {
		messagesReq.AddCookie(cookie)
	}
	messagesRec := httptest.NewRecorder()
	router.ServeHTTP(messagesRec, messagesReq)
	if messagesRec.Code != http.StatusOK || !strings.Contains(messagesRec.Body.String(), "你的投稿已通过并发布") {
		t.Fatalf("expected review message, got status=%d body=%q", messagesRec.Code, messagesRec.Body.String())
	}
}

func loginForTest(t *testing.T, router http.Handler, email string, password string) []*http.Cookie {
	t.Helper()

	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"`+email+`","password":"`+password+`"}`))
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

	return cookies
}
