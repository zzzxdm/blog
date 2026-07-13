package operations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"blog/api/internal/idgen"
)

var (
	ErrMediaNotFound = errors.New("media asset not found")
	ErrMediaInUse    = errors.New("media asset is in use")
)

type Repository interface {
	GetSettings(ctx context.Context) (Settings, error)
	UpdateSettings(ctx context.Context, settings Settings) (Settings, error)
	SendTestMail(ctx context.Context) (TestMailResult, error)
	RunBackup(ctx context.Context) (BackupResult, error)
	GetNavigation(ctx context.Context) (Navigation, error)
	UpdateNavigation(ctx context.Context, navigation Navigation) (Navigation, error)
	ListMedia(ctx context.Context, query MediaListQuery) (MediaListResult, error)
	ListMediaReferences(ctx context.Context, id string, page int, pageSize int) (MediaReferenceListResult, error)
	CreateMedia(ctx context.Context, asset MediaAsset) (MediaAsset, error)
	GetMedia(ctx context.Context, id string) (MediaAsset, error)
	GetMediaFile(ctx context.Context, id string) (MediaAsset, error)
	UpdateMedia(ctx context.Context, id string, request MediaUpdateRequest) (MediaAsset, error)
	UpdateMediaFile(ctx context.Context, id string, request MediaFileUpdateRequest) (MediaAsset, error)
	DeleteMedia(ctx context.Context, id string) (MediaAsset, error)
	GetStats(ctx context.Context, rangeKey string) (Stats, error)
	ListAuditLogs(ctx context.Context, query AuditLogQuery) (AuditLogListResult, error)
	RecordAuditLog(ctx context.Context, item AuditLog) error
}

func cloneSettings(settings Settings) Settings {
	settings.BlockedWords = append([]string{}, settings.BlockedWords...)
	return settings
}

func settingsForUpdate(settings Settings, now time.Time) Settings {
	settings = normalizeSettings(settings)
	settings.UpdatedAt = now
	return settings
}

func normalizeSettings(settings Settings) Settings {
	defaults := seedSettings()

	settings.SiteName = defaultString(settings.SiteName, defaults.SiteName)
	settings.SiteDescription = strings.TrimSpace(settings.SiteDescription)
	settings.SiteURL = defaultString(settings.SiteURL, defaults.SiteURL)
	settings.Beian = strings.TrimSpace(settings.Beian)
	settings.ThemePrimary = normalizeThemeColor(settings.ThemePrimary, defaults.ThemePrimary)
	settings.HomepageLayout = defaultString(settings.HomepageLayout, defaults.HomepageLayout)
	settings.BlockedWords = normalizeBlockedWords(settings.BlockedWords)
	settings.LoginRequiredForComment = true
	settings.SubmissionManualReview = true
	settings.SubmissionLimit = defaultString(settings.SubmissionLimit, defaults.SubmissionLimit)
	settings.SubmissionGuide = strings.TrimSpace(settings.SubmissionGuide)
	settings.MailProvider = defaultString(settings.MailProvider, defaults.MailProvider)
	settings.FromEmail = defaultString(settings.FromEmail, defaults.FromEmail)
	settings.TurnstileSiteKey = strings.TrimSpace(settings.TurnstileSiteKey)
	settings.TurnstileSecretKey = strings.TrimSpace(settings.TurnstileSecretKey)
	settings.SessionDays = clampInt(settings.SessionDays, 1, 90, defaults.SessionDays)
	settings.BackupCycle = defaultString(settings.BackupCycle, defaults.BackupCycle)
	settings.BackupRetentionDays = clampInt(settings.BackupRetentionDays, 1, 365, defaults.BackupRetentionDays)
	if settings.LastBackupAt.IsZero() {
		settings.LastBackupAt = defaults.LastBackupAt
	}
	if settings.UpdatedAt.IsZero() {
		settings.UpdatedAt = defaults.UpdatedAt
	}

	return cloneSettings(settings)
}

func normalizeBlockedWords(words []string) []string {
	result := make([]string, 0, len(words))
	seen := map[string]bool{}
	for _, item := range words {
		value := strings.TrimSpace(item)
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}

	return result
}

func normalizeThemeColor(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if len(value) != 7 || !strings.HasPrefix(value, "#") {
		return fallback
	}
	for _, item := range value[1:] {
		isDigit := item >= '0' && item <= '9'
		isLowerHex := item >= 'a' && item <= 'f'
		isUpperHex := item >= 'A' && item <= 'F'
		if !isDigit && !isLowerHex && !isUpperHex {
			return fallback
		}
	}

	return strings.ToLower(value)
}

func clampInt(value int, minValue int, maxValue int, fallback int) int {
	if value < minValue {
		return fallback
	}
	if value > maxValue {
		return maxValue
	}

	return value
}

func publicSettings(settings Settings) PublicSettings {
	settings = normalizeSettings(settings)
	return PublicSettings{
		SiteName:                settings.SiteName,
		SiteDescription:         settings.SiteDescription,
		SiteURL:                 settings.SiteURL,
		Beian:                   settings.Beian,
		ThemePrimary:            settings.ThemePrimary,
		HomepageLayout:          settings.HomepageLayout,
		DarkModeEnabled:         settings.DarkModeEnabled,
		ReadingProgressEnabled:  settings.ReadingProgressEnabled,
		CommentsEnabled:         settings.CommentsEnabled,
		LoginRequiredForComment: settings.LoginRequiredForComment,
		SubmissionsEnabled:      settings.SubmissionsEnabled,
		SubmissionLimit:         settings.SubmissionLimit,
		SubmissionGuide:         settings.SubmissionGuide,
		TurnstileEnabled:        settings.TurnstileEnabled,
		TurnstileSiteKey:        settings.TurnstileSiteKey,
		TurnstileRegister:       settings.TurnstileRegister,
		TurnstileLogin:          settings.TurnstileLogin,
		TurnstileSubmission:     settings.TurnstileSubmission,
		UpdatedAt:               settings.UpdatedAt,
	}
}

func cloneNavigation(navigation Navigation) Navigation {
	navigation.TopItems = append([]NavItem{}, navigation.TopItems...)
	navigation.FooterItems = append([]NavItem{}, navigation.FooterItems...)
	navigation.Redirects = append([]RedirectRule{}, navigation.Redirects...)
	return navigation
}

func navigationForUpdate(navigation Navigation, now time.Time) Navigation {
	navigation = normalizeNavigation(navigation)
	navigation.UpdatedAt = now
	return navigation
}

func normalizeNavigation(navigation Navigation) Navigation {
	defaults := seedNavigation()
	navigation.TopItems = normalizeNavItems(navigation.TopItems, defaults.TopItems, "nav_top")
	navigation.FooterItems = normalizeNavItems(navigation.FooterItems, defaults.FooterItems, "nav_footer")
	navigation.GitHubURL = strings.TrimSpace(navigation.GitHubURL)
	navigation.ContactEmail = strings.TrimSpace(navigation.ContactEmail)
	navigation.RSSURL = strings.TrimSpace(navigation.RSSURL)
	navigation.Redirects = normalizeRedirects(navigation.Redirects)
	if navigation.UpdatedAt.IsZero() {
		navigation.UpdatedAt = defaults.UpdatedAt
	}

	return cloneNavigation(navigation)
}

func normalizeNavItems(items []NavItem, fallback []NavItem, prefix string) []NavItem {
	result := make([]NavItem, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		url := strings.TrimSpace(item.URL)
		if label == "" || url == "" {
			continue
		}

		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = fmt.Sprintf("%s_%d", prefix, len(result)+1)
		}
		result = append(result, NavItem{
			ID:    id,
			Label: label,
			URL:   url,
			Order: len(result) + 1,
		})
	}
	if len(result) == 0 {
		return append([]NavItem{}, fallback...)
	}

	return result
}

func normalizeRedirects(rules []RedirectRule) []RedirectRule {
	result := make([]RedirectRule, 0, len(rules))
	for _, rule := range rules {
		from := strings.TrimSpace(rule.From)
		to := strings.TrimSpace(rule.To)
		if from == "" || to == "" || from == to {
			continue
		}
		code := rule.Code
		if code != http.StatusMovedPermanently &&
			code != http.StatusFound &&
			code != http.StatusTemporaryRedirect &&
			code != http.StatusPermanentRedirect {
			code = http.StatusMovedPermanently
		}

		result = append(result, RedirectRule{
			From: from,
			To:   to,
			Code: code,
		})
	}

	return result
}

func cloneStats(stats Stats) Stats {
	stats.Metrics = append([]Metric{}, stats.Metrics...)
	stats.Trend = append([]BarPoint{}, stats.Trend...)
	stats.TopPosts = append([]TopPost{}, stats.TopPosts...)
	stats.Sources = append([]BarPoint{}, stats.Sources...)
	stats.SearchTerms = append([]SearchTerm{}, stats.SearchTerms...)
	stats.Suggestions = append([]ContentSuggestion{}, stats.Suggestions...)
	return stats
}

func cloneAuditLog(item AuditLog) AuditLog {
	return item
}

func MediaAssetURL(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}

	return "/api/media/" + id + "/file"
}

func MediaAssetReferenceTokens(id string) []string {
	publicURL := MediaAssetURL(id)
	if publicURL == "" {
		return []string{}
	}

	return []string{publicURL}
}

func pagedMediaResult(items []MediaAsset, query MediaListQuery) MediaListResult {
	total := len(items)
	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	paged := items
	if query.All {
		page = 1
		pageSize = total
	} else {
		start := (page - 1) * pageSize
		if start > total {
			start = total
		}
		end := start + pageSize
		if end > total {
			end = total
		}
		paged = items[start:end]
	}

	return MediaListResult{
		Items:    paged,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
}

func filterMedia(items []MediaAsset, query MediaListQuery) []MediaAsset {
	keyword := strings.ToLower(strings.TrimSpace(query.Keyword))
	mediaType := strings.ToLower(strings.TrimSpace(query.Type))
	filtered := make([]MediaAsset, 0, len(items))

	for _, item := range items {
		if mediaType != "" && mediaType != "all" && item.Type != mediaType {
			continue
		}
		if keyword != "" && !mediaContainsKeyword(item, keyword) {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered
}

func mediaContainsKeyword(item MediaAsset, keyword string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		item.FileName,
		item.Alt,
		item.UploadedBy,
		item.Category,
		item.URL,
		item.Type,
	}, " "))
	return strings.Contains(haystack, keyword)
}

func sortMedia(items []MediaAsset, mode string) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "size":
		sort.SliceStable(items, func(i, j int) bool {
			return mediaSizeValue(items[i].SizeLabel) > mediaSizeValue(items[j].SizeLabel)
		})
	case "usage":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].UsageCount > items[j].UsageCount
		})
	default:
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].UploadedAt.After(items[j].UploadedAt)
		})
	}
}

func mediaSizeValue(label string) float64 {
	value := firstFloat(label)
	upper := strings.ToUpper(label)
	switch {
	case strings.Contains(upper, "MB"):
		return value * 1024 * 1024
	case strings.Contains(upper, "KB"):
		return value * 1024
	default:
		return value
	}
}

func firstFloat(value string) float64 {
	number := strings.Builder{}
	seenDigit := false
	for _, item := range value {
		if item >= '0' && item <= '9' || item == '.' {
			number.WriteRune(item)
			if item != '.' {
				seenDigit = true
			}
			continue
		}
		if seenDigit {
			break
		}
	}
	if !seenDigit {
		return 0
	}
	parsed, err := strconv.ParseFloat(number.String(), 64)
	if err != nil {
		return 0
	}
	return parsed
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return 12
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

func normalizeAuditLogQuery(query AuditLogQuery) AuditLogQuery {
	query.Action = strings.TrimSpace(query.Action)
	query.ResourceType = strings.TrimSpace(query.ResourceType)
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	return query
}

func normalizeAuditLog(item AuditLog, now time.Time) AuditLog {
	if item.ID == "" {
		item.ID = idgen.NextString()
	}
	item.ActorID = strings.TrimSpace(item.ActorID)
	item.ActorName = defaultString(item.ActorName, "匿名用户")
	item.Action = defaultString(item.Action, "admin.write")
	item.ResourceType = defaultString(item.ResourceType, "admin")
	item.ResourceID = strings.TrimSpace(item.ResourceID)
	item.ResourceTitle = strings.TrimSpace(item.ResourceTitle)
	item.Status = defaultString(item.Status, "success")
	item.IP = strings.TrimSpace(item.IP)
	item.UserAgent = strings.TrimSpace(item.UserAgent)
	item.Detail = strings.TrimSpace(item.Detail)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}

	return item
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}

	return value
}

func testMailResult(settings Settings, now time.Time) TestMailResult {
	provider := defaultString(settings.MailProvider, "SMTP")
	fromEmail := defaultString(settings.FromEmail, "noreply@example.com")
	return TestMailResult{
		OK:        true,
		Provider:  provider,
		FromEmail: fromEmail,
		Delivery:  "dev-response",
		Message:   fmt.Sprintf("测试邮件已生成：%s -> %s", provider, fromEmail),
		TestedAt:  now,
	}
}

func backupResult(settings Settings, now time.Time) BackupResult {
	return BackupResult{
		OK:        true,
		ID:        fmt.Sprintf("backup_%d", now.UnixNano()),
		Status:    "completed",
		FileName:  fmt.Sprintf("blog-backup-%s.zip", now.Format("20060102-150405")),
		SizeLabel: "4.8 MB",
		Message:   "备份已完成，生产环境可接入对象存储归档。",
		CreatedAt: now,
		Settings:  cloneSettings(settings),
	}
}

func seedSettings() Settings {
	now := time.Now()
	return Settings{
		SiteName:                "云间笔记",
		SiteDescription:         "技术、产品、工程实践和长期写作的沉淀。",
		SiteURL:                 "https://blog.example.com",
		Beian:                   "京ICP备00000000号",
		ThemePrimary:            "#295b4b",
		HomepageLayout:          "精选文章 + 最新列表",
		DarkModeEnabled:         true,
		ReadingProgressEnabled:  true,
		CommentsEnabled:         true,
		LoginRequiredForComment: true,
		AutoApproveComments:     false,
		BlockedWords:            []string{"推广", "返利", "博彩"},
		SubmissionsEnabled:      true,
		SubmissionManualReview:  true,
		SubmissionLimit:         "每天最多 3 篇",
		SubmissionGuide:         "投稿需要原创、结构清晰，并补充必要的图片 alt 文本和参考来源。",
		MailEnabled:             false,
		MailProvider:            "Resend",
		FromEmail:               "noreply@example.com",
		TurnstileEnabled:        false,
		TurnstileSiteKey:        "",
		TurnstileSecretKey:      "",
		TurnstileRegister:       false,
		TurnstileLogin:          false,
		TurnstileSubmission:     false,
		AdminTwoFactorRequired:  true,
		LoginFailureLock:        true,
		SessionDays:             7,
		BackupCycle:             "每日全量备份",
		LastBackupAt:            time.Date(2026, 7, 4, 3, 0, 0, 0, time.Local),
		BackupRetentionDays:     7,
		UpdatedAt:               now,
	}
}

func seedNavigation() Navigation {
	return Navigation{
		TopItems: []NavItem{
			{ID: "nav_top_1", Label: "首页", URL: "/", Order: 1},
			{ID: "nav_top_2", Label: "归档", URL: "/archive", Order: 2},
			{ID: "nav_top_3", Label: "专题", URL: "/topics", Order: 3},
			{ID: "nav_top_4", Label: "投稿", URL: "/submit", Order: 4},
		},
		FooterItems:            []NavItem{},
		MobileCollapse:         true,
		ExternalLinksNewWindow: true,
		ShowLoginEntry:         true,
		GitHubURL:              "https://github.com/zzzxdm/blog",
		ContactEmail:           "admin@jecyai.com",
		RSSURL:                 "",
		Redirects: []RedirectRule{
			{From: "/old-blog-design", To: "/posts/blog-system-design", Code: 301},
		},
		UpdatedAt: time.Now(),
	}
}

func seedMedia() []MediaAsset {
	now := time.Now()
	return []MediaAsset{
		{ID: "9001", FileName: "cover-code-desk.jpg", URL: "https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=700&q=80", Alt: "代码编辑器封面图", Type: "image", Category: "封面图", SizeLabel: "1.2 MB", Width: 1400, Height: 930, UsageCount: 8, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -1)},
		{ID: "9002", FileName: "server-room.jpg", URL: "https://images.unsplash.com/photo-1558494949-ef010cbdcc31?auto=format&fit=crop&w=700&q=80", Alt: "服务器设备图片", Type: "image", Category: "架构", SizeLabel: "980 KB", Width: 1400, Height: 930, UsageCount: 3, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -2)},
		{ID: "9003", FileName: "writing-notes.jpg", URL: "https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=700&q=80", Alt: "写作桌面图片", Type: "image", Category: "写作", SizeLabel: "860 KB", Width: 1400, Height: 930, UsageCount: 5, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -3)},
		{ID: "9004", FileName: "product-dashboard.jpg", URL: "https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=700&q=80", Alt: "产品分析界面", Type: "image", Category: "产品设计", SizeLabel: "1.4 MB", Width: 1400, Height: 930, UsageCount: 2, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -4)},
		{ID: "9005", FileName: "javascript-editor.jpg", URL: "https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=700&q=80", Alt: "代码窗口", Type: "image", Category: "Vue3", SizeLabel: "740 KB", Width: 1400, Height: 930, UsageCount: 4, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -5)},
		{ID: "9006", FileName: "quiet-workspace.jpg", URL: "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=700&q=80", Alt: "自然光工作区", Type: "image", Category: "备用", SizeLabel: "690 KB", Width: 1400, Height: 930, UsageCount: 0, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -6)},
	}
}

func seedStats() Stats {
	return Stats{
		Range:      "30d",
		RangeLabel: "最近 30 天",
		Metrics: []Metric{
			{Label: "阅读量", Value: "86,400", Delta: "当前范围累计"},
			{Label: "发布文章", Value: "18", Delta: "已公开"},
			{Label: "平均阅读", Value: "4.3 分钟", Delta: "按阅读时长估算"},
			{Label: "互动数", Value: "1,246", Delta: "评论、点赞和收藏"},
		},
		Trend: []BarPoint{
			{Label: "周一", Value: "8.4k", Percent: 68},
			{Label: "周二", Value: "9.1k", Percent: 74},
			{Label: "周三", Value: "11.3k", Percent: 92, Tone: "rust"},
			{Label: "周四", Value: "9.9k", Percent: 81},
			{Label: "周五", Value: "7.8k", Percent: 64, Tone: "amber"},
			{Label: "周六", Value: "5.9k", Percent: 48},
		},
		TopPosts: []TopPost{
			{Title: "Vue3 内容站的缓存与 SEO 边界", Views: "12,420", Bookmarks: 312, Comments: 48, EngagementRate: "3.8%"},
			{Title: "如何设计一个内容长期增长的博客系统", Views: "9,884", Bookmarks: 286, Comments: 34, EngagementRate: "4.1%"},
			{Title: "让旧文章继续被搜索引擎找到", Views: "7,209", Bookmarks: 190, Comments: 19, EngagementRate: "2.4%"},
		},
		Sources: []BarPoint{
			{Label: "后台发布", Value: "72%", Percent: 72},
			{Label: "用户投稿", Value: "28%", Percent: 28, Tone: "rust"},
		},
		SearchTerms: []SearchTerm{
			{Term: "Vue3", Count: 8},
			{Term: "博客系统", Count: 6},
			{Term: "Markdown", Count: 4},
		},
		Suggestions: []ContentSuggestion{
			{Title: "评论审核相关文章增长明显", Body: "可以补一篇用户评论系统设计。"},
			{Title: "专题页带来较多收藏", Body: "建议继续完善专题导航。"},
		},
	}
}

func statsForRange(stats Stats, rangeKey string) Stats {
	switch normalizeStatsRange(rangeKey) {
	case "7d":
		stats.Range = "7d"
		stats.RangeLabel = "最近 7 天"
		stats.Metrics = []Metric{
			{Label: "阅读量", Value: "18,200", Delta: "当前范围累计"},
			{Label: "发布文章", Value: "5", Delta: "已公开"},
			{Label: "平均阅读", Value: "4.0 分钟", Delta: "按阅读时长估算"},
			{Label: "互动数", Value: "318", Delta: "评论、点赞和收藏"},
		}
		stats.Trend = []BarPoint{
			{Label: "周一", Value: "2.1k", Percent: 58},
			{Label: "周二", Value: "2.4k", Percent: 66},
			{Label: "周三", Value: "3.2k", Percent: 88, Tone: "rust"},
			{Label: "周四", Value: "2.8k", Percent: 77},
			{Label: "周五", Value: "2.5k", Percent: 69, Tone: "amber"},
			{Label: "周六", Value: "1.9k", Percent: 52},
			{Label: "周日", Value: "1.3k", Percent: 36},
		}
		stats.SearchTerms = []SearchTerm{
			{Term: "站内信", Count: 3},
			{Term: "投稿审核", Count: 2},
			{Term: "Vue3", Count: 2},
		}
		stats.Suggestions = []ContentSuggestion{
			{Title: "投稿审核相关内容增加", Body: "可以补充一篇投稿流程说明。"},
			{Title: "站内信功能关注度提升", Body: "建议完善账号通知文档。"},
		}
	case "ytd":
		stats.Range = "ytd"
		stats.RangeLabel = "今年"
		stats.Metrics = []Metric{
			{Label: "阅读量", Value: "642,800", Delta: "当前范围累计"},
			{Label: "发布文章", Value: "96", Delta: "已公开"},
			{Label: "平均阅读", Value: "4.6 分钟", Delta: "按阅读时长估算"},
			{Label: "互动数", Value: "9,884", Delta: "评论、点赞和收藏"},
		}
		stats.Trend = []BarPoint{
			{Label: "1月", Value: "62k", Percent: 54},
			{Label: "2月", Value: "58k", Percent: 50},
			{Label: "3月", Value: "79k", Percent: 69, Tone: "amber"},
			{Label: "4月", Value: "93k", Percent: 81},
			{Label: "5月", Value: "112k", Percent: 97, Tone: "rust"},
			{Label: "6月", Value: "104k", Percent: 90},
			{Label: "7月", Value: "86k", Percent: 75},
		}
		stats.TopPosts = []TopPost{
			{Title: "如何设计一个内容长期增长的博客系统", Views: "48,210", Bookmarks: 1204, Comments: 184, EngagementRate: "4.6%"},
			{Title: "Vue3 内容站的缓存与 SEO 边界", Views: "41,884", Bookmarks: 986, Comments: 128, EngagementRate: "3.9%"},
			{Title: "Redis 和 PostgreSQL 在博客中的分工", Views: "33,406", Bookmarks: 712, Comments: 93, EngagementRate: "3.1%"},
		}
		stats.SearchTerms = []SearchTerm{
			{Term: "博客系统", Count: 24},
			{Term: "Vue3", Count: 21},
			{Term: "PostgreSQL", Count: 12},
		}
		stats.Suggestions = []ContentSuggestion{
			{Title: "架构类文章全年表现稳定", Body: "可以将数据库、缓存、搜索整理成系列专题。"},
			{Title: "Vue3 内容长期被收藏", Body: "建议补充 SSR 和预渲染边界文章。"},
		}
	default:
		stats.Range = "30d"
		stats.RangeLabel = "最近 30 天"
	}

	return stats
}

func normalizeStatsRange(rangeKey string) string {
	switch strings.ToLower(strings.TrimSpace(rangeKey)) {
	case "7d", "7":
		return "7d"
	case "ytd", "year":
		return "ytd"
	default:
		return "30d"
	}
}

func seedAuditLogs() []AuditLog {
	now := time.Now()
	return []AuditLog{
		{
			ID:            "10001",
			ActorID:       "5002",
			ActorName:     "管理员",
			Action:        "post.publish",
			ResourceType:  "post",
			ResourceID:    "11001",
			ResourceTitle: "如何设计一个内容长期增长的博客系统",
			Status:        "success",
			IP:            "127.0.0.1",
			UserAgent:     "seed",
			Detail:        "发布文章到前台",
			CreatedAt:     now.Add(-2 * time.Hour),
		},
		{
			ID:            "10002",
			ActorID:       "5002",
			ActorName:     "管理员",
			Action:        "comment.moderate",
			ResourceType:  "comment",
			ResourceID:    "6001",
			ResourceTitle: "评论审核",
			Status:        "success",
			IP:            "127.0.0.1",
			UserAgent:     "seed",
			Detail:        "审核通过读者评论",
			CreatedAt:     now.Add(-5 * time.Hour),
		},
		{
			ID:            "10003",
			ActorID:       "5002",
			ActorName:     "管理员",
			Action:        "settings.update",
			ResourceType:  "settings",
			ResourceTitle: "系统设置",
			Status:        "success",
			IP:            "127.0.0.1",
			UserAgent:     "seed",
			Detail:        "更新评论和投稿策略",
			CreatedAt:     now.Add(-24 * time.Hour),
		},
	}
}
