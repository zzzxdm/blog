package server

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
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

	updateSubmissionReq := httptest.NewRequest(http.MethodPut, "/api/admin/submissions/"+created.ID, bytes.NewBufferString(`{
		"title":"管理员修订后的测试投稿",
		"summary":"管理员在审核台修订摘要后再发布。",
		"content":"这篇投稿已经由管理员修订正文，发布时应该使用修订后的内容。",
		"category":"工程实践",
		"tags":["投稿","审核","修订"],
		"slug":"approved-submission-test"
	}`))
	updateSubmissionReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		updateSubmissionReq.AddCookie(cookie)
	}
	updateSubmissionRec := httptest.NewRecorder()
	router.ServeHTTP(updateSubmissionRec, updateSubmissionReq)
	if updateSubmissionRec.Code != http.StatusOK || !strings.Contains(updateSubmissionRec.Body.String(), "管理员修订后的测试投稿") || !strings.Contains(updateSubmissionRec.Body.String(), `"version":2`) {
		t.Fatalf("expected admin submission update, got status=%d body=%q", updateSubmissionRec.Code, updateSubmissionRec.Body.String())
	}

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
	if postRec.Code != http.StatusOK || !strings.Contains(postRec.Body.String(), "管理员修订后的测试投稿") || !strings.Contains(postRec.Body.String(), "修订后的内容") {
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

	adminMessagesExportReq := httptest.NewRequest(http.MethodGet, "/api/admin/messages/export", nil)
	for _, cookie := range adminCookies {
		adminMessagesExportReq.AddCookie(cookie)
	}
	adminMessagesExportRec := httptest.NewRecorder()
	router.ServeHTTP(adminMessagesExportRec, adminMessagesExportReq)
	if adminMessagesExportRec.Code != http.StatusOK || !strings.Contains(adminMessagesExportRec.Body.String(), `"scope":"messages"`) || !strings.Contains(adminMessagesExportRec.Body.String(), "你的投稿已通过并发布") {
		t.Fatalf("expected messages export, got status=%d body=%q", adminMessagesExportRec.Code, adminMessagesExportRec.Body.String())
	}

	scheduledMessageReq := httptest.NewRequest(http.MethodPost, "/api/admin/messages", bytes.NewBufferString(`{
		"recipientId":"user_linyi",
		"recipientName":"林一",
		"type":"admin",
		"priority":"normal",
		"title":"明天发送的站内信",
		"body":"这条消息应该先停留在后台定时列表。",
		"targetType":"admin-message",
		"targetTitle":"定时发送",
		"scheduledAt":"2099-01-01T09:00:00Z"
	}`))
	scheduledMessageReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		scheduledMessageReq.AddCookie(cookie)
	}
	scheduledMessageRec := httptest.NewRecorder()
	router.ServeHTTP(scheduledMessageRec, scheduledMessageReq)
	if scheduledMessageRec.Code != http.StatusCreated || !strings.Contains(scheduledMessageRec.Body.String(), `"status":"scheduled"`) {
		t.Fatalf("expected scheduled admin message, got status=%d body=%q", scheduledMessageRec.Code, scheduledMessageRec.Body.String())
	}

	adminScheduledReq := httptest.NewRequest(http.MethodGet, "/api/admin/messages?status=scheduled", nil)
	for _, cookie := range adminCookies {
		adminScheduledReq.AddCookie(cookie)
	}
	adminScheduledRec := httptest.NewRecorder()
	router.ServeHTTP(adminScheduledRec, adminScheduledReq)
	if adminScheduledRec.Code != http.StatusOK || !strings.Contains(adminScheduledRec.Body.String(), "明天发送的站内信") {
		t.Fatalf("expected scheduled message in admin list, got status=%d body=%q", adminScheduledRec.Code, adminScheduledRec.Body.String())
	}

	userScheduledReq := httptest.NewRequest(http.MethodGet, "/api/messages", nil)
	for _, cookie := range userCookies {
		userScheduledReq.AddCookie(cookie)
	}
	userScheduledRec := httptest.NewRecorder()
	router.ServeHTTP(userScheduledRec, userScheduledReq)
	if userScheduledRec.Code != http.StatusOK || strings.Contains(userScheduledRec.Body.String(), "明天发送的站内信") {
		t.Fatalf("expected scheduled message hidden from user inbox, got status=%d body=%q", userScheduledRec.Code, userScheduledRec.Body.String())
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

	replyReq := httptest.NewRequest(http.MethodPost, "/api/posts/blog-system-design/comments", bytes.NewBufferString(`{"body":"这是对首条评论的回复。","parentId":"comment_001"}`))
	replyReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range userCookies {
		replyReq.AddCookie(cookie)
	}
	replyRec := httptest.NewRecorder()
	router.ServeHTTP(replyRec, replyReq)
	if replyRec.Code != http.StatusCreated || !strings.Contains(replyRec.Body.String(), `"parentId":"comment_001"`) {
		t.Fatalf("expected reply comment created, got status=%d body=%q", replyRec.Code, replyRec.Body.String())
	}

	likeCommentReq := httptest.NewRequest(http.MethodPut, "/api/comments/comment_001/like", nil)
	for _, cookie := range userCookies {
		likeCommentReq.AddCookie(cookie)
	}
	likeCommentRec := httptest.NewRecorder()
	router.ServeHTTP(likeCommentRec, likeCommentReq)
	if likeCommentRec.Code != http.StatusOK || !strings.Contains(likeCommentRec.Body.String(), `"liked":true`) {
		t.Fatalf("expected liked comment, got status=%d body=%q", likeCommentRec.Code, likeCommentRec.Body.String())
	}

	reportReq := httptest.NewRequest(http.MethodPost, "/api/comments/comment_001/report", bytes.NewBufferString(`{"reason":"包含不准确信息"}`))
	reportReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range userCookies {
		reportReq.AddCookie(cookie)
	}
	reportRec := httptest.NewRecorder()
	router.ServeHTTP(reportRec, reportReq)
	if reportRec.Code != http.StatusOK || !strings.Contains(reportRec.Body.String(), `"ok":true`) {
		t.Fatalf("expected comment report accepted, got status=%d body=%q", reportRec.Code, reportRec.Body.String())
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

	exportCommentsReq := httptest.NewRequest(http.MethodGet, "/api/admin/comments/export?status=pending", nil)
	for _, cookie := range adminCookies {
		exportCommentsReq.AddCookie(cookie)
	}
	exportCommentsRec := httptest.NewRecorder()
	router.ServeHTTP(exportCommentsRec, exportCommentsReq)
	if exportCommentsRec.Code != http.StatusOK || !strings.Contains(exportCommentsRec.Body.String(), `"scope":"comments"`) || !strings.Contains(exportCommentsRec.Body.String(), createdComment.ID) {
		t.Fatalf("expected comments export, got status=%d body=%q", exportCommentsRec.Code, exportCommentsRec.Body.String())
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
		UploadDir: t.TempDir(),
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

	testMailReq := httptest.NewRequest(http.MethodPost, "/api/admin/settings/test-mail", nil)
	for _, cookie := range adminCookies {
		testMailReq.AddCookie(cookie)
	}
	testMailRec := httptest.NewRecorder()
	router.ServeHTTP(testMailRec, testMailReq)
	if testMailRec.Code != http.StatusOK || !strings.Contains(testMailRec.Body.String(), `"provider":"Resend"`) || !strings.Contains(testMailRec.Body.String(), `"delivery":"dev-response"`) {
		t.Fatalf("expected test mail response, got status=%d body=%q", testMailRec.Code, testMailRec.Body.String())
	}

	backupReq := httptest.NewRequest(http.MethodPost, "/api/admin/backups", nil)
	for _, cookie := range adminCookies {
		backupReq.AddCookie(cookie)
	}
	backupRec := httptest.NewRecorder()
	router.ServeHTTP(backupRec, backupReq)
	if backupRec.Code != http.StatusOK || !strings.Contains(backupRec.Body.String(), `"status":"completed"`) {
		t.Fatalf("expected backup response, got status=%d body=%q", backupRec.Code, backupRec.Body.String())
	}

	var backupResult struct {
		ID       string `json:"id"`
		FileName string `json:"fileName"`
		Settings struct {
			LastBackupAt string `json:"lastBackupAt"`
		} `json:"settings"`
	}
	if err := json.Unmarshal(backupRec.Body.Bytes(), &backupResult); err != nil {
		t.Fatalf("decode backup response: %v", err)
	}
	if backupResult.ID == "" || backupResult.FileName == "" || backupResult.Settings.LastBackupAt == "" {
		t.Fatalf("expected backup metadata and updated settings, got %+v", backupResult)
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

	var uploadBody bytes.Buffer
	uploadWriter := multipart.NewWriter(&uploadBody)
	uploadPart, err := uploadWriter.CreateFormFile("file", "tiny.png")
	if err != nil {
		t.Fatalf("create upload form file: %v", err)
	}
	if _, err := uploadPart.Write(tinyPNG()); err != nil {
		t.Fatalf("write upload file: %v", err)
	}
	if err := uploadWriter.WriteField("alt", "测试上传图片"); err != nil {
		t.Fatalf("write upload alt: %v", err)
	}
	if err := uploadWriter.WriteField("category", "测试上传"); err != nil {
		t.Fatalf("write upload category: %v", err)
	}
	if err := uploadWriter.Close(); err != nil {
		t.Fatalf("close upload writer: %v", err)
	}

	uploadReq := httptest.NewRequest(http.MethodPost, "/api/admin/media", &uploadBody)
	uploadReq.Header.Set("Content-Type", uploadWriter.FormDataContentType())
	for _, cookie := range adminCookies {
		uploadReq.AddCookie(cookie)
	}
	uploadRec := httptest.NewRecorder()
	router.ServeHTTP(uploadRec, uploadReq)
	if uploadRec.Code != http.StatusCreated || !strings.Contains(uploadRec.Body.String(), `"category":"测试上传"`) {
		t.Fatalf("expected media uploaded, got status=%d body=%q", uploadRec.Code, uploadRec.Body.String())
	}

	var uploaded struct {
		ID     string `json:"id"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	if err := json.Unmarshal(uploadRec.Body.Bytes(), &uploaded); err != nil {
		t.Fatalf("decode uploaded media: %v", err)
	}
	if !strings.HasPrefix(uploaded.URL, "/uploads/") || uploaded.Width != 1 || uploaded.Height != 1 {
		t.Fatalf("expected uploaded media metadata, got %+v", uploaded)
	}

	mediaDetailReq := httptest.NewRequest(http.MethodGet, "/api/admin/media/"+uploaded.ID, nil)
	for _, cookie := range adminCookies {
		mediaDetailReq.AddCookie(cookie)
	}
	mediaDetailRec := httptest.NewRecorder()
	router.ServeHTTP(mediaDetailRec, mediaDetailReq)
	if mediaDetailRec.Code != http.StatusOK || !strings.Contains(mediaDetailRec.Body.String(), "tiny.png") {
		t.Fatalf("expected media detail, got status=%d body=%q", mediaDetailRec.Code, mediaDetailRec.Body.String())
	}

	updateMediaReq := httptest.NewRequest(http.MethodPatch, "/api/admin/media/"+uploaded.ID, bytes.NewBufferString(`{"alt":"新的 Alt 文本","category":"正文配图"}`))
	updateMediaReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		updateMediaReq.AddCookie(cookie)
	}
	updateMediaRec := httptest.NewRecorder()
	router.ServeHTTP(updateMediaRec, updateMediaReq)
	if updateMediaRec.Code != http.StatusOK || !strings.Contains(updateMediaRec.Body.String(), `"alt":"新的 Alt 文本"`) || !strings.Contains(updateMediaRec.Body.String(), `"category":"正文配图"`) {
		t.Fatalf("expected media metadata updated, got status=%d body=%q", updateMediaRec.Code, updateMediaRec.Body.String())
	}

	uploadedListReq := httptest.NewRequest(http.MethodGet, "/api/admin/media", nil)
	for _, cookie := range adminCookies {
		uploadedListReq.AddCookie(cookie)
	}
	uploadedListRec := httptest.NewRecorder()
	router.ServeHTTP(uploadedListRec, uploadedListReq)
	if uploadedListRec.Code != http.StatusOK || !strings.Contains(uploadedListRec.Body.String(), "tiny.png") {
		t.Fatalf("expected uploaded media in list, got status=%d body=%q", uploadedListRec.Code, uploadedListRec.Body.String())
	}

	staticReq := httptest.NewRequest(http.MethodGet, uploaded.URL, nil)
	staticRec := httptest.NewRecorder()
	router.ServeHTTP(staticRec, staticReq)
	if staticRec.Code != http.StatusOK {
		t.Fatalf("expected uploaded file to be served, got status=%d", staticRec.Code)
	}

	deleteUsedReq := httptest.NewRequest(http.MethodDelete, "/api/admin/media/media_001", nil)
	for _, cookie := range adminCookies {
		deleteUsedReq.AddCookie(cookie)
	}
	deleteUsedRec := httptest.NewRecorder()
	router.ServeHTTP(deleteUsedRec, deleteUsedReq)
	if deleteUsedRec.Code != http.StatusConflict {
		t.Fatalf("expected used media delete status 409, got %d body=%q", deleteUsedRec.Code, deleteUsedRec.Body.String())
	}

	deleteMediaReq := httptest.NewRequest(http.MethodDelete, "/api/admin/media/"+uploaded.ID, nil)
	for _, cookie := range adminCookies {
		deleteMediaReq.AddCookie(cookie)
	}
	deleteMediaRec := httptest.NewRecorder()
	router.ServeHTTP(deleteMediaRec, deleteMediaReq)
	if deleteMediaRec.Code != http.StatusOK || !strings.Contains(deleteMediaRec.Body.String(), `"ok":true`) {
		t.Fatalf("expected uploaded media deleted, got status=%d body=%q", deleteMediaRec.Code, deleteMediaRec.Body.String())
	}

	deletedStaticRec := httptest.NewRecorder()
	router.ServeHTTP(deletedStaticRec, staticReq)
	if deletedStaticRec.Code != http.StatusNotFound {
		t.Fatalf("expected uploaded file to be removed, got status=%d", deletedStaticRec.Code)
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

	statsExportReq := httptest.NewRequest(http.MethodGet, "/api/admin/stats/export", nil)
	for _, cookie := range adminCookies {
		statsExportReq.AddCookie(cookie)
	}
	statsExportRec := httptest.NewRecorder()
	router.ServeHTTP(statsExportRec, statsExportReq)
	if statsExportRec.Code != http.StatusOK || !strings.Contains(statsExportRec.Body.String(), `"scope":"stats"`) || !strings.Contains(statsExportRec.Body.String(), `"label":"PV"`) {
		t.Fatalf("expected stats export, got status=%d body=%q", statsExportRec.Code, statsExportRec.Body.String())
	}

	auditReq := httptest.NewRequest(http.MethodGet, "/api/admin/audit-logs?pageSize=20", nil)
	for _, cookie := range adminCookies {
		auditReq.AddCookie(cookie)
	}
	auditRec := httptest.NewRecorder()
	router.ServeHTTP(auditRec, auditReq)
	if auditRec.Code != http.StatusOK {
		t.Fatalf("expected audit logs response, got status=%d body=%q", auditRec.Code, auditRec.Body.String())
	}

	var auditLogs struct {
		Items []struct {
			Action       string `json:"action"`
			ActorName    string `json:"actorName"`
			ResourceType string `json:"resourceType"`
			Status       string `json:"status"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(auditRec.Body.Bytes(), &auditLogs); err != nil {
		t.Fatalf("decode audit logs: %v", err)
	}
	if auditLogs.Total == 0 {
		t.Fatalf("expected audit logs to be recorded, got %+v", auditLogs)
	}

	var foundSettingsUpdate bool
	for _, item := range auditLogs.Items {
		if item.Action == "settings.update" && item.ActorName == "管理员" && item.ResourceType == "settings" && item.Status == "success" {
			foundSettingsUpdate = true
			break
		}
	}
	if !foundSettingsUpdate {
		t.Fatalf("expected settings update audit log, got %+v", auditLogs)
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

	changePasswordReq := httptest.NewRequest(http.MethodPut, "/api/me/password", bytes.NewBufferString(`{"currentPassword":"password","newPassword":"new-password"}`))
	changePasswordReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range userCookies {
		changePasswordReq.AddCookie(cookie)
	}
	changePasswordRec := httptest.NewRecorder()
	router.ServeHTTP(changePasswordRec, changePasswordReq)
	if changePasswordRec.Code != http.StatusOK || !strings.Contains(changePasswordRec.Body.String(), `"ok":true`) {
		t.Fatalf("expected password changed, got status=%d body=%q", changePasswordRec.Code, changePasswordRec.Body.String())
	}

	oldPasswordReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"linyi@example.com","password":"password"}`))
	oldPasswordReq.Header.Set("Content-Type", "application/json")
	oldPasswordRec := httptest.NewRecorder()
	router.ServeHTTP(oldPasswordRec, oldPasswordReq)
	if oldPasswordRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password rejected, got %d body=%q", oldPasswordRec.Code, oldPasswordRec.Body.String())
	}

	newPasswordReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"linyi@example.com","password":"new-password"}`))
	newPasswordReq.Header.Set("Content-Type", "application/json")
	newPasswordRec := httptest.NewRecorder()
	router.ServeHTTP(newPasswordRec, newPasswordReq)
	if newPasswordRec.Code != http.StatusOK {
		t.Fatalf("expected new password accepted, got %d body=%q", newPasswordRec.Code, newPasswordRec.Body.String())
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

	inviteReq := httptest.NewRequest(http.MethodPost, "/api/admin/users/invitations", bytes.NewBufferString(`{"email":"writer@example.com","displayName":"特约作者","role":"author"}`))
	inviteReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		inviteReq.AddCookie(cookie)
	}
	inviteRec := httptest.NewRecorder()
	router.ServeHTTP(inviteRec, inviteReq)
	if inviteRec.Code != http.StatusCreated || !strings.Contains(inviteRec.Body.String(), `"role":"author"`) || !strings.Contains(inviteRec.Body.String(), `"delivery":"dev-response"`) {
		t.Fatalf("expected author invited, got status=%d body=%q", inviteRec.Code, inviteRec.Body.String())
	}

	var invitation struct {
		ResetToken string `json:"resetToken"`
		User       struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	}
	if err := json.Unmarshal(inviteRec.Body.Bytes(), &invitation); err != nil {
		t.Fatalf("decode invitation: %v", err)
	}
	if invitation.ResetToken == "" || invitation.User.Email != "writer@example.com" || invitation.User.Role != "author" {
		t.Fatalf("expected invitation token and author user, got %+v", invitation)
	}

	resetInvitedReq := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewBufferString(`{"token":"`+invitation.ResetToken+`","newPassword":"writer-password"}`))
	resetInvitedReq.Header.Set("Content-Type", "application/json")
	resetInvitedRec := httptest.NewRecorder()
	router.ServeHTTP(resetInvitedRec, resetInvitedReq)
	if resetInvitedRec.Code != http.StatusOK || !strings.Contains(resetInvitedRec.Body.String(), `"ok":true`) {
		t.Fatalf("expected invited author password reset, got status=%d body=%q", resetInvitedRec.Code, resetInvitedRec.Body.String())
	}

	writerLoginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"writer@example.com","password":"writer-password"}`))
	writerLoginReq.Header.Set("Content-Type", "application/json")
	writerLoginRec := httptest.NewRecorder()
	router.ServeHTTP(writerLoginRec, writerLoginReq)
	if writerLoginRec.Code != http.StatusOK || !strings.Contains(writerLoginRec.Body.String(), `"role":"author"`) {
		t.Fatalf("expected invited author login, got status=%d body=%q", writerLoginRec.Code, writerLoginRec.Body.String())
	}

	upgradeRoleReq := httptest.NewRequest(http.MethodPut, "/api/admin/users/user_linyi/role", bytes.NewBufferString(`{"role":"author"}`))
	upgradeRoleReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		upgradeRoleReq.AddCookie(cookie)
	}
	upgradeRoleRec := httptest.NewRecorder()
	router.ServeHTTP(upgradeRoleRec, upgradeRoleReq)
	if upgradeRoleRec.Code != http.StatusOK || !strings.Contains(upgradeRoleRec.Body.String(), `"role":"author"`) {
		t.Fatalf("expected existing user upgraded to author, got status=%d body=%q", upgradeRoleRec.Code, upgradeRoleRec.Body.String())
	}

	upgradedLoginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"linyi@example.com","password":"new-password"}`))
	upgradedLoginReq.Header.Set("Content-Type", "application/json")
	upgradedLoginRec := httptest.NewRecorder()
	router.ServeHTTP(upgradedLoginRec, upgradedLoginReq)
	if upgradedLoginRec.Code != http.StatusOK || !strings.Contains(upgradedLoginRec.Body.String(), `"role":"author"`) {
		t.Fatalf("expected upgraded user login role author, got status=%d body=%q", upgradedLoginRec.Code, upgradedLoginRec.Body.String())
	}

	usersExportReq := httptest.NewRequest(http.MethodGet, "/api/admin/users/export", nil)
	for _, cookie := range adminCookies {
		usersExportReq.AddCookie(cookie)
	}
	usersExportRec := httptest.NewRecorder()
	router.ServeHTTP(usersExportRec, usersExportReq)
	if usersExportRec.Code != http.StatusOK || !strings.Contains(usersExportRec.Body.String(), `"scope":"users"`) || !strings.Contains(usersExportRec.Body.String(), "market_user") {
		t.Fatalf("expected users export, got status=%d body=%q", usersExportRec.Code, usersExportRec.Body.String())
	}

	resetUserPasswordReq := httptest.NewRequest(http.MethodPost, "/api/admin/users/user_linyi/password-reset", nil)
	for _, cookie := range adminCookies {
		resetUserPasswordReq.AddCookie(cookie)
	}
	resetUserPasswordRec := httptest.NewRecorder()
	router.ServeHTTP(resetUserPasswordRec, resetUserPasswordReq)
	if resetUserPasswordRec.Code != http.StatusOK || !strings.Contains(resetUserPasswordRec.Body.String(), `"delivery":"dev-response"`) || !strings.Contains(resetUserPasswordRec.Body.String(), `"resetToken"`) {
		t.Fatalf("expected admin password reset token, got status=%d body=%q", resetUserPasswordRec.Code, resetUserPasswordRec.Body.String())
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

func TestEmailVerificationAndPasswordReset(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
	})

	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(`{"email":"verify@example.com","password":"secret123","displayName":"验证用户"}`))
	registerReq.Header.Set("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusCreated || !strings.Contains(registerRec.Body.String(), `"emailVerified":false`) {
		t.Fatalf("expected unverified registered user, got status=%d body=%q", registerRec.Code, registerRec.Body.String())
	}

	var registered struct {
		VerificationToken string `json:"verificationToken"`
	}
	if err := json.Unmarshal(registerRec.Body.Bytes(), &registered); err != nil {
		t.Fatalf("decode registered user: %v", err)
	}
	if registered.VerificationToken == "" {
		t.Fatalf("expected verification token in dev response")
	}

	verifyReq := httptest.NewRequest(http.MethodPost, "/api/auth/verify-email", bytes.NewBufferString(`{"token":"`+registered.VerificationToken+`"}`))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	router.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK || !strings.Contains(verifyRec.Body.String(), `"emailVerified":true`) {
		t.Fatalf("expected verified user, got status=%d body=%q", verifyRec.Code, verifyRec.Body.String())
	}

	forgotReq := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", bytes.NewBufferString(`{"email":"verify@example.com"}`))
	forgotReq.Header.Set("Content-Type", "application/json")
	forgotRec := httptest.NewRecorder()
	router.ServeHTTP(forgotRec, forgotReq)
	if forgotRec.Code != http.StatusOK {
		t.Fatalf("expected forgot password accepted, got status=%d body=%q", forgotRec.Code, forgotRec.Body.String())
	}

	var forgot struct {
		ResetToken string `json:"resetToken"`
	}
	if err := json.Unmarshal(forgotRec.Body.Bytes(), &forgot); err != nil {
		t.Fatalf("decode reset token: %v", err)
	}
	if forgot.ResetToken == "" {
		t.Fatalf("expected reset token in dev response")
	}

	resetReq := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewBufferString(`{"token":"`+forgot.ResetToken+`","newPassword":"reset123"}`))
	resetReq.Header.Set("Content-Type", "application/json")
	resetRec := httptest.NewRecorder()
	router.ServeHTTP(resetRec, resetReq)
	if resetRec.Code != http.StatusOK || !strings.Contains(resetRec.Body.String(), `"ok":true`) {
		t.Fatalf("expected password reset, got status=%d body=%q", resetRec.Code, resetRec.Body.String())
	}

	oldLoginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"verify@example.com","password":"secret123"}`))
	oldLoginReq.Header.Set("Content-Type", "application/json")
	oldLoginRec := httptest.NewRecorder()
	router.ServeHTTP(oldLoginRec, oldLoginReq)
	if oldLoginRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password rejected, got status=%d body=%q", oldLoginRec.Code, oldLoginRec.Body.String())
	}

	newLoginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"verify@example.com","password":"reset123"}`))
	newLoginReq.Header.Set("Content-Type", "application/json")
	newLoginRec := httptest.NewRecorder()
	router.ServeHTTP(newLoginRec, newLoginReq)
	if newLoginRec.Code != http.StatusOK || !strings.Contains(newLoginRec.Body.String(), `"emailVerified":true`) {
		t.Fatalf("expected new password login, got status=%d body=%q", newLoginRec.Code, newLoginRec.Body.String())
	}
}

func TestAccountSessionsExportAndDelete(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
	})

	firstCookies := loginForTest(t, router, "linyi@example.com", "password")
	secondCookies := loginForTest(t, router, "linyi@example.com", "password")
	currentToken := sessionCookieValue(t, secondCookies)

	sessionsReq := httptest.NewRequest(http.MethodGet, "/api/me/sessions", nil)
	for _, cookie := range secondCookies {
		sessionsReq.AddCookie(cookie)
	}
	sessionsRec := httptest.NewRecorder()
	router.ServeHTTP(sessionsRec, sessionsReq)
	if sessionsRec.Code != http.StatusOK || !strings.Contains(sessionsRec.Body.String(), `"current":true`) {
		t.Fatalf("expected session list, got status=%d body=%q", sessionsRec.Code, sessionsRec.Body.String())
	}

	var sessions struct {
		Items []struct {
			ID      string `json:"id"`
			Current bool   `json:"current"`
		} `json:"items"`
	}
	if err := json.Unmarshal(sessionsRec.Body.Bytes(), &sessions); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}
	if len(sessions.Items) < 2 {
		t.Fatalf("expected at least two sessions, got %+v", sessions)
	}

	var oldSessionID string
	for _, session := range sessions.Items {
		if !session.Current {
			oldSessionID = session.ID
			break
		}
	}
	if oldSessionID == "" || oldSessionID == currentToken {
		t.Fatalf("expected non-current session id, got %+v current=%q", sessions, currentToken)
	}

	deleteSessionReq := httptest.NewRequest(http.MethodDelete, "/api/me/sessions/"+oldSessionID, nil)
	for _, cookie := range secondCookies {
		deleteSessionReq.AddCookie(cookie)
	}
	deleteSessionRec := httptest.NewRecorder()
	router.ServeHTTP(deleteSessionRec, deleteSessionReq)
	if deleteSessionRec.Code != http.StatusOK || !strings.Contains(deleteSessionRec.Body.String(), `"ok":true`) {
		t.Fatalf("expected session deleted, got status=%d body=%q", deleteSessionRec.Code, deleteSessionRec.Body.String())
	}

	oldMeReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	for _, cookie := range firstCookies {
		oldMeReq.AddCookie(cookie)
	}
	oldMeRec := httptest.NewRecorder()
	router.ServeHTTP(oldMeRec, oldMeReq)
	if oldMeRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected removed session unauthorized, got status=%d body=%q", oldMeRec.Code, oldMeRec.Body.String())
	}

	exportReq := httptest.NewRequest(http.MethodPost, "/api/me/export", nil)
	for _, cookie := range secondCookies {
		exportReq.AddCookie(cookie)
	}
	exportRec := httptest.NewRecorder()
	router.ServeHTTP(exportRec, exportReq)
	if exportRec.Code != http.StatusOK || !strings.Contains(exportRec.Body.String(), `"email":"linyi@example.com"`) || !strings.Contains(exportRec.Body.String(), `"sessions"`) {
		t.Fatalf("expected account export, got status=%d body=%q", exportRec.Code, exportRec.Body.String())
	}

	deleteAccountReq := httptest.NewRequest(http.MethodDelete, "/api/me", nil)
	for _, cookie := range secondCookies {
		deleteAccountReq.AddCookie(cookie)
	}
	deleteAccountRec := httptest.NewRecorder()
	router.ServeHTTP(deleteAccountRec, deleteAccountReq)
	if deleteAccountRec.Code != http.StatusOK || !strings.Contains(deleteAccountRec.Body.String(), `"ok":true`) {
		t.Fatalf("expected account deleted, got status=%d body=%q", deleteAccountRec.Code, deleteAccountRec.Body.String())
	}

	deletedMeReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	for _, cookie := range secondCookies {
		deletedMeReq.AddCookie(cookie)
	}
	deletedMeRec := httptest.NewRecorder()
	router.ServeHTTP(deletedMeRec, deletedMeReq)
	if deletedMeRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected deleted account session invalid, got status=%d body=%q", deletedMeRec.Code, deletedMeRec.Body.String())
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"linyi@example.com","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected deleted account login rejected, got status=%d body=%q", loginRec.Code, loginRec.Body.String())
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
		ID      string `json:"id"`
		Version int    `json:"version"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode admin post: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected admin post id, got %q", createRec.Body.String())
	}
	if created.Version != 1 {
		t.Fatalf("expected initial version 1, got %+v", created)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/admin/posts/"+created.ID, bytes.NewBufferString(`{
		"title":"后台发布流程验证第二版",
		"summary":"验证管理员保存草稿、版本历史和发布到前台。",
		"content":"这是第二版内容，用于验证版本历史可以记录每一次保存。",
		"status":"draft",
		"category":"工程实践",
		"tags":["后台","发布","版本"],
		"slug":"admin-publish-flow-check",
		"coverImage":"https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=1200&q=80",
		"seoTitle":"后台发布流程验证第二版",
		"seoDescription":"验证管理员保存草稿、版本历史和发布到前台。"
	}`))
	updateReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		updateReq.AddCookie(cookie)
	}
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected admin post updated, got status=%d body=%q", updateRec.Code, updateRec.Body.String())
	}

	var updated struct {
		Version int    `json:"version"`
		Title   string `json:"title"`
	}
	if err := json.Unmarshal(updateRec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode updated admin post: %v", err)
	}
	if updated.Version != 2 || updated.Title != "后台发布流程验证第二版" {
		t.Fatalf("expected version 2 update, got %+v", updated)
	}

	revisionsReq := httptest.NewRequest(http.MethodGet, "/api/admin/posts/"+created.ID+"/revisions", nil)
	for _, cookie := range adminCookies {
		revisionsReq.AddCookie(cookie)
	}
	revisionsRec := httptest.NewRecorder()
	router.ServeHTTP(revisionsRec, revisionsReq)
	if revisionsRec.Code != http.StatusOK {
		t.Fatalf("expected admin post revisions, got status=%d body=%q", revisionsRec.Code, revisionsRec.Body.String())
	}

	var revisions struct {
		Items []struct {
			ID      string `json:"id"`
			Version int    `json:"version"`
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(revisionsRec.Body.Bytes(), &revisions); err != nil {
		t.Fatalf("decode admin post revisions: %v", err)
	}
	if revisions.Total != 2 || len(revisions.Items) != 2 {
		t.Fatalf("expected two revisions, got %+v", revisions)
	}

	var firstRevisionID string
	for _, revision := range revisions.Items {
		if revision.Version == 1 {
			firstRevisionID = revision.ID
			if revision.Title != "后台发布流程验证" || !strings.Contains(revision.Content, "发布动作应该调用公开文章发布能力") {
				t.Fatalf("expected first revision snapshot, got %+v", revision)
			}
		}
	}
	if firstRevisionID == "" {
		t.Fatalf("expected first revision id, got %+v", revisions)
	}

	restoreReq := httptest.NewRequest(http.MethodPost, "/api/admin/posts/"+created.ID+"/revisions/"+firstRevisionID+"/restore", nil)
	for _, cookie := range adminCookies {
		restoreReq.AddCookie(cookie)
	}
	restoreRec := httptest.NewRecorder()
	router.ServeHTTP(restoreRec, restoreReq)
	if restoreRec.Code != http.StatusOK {
		t.Fatalf("expected revision restored, got status=%d body=%q", restoreRec.Code, restoreRec.Body.String())
	}

	var restored struct {
		Version int    `json:"version"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(restoreRec.Body.Bytes(), &restored); err != nil {
		t.Fatalf("decode restored admin post: %v", err)
	}
	if restored.Version != 3 || restored.Title != "后台发布流程验证" || !strings.Contains(restored.Content, "发布动作应该调用公开文章发布能力") {
		t.Fatalf("expected restored first revision as version 3, got %+v", restored)
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

func TestTaxonomyAPIs(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:    "test",
		HTTPAddr:  ":0",
		WebOrigin: "http://localhost:5173",
	})

	categoriesReq := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	categoriesRec := httptest.NewRecorder()
	router.ServeHTTP(categoriesRec, categoriesReq)
	if categoriesRec.Code != http.StatusOK || !strings.Contains(categoriesRec.Body.String(), `"name":"工程实践"`) || !strings.Contains(categoriesRec.Body.String(), `"postCount":3`) {
		t.Fatalf("expected public categories, got status=%d body=%q", categoriesRec.Code, categoriesRec.Body.String())
	}

	tagsReq := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	tagsRec := httptest.NewRecorder()
	router.ServeHTTP(tagsRec, tagsReq)
	if tagsRec.Code != http.StatusOK || !strings.Contains(tagsRec.Body.String(), `"name":"博客系统"`) {
		t.Fatalf("expected public tags, got status=%d body=%q", tagsRec.Code, tagsRec.Body.String())
	}

	anonCreateReq := httptest.NewRequest(http.MethodPost, "/api/admin/categories", bytes.NewBufferString(`{"name":"读书笔记","slug":"reading-notes"}`))
	anonCreateReq.Header.Set("Content-Type", "application/json")
	anonCreateRec := httptest.NewRecorder()
	router.ServeHTTP(anonCreateRec, anonCreateReq)
	if anonCreateRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected anonymous category create status 401, got %d", anonCreateRec.Code)
	}

	adminCookies := loginForTest(t, router, "admin@example.com", "password")

	createCategoryReq := httptest.NewRequest(http.MethodPost, "/api/admin/categories", bytes.NewBufferString(`{
		"name":"读书笔记",
		"slug":"reading-notes",
		"description":"书评、阅读记录和资料整理。",
		"sortOrder":70
	}`))
	createCategoryReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		createCategoryReq.AddCookie(cookie)
	}
	createCategoryRec := httptest.NewRecorder()
	router.ServeHTTP(createCategoryRec, createCategoryReq)
	if createCategoryRec.Code != http.StatusCreated || !strings.Contains(createCategoryRec.Body.String(), `"slug":"reading-notes"`) {
		t.Fatalf("expected category created, got status=%d body=%q", createCategoryRec.Code, createCategoryRec.Body.String())
	}

	var createdCategory struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createCategoryRec.Body.Bytes(), &createdCategory); err != nil {
		t.Fatalf("decode category: %v", err)
	}

	updateCategoryReq := httptest.NewRequest(http.MethodPut, "/api/admin/categories/"+createdCategory.ID, bytes.NewBufferString(`{
		"name":"阅读笔记",
		"slug":"reading",
		"description":"阅读和资料整理。",
		"sortOrder":75
	}`))
	updateCategoryReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		updateCategoryReq.AddCookie(cookie)
	}
	updateCategoryRec := httptest.NewRecorder()
	router.ServeHTTP(updateCategoryRec, updateCategoryReq)
	if updateCategoryRec.Code != http.StatusOK || !strings.Contains(updateCategoryRec.Body.String(), `"slug":"reading"`) {
		t.Fatalf("expected category updated, got status=%d body=%q", updateCategoryRec.Code, updateCategoryRec.Body.String())
	}

	duplicateCategoryReq := httptest.NewRequest(http.MethodPost, "/api/admin/categories", bytes.NewBufferString(`{"name":"工程实践","slug":"engineering"}`))
	duplicateCategoryReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		duplicateCategoryReq.AddCookie(cookie)
	}
	duplicateCategoryRec := httptest.NewRecorder()
	router.ServeHTTP(duplicateCategoryRec, duplicateCategoryReq)
	if duplicateCategoryRec.Code != http.StatusConflict {
		t.Fatalf("expected duplicate category status 409, got %d body=%q", duplicateCategoryRec.Code, duplicateCategoryRec.Body.String())
	}

	deleteUsedCategoryReq := httptest.NewRequest(http.MethodDelete, "/api/admin/categories/category_engineering", nil)
	for _, cookie := range adminCookies {
		deleteUsedCategoryReq.AddCookie(cookie)
	}
	deleteUsedCategoryRec := httptest.NewRecorder()
	router.ServeHTTP(deleteUsedCategoryRec, deleteUsedCategoryReq)
	if deleteUsedCategoryRec.Code != http.StatusConflict {
		t.Fatalf("expected used category delete status 409, got %d body=%q", deleteUsedCategoryRec.Code, deleteUsedCategoryRec.Body.String())
	}

	deleteCategoryReq := httptest.NewRequest(http.MethodDelete, "/api/admin/categories/"+createdCategory.ID, nil)
	for _, cookie := range adminCookies {
		deleteCategoryReq.AddCookie(cookie)
	}
	deleteCategoryRec := httptest.NewRecorder()
	router.ServeHTTP(deleteCategoryRec, deleteCategoryReq)
	if deleteCategoryRec.Code != http.StatusOK || !strings.Contains(deleteCategoryRec.Body.String(), `"ok":true`) {
		t.Fatalf("expected category deleted, got status=%d body=%q", deleteCategoryRec.Code, deleteCategoryRec.Body.String())
	}

	createTagReq := httptest.NewRequest(http.MethodPost, "/api/admin/tags", bytes.NewBufferString(`{"name":"数据库","slug":"database"}`))
	createTagReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		createTagReq.AddCookie(cookie)
	}
	createTagRec := httptest.NewRecorder()
	router.ServeHTTP(createTagRec, createTagReq)
	if createTagRec.Code != http.StatusCreated || !strings.Contains(createTagRec.Body.String(), `"slug":"database"`) {
		t.Fatalf("expected tag created, got status=%d body=%q", createTagRec.Code, createTagRec.Body.String())
	}

	var createdTag struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createTagRec.Body.Bytes(), &createdTag); err != nil {
		t.Fatalf("decode tag: %v", err)
	}

	updateTagReq := httptest.NewRequest(http.MethodPut, "/api/admin/tags/"+createdTag.ID, bytes.NewBufferString(`{"name":"数据库实践","slug":"database-practice"}`))
	updateTagReq.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminCookies {
		updateTagReq.AddCookie(cookie)
	}
	updateTagRec := httptest.NewRecorder()
	router.ServeHTTP(updateTagRec, updateTagReq)
	if updateTagRec.Code != http.StatusOK || !strings.Contains(updateTagRec.Body.String(), `"slug":"database-practice"`) {
		t.Fatalf("expected tag updated, got status=%d body=%q", updateTagRec.Code, updateTagRec.Body.String())
	}

	deleteUsedTagReq := httptest.NewRequest(http.MethodDelete, "/api/admin/tags/tag_blog_system", nil)
	for _, cookie := range adminCookies {
		deleteUsedTagReq.AddCookie(cookie)
	}
	deleteUsedTagRec := httptest.NewRecorder()
	router.ServeHTTP(deleteUsedTagRec, deleteUsedTagReq)
	if deleteUsedTagRec.Code != http.StatusConflict {
		t.Fatalf("expected used tag delete status 409, got %d body=%q", deleteUsedTagRec.Code, deleteUsedTagRec.Body.String())
	}

	deleteTagReq := httptest.NewRequest(http.MethodDelete, "/api/admin/tags/"+createdTag.ID, nil)
	for _, cookie := range adminCookies {
		deleteTagReq.AddCookie(cookie)
	}
	deleteTagRec := httptest.NewRecorder()
	router.ServeHTTP(deleteTagRec, deleteTagReq)
	if deleteTagRec.Code != http.StatusOK || !strings.Contains(deleteTagRec.Body.String(), `"ok":true`) {
		t.Fatalf("expected tag deleted, got status=%d body=%q", deleteTagRec.Code, deleteTagRec.Body.String())
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

func sessionCookieValue(t *testing.T, cookies []*http.Cookie) string {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == "blog_session" {
			return cookie.Value
		}
	}

	t.Fatalf("expected session cookie, got %+v", cookies)
	return ""
}

func tinyPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
}
