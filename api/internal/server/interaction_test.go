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

func TestAccountCommentsBookmarksAndAdminModeration(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
	})

	userCookies := loginForTest(t, router, "linyi@example.com", "password")

	bookmarksReq := httptest.NewRequest(http.MethodGet, "/api/bookmarks/mine", nil)
	for _, cookie := range userCookies {
		bookmarksReq.AddCookie(cookie)
	}
	bookmarksRec := httptest.NewRecorder()
	router.ServeHTTP(bookmarksRec, bookmarksReq)
	if bookmarksRec.Code != http.StatusOK || !strings.Contains(bookmarksRec.Body.String(), "blog-system-design") {
		t.Fatalf("expected bookmark list, got status=%d body=%q", bookmarksRec.Code, bookmarksRec.Body.String())
	}

	removeBookmarkReq := httptest.NewRequest(http.MethodPut, "/api/posts/blog-system-design/bookmark", bytes.NewBufferString(`{"bookmarked":false}`))
	removeBookmarkReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range userCookies {
		removeBookmarkReq.AddCookie(cookie)
	}
	removeBookmarkRec := httptest.NewRecorder()
	router.ServeHTTP(removeBookmarkRec, removeBookmarkReq)
	if removeBookmarkRec.Code != http.StatusOK || !strings.Contains(removeBookmarkRec.Body.String(), `"bookmarked":false`) {
		t.Fatalf("expected bookmark removed, got status=%d body=%q", removeBookmarkRec.Code, removeBookmarkRec.Body.String())
	}

	commentReq := httptest.NewRequest(http.MethodPost, "/api/posts/blog-system-design/comments", bytes.NewBufferString(`{"body":"这条评论会进入后台审核。","parentId":""}`))
	commentReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range userCookies {
		commentReq.AddCookie(cookie)
	}
	commentRec := httptest.NewRecorder()
	router.ServeHTTP(commentRec, commentReq)
	if commentRec.Code != http.StatusCreated {
		t.Fatalf("expected comment created, got status=%d body=%q", commentRec.Code, commentRec.Body.String())
	}

	var createdComment struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(commentRec.Body.Bytes(), &createdComment); err != nil {
		t.Fatalf("decode created comment: %v", err)
	}

	mineReq := httptest.NewRequest(http.MethodGet, "/api/comments/mine", nil)
	for _, cookie := range userCookies {
		mineReq.AddCookie(cookie)
	}
	mineRec := httptest.NewRecorder()
	router.ServeHTTP(mineRec, mineReq)
	if mineRec.Code != http.StatusOK || !strings.Contains(mineRec.Body.String(), createdComment.ID) {
		t.Fatalf("expected my comments list, got status=%d body=%q", mineRec.Code, mineRec.Body.String())
	}

	adminCookies := loginForTest(t, router, "admin@example.com", "password")
	adminReq := httptest.NewRequest(http.MethodGet, "/api/admin/comments?status=pending", nil)
	for _, cookie := range adminCookies {
		adminReq.AddCookie(cookie)
	}
	adminRec := httptest.NewRecorder()
	router.ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusOK || !strings.Contains(adminRec.Body.String(), createdComment.ID) {
		t.Fatalf("expected admin comments list, got status=%d body=%q", adminRec.Code, adminRec.Body.String())
	}

	approveReq := httptest.NewRequest(http.MethodPut, "/api/admin/comments/"+createdComment.ID+"/status", bytes.NewBufferString(`{"status":"approved"}`))
	approveReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		approveReq.AddCookie(cookie)
	}
	approveRec := httptest.NewRecorder()
	router.ServeHTTP(approveRec, approveReq)
	if approveRec.Code != http.StatusOK || !strings.Contains(approveRec.Body.String(), `"status":"approved"`) {
		t.Fatalf("expected approved comment, got status=%d body=%q", approveRec.Code, approveRec.Body.String())
	}

	publicCommentsReq := httptest.NewRequest(http.MethodGet, "/api/posts/blog-system-design/comments", nil)
	publicCommentsRec := httptest.NewRecorder()
	router.ServeHTTP(publicCommentsRec, publicCommentsReq)
	if publicCommentsRec.Code != http.StatusOK || !strings.Contains(publicCommentsRec.Body.String(), "这条评论会进入后台审核。") {
		t.Fatalf("expected approved comment public, got status=%d body=%q", publicCommentsRec.Code, publicCommentsRec.Body.String())
	}
}

func TestAdminOperationsAPIs(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
	})

	anonReq := httptest.NewRequest(http.MethodGet, "/api/admin/settings", nil)
	anonRec := httptest.NewRecorder()
	router.ServeHTTP(anonRec, anonReq)
	if anonRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected anonymous settings status 401, got %d", anonRec.Code)
	}

	userCookies := loginForTest(t, router, "linyi@example.com", "password")
	userReq := httptest.NewRequest(http.MethodGet, "/api/admin/navigation", nil)
	for _, cookie := range userCookies {
		userReq.AddCookie(cookie)
	}
	userRec := httptest.NewRecorder()
	router.ServeHTTP(userRec, userReq)
	if userRec.Code != http.StatusForbidden {
		t.Fatalf("expected user navigation status 403, got %d", userRec.Code)
	}

	adminCookies := loginForTest(t, router, "admin@example.com", "password")

	settingsReq := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewBufferString(`{
		"siteName":"云间笔记 Pro",
		"siteDescription":"更新后的站点描述",
		"siteUrl":"https://blog.example.com",
		"beian":"京ICP备00000000号",
		"themePrimary":"#295b4b",
		"homepageLayout":"专题优先",
		"darkModeEnabled":true,
		"readingProgressEnabled":true,
		"commentsEnabled":true,
		"loginRequiredForComment":true,
		"autoApproveComments":false,
		"blockedWords":["推广"],
		"submissionsEnabled":true,
		"submissionManualReview":true,
		"submissionLimit":"每天最多 3 篇",
		"submissionGuide":"保持原创。",
		"mailEnabled":false,
		"mailProvider":"Resend",
		"fromEmail":"newsletter@example.com",
		"adminTwoFactorRequired":true,
		"loginFailureLock":true,
		"sessionDays":7,
		"backupCycle":"每日全量备份",
		"backupRetentionDays":7
	}`))
	settingsReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		settingsReq.AddCookie(cookie)
	}
	settingsRec := httptest.NewRecorder()
	router.ServeHTTP(settingsRec, settingsReq)
	if settingsRec.Code != http.StatusOK || !strings.Contains(settingsRec.Body.String(), "云间笔记 Pro") {
		t.Fatalf("expected settings updated, got status=%d body=%q", settingsRec.Code, settingsRec.Body.String())
	}

	navigationReq := httptest.NewRequest(http.MethodPut, "/api/admin/navigation", bytes.NewBufferString(`{
		"topItems":[{"id":"top_1","label":"首页","url":"/","order":1},{"id":"top_2","label":"归档","url":"/archive","order":2}],
		"footerItems":[{"id":"footer_1","label":"RSS","url":"/rss.xml","order":1}],
		"mobileCollapse":true,
		"externalLinksNewWindow":true,
		"showLoginEntry":true,
		"githubUrl":"https://github.com/example",
		"contactEmail":"hello@example.com",
		"rssUrl":"/rss.xml",
		"redirects":[{"from":"/old","to":"/new","code":301}]
	}`))
	navigationReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		navigationReq.AddCookie(cookie)
	}
	navigationRec := httptest.NewRecorder()
	router.ServeHTTP(navigationRec, navigationReq)
	if navigationRec.Code != http.StatusOK || !strings.Contains(navigationRec.Body.String(), `"label":"归档"`) {
		t.Fatalf("expected navigation updated, got status=%d body=%q", navigationRec.Code, navigationRec.Body.String())
	}

	mediaReq := httptest.NewRequest(http.MethodGet, "/api/admin/media", nil)
	for _, cookie := range adminCookies {
		mediaReq.AddCookie(cookie)
	}
	mediaRec := httptest.NewRecorder()
	router.ServeHTTP(mediaRec, mediaReq)
	if mediaRec.Code != http.StatusOK || !strings.Contains(mediaRec.Body.String(), "cover-code-desk.jpg") {
		t.Fatalf("expected media list, got status=%d body=%q", mediaRec.Code, mediaRec.Body.String())
	}

	statsReq := httptest.NewRequest(http.MethodGet, "/api/admin/stats", nil)
	for _, cookie := range adminCookies {
		statsReq.AddCookie(cookie)
	}
	statsRec := httptest.NewRecorder()
	router.ServeHTTP(statsRec, statsReq)
	if statsRec.Code != http.StatusOK || !strings.Contains(statsRec.Body.String(), `"label":"PV"`) {
		t.Fatalf("expected stats response, got status=%d body=%q", statsRec.Code, statsRec.Body.String())
	}
}

func TestUsersAndAccountSettingsAPIs(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
	})

	userCookies := loginForTest(t, router, "linyi@example.com", "password")
	userAdminReq := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	for _, cookie := range userCookies {
		userAdminReq.AddCookie(cookie)
	}
	userAdminRec := httptest.NewRecorder()
	router.ServeHTTP(userAdminRec, userAdminReq)
	if userAdminRec.Code != http.StatusForbidden {
		t.Fatalf("expected user admin list status 403, got %d", userAdminRec.Code)
	}

	accountReq := httptest.NewRequest(http.MethodPut, "/api/account/settings", bytes.NewBufferString(`{
		"displayName":"林一新版",
		"username":"linyi",
		"email":"linyi@example.com",
		"avatarText":"林",
		"bio":"更新后的个人简介",
		"twoFactor":true,
		"loginAlert":true,
		"notifyReview":true,
		"notifyComment":true,
		"notifyAnnouncement":true,
		"emailNotification":false,
		"publicProfile":true,
		"publicBookmarks":false,
		"profileUrl":"https://blog.example.com/authors/linyi",
		"timezone":"Asia/Shanghai"
	}`))
	accountReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range userCookies {
		accountReq.AddCookie(cookie)
	}
	accountRec := httptest.NewRecorder()
	router.ServeHTTP(accountRec, accountReq)
	if accountRec.Code != http.StatusOK || !strings.Contains(accountRec.Body.String(), "林一新版") {
		t.Fatalf("expected account settings updated, got status=%d body=%q", accountRec.Code, accountRec.Body.String())
	}

	adminCookies := loginForTest(t, router, "admin@example.com", "password")
	adminListReq := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	for _, cookie := range adminCookies {
		adminListReq.AddCookie(cookie)
	}
	adminListRec := httptest.NewRecorder()
	router.ServeHTTP(adminListRec, adminListReq)
	if adminListRec.Code != http.StatusOK || !strings.Contains(adminListRec.Body.String(), "market_user") {
		t.Fatalf("expected admin users list, got status=%d body=%q", adminListRec.Code, adminListRec.Body.String())
	}

	muteReq := httptest.NewRequest(http.MethodPut, "/api/admin/users/user_linyi/status", bytes.NewBufferString(`{"status":"muted"}`))
	muteReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		muteReq.AddCookie(cookie)
	}
	muteRec := httptest.NewRecorder()
	router.ServeHTTP(muteRec, muteReq)
	if muteRec.Code != http.StatusOK || !strings.Contains(muteRec.Body.String(), `"status":"muted"`) {
		t.Fatalf("expected muted user, got status=%d body=%q", muteRec.Code, muteRec.Body.String())
	}
}

func TestAdminPostSaveAndPublish(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
	})

	adminCookies := loginForTest(t, router, "admin@example.com", "password")

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/posts", bytes.NewBufferString(`{
		"title":"后台发布流程验证",
		"summary":"验证管理员保存草稿后发布到前台。",
		"content":"后台编辑器保存草稿后，发布动作应该调用公开文章发布能力。",
		"status":"draft",
		"category":"工程实践",
		"tags":["后台","发布"],
		"slug":"admin-publish-flow-check",
		"coverImage":"https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=1200&q=80",
		"seoTitle":"后台发布流程验证",
		"seoDescription":"验证管理员保存草稿后发布到前台。"
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		createReq.AddCookie(cookie)
	}
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected admin post created, got status=%d body=%q", createRec.Code, createRec.Body.String())
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode admin post: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected admin post id, got %q", createRec.Body.String())
	}

	publishReq := httptest.NewRequest(http.MethodPost, "/api/admin/posts/"+created.ID+"/publish", nil)
	for _, cookie := range adminCookies {
		publishReq.AddCookie(cookie)
	}
	publishRec := httptest.NewRecorder()
	router.ServeHTTP(publishRec, publishReq)
	if publishRec.Code != http.StatusOK || !strings.Contains(publishRec.Body.String(), `"status":"published"`) {
		t.Fatalf("expected admin post published, got status=%d body=%q", publishRec.Code, publishRec.Body.String())
	}

	publicReq := httptest.NewRequest(http.MethodGet, "/api/posts/admin-publish-flow-check", nil)
	publicRec := httptest.NewRecorder()
	router.ServeHTTP(publicRec, publicReq)
	if publicRec.Code != http.StatusOK || !strings.Contains(publicRec.Body.String(), "后台发布流程验证") {
		t.Fatalf("expected public post after publish, got status=%d body=%q", publicRec.Code, publicRec.Body.String())
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
