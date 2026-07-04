package operations

import (
	"context"
	"sync"
	"time"
)

type Repository interface {
	GetSettings(ctx context.Context) (Settings, error)
	UpdateSettings(ctx context.Context, settings Settings) (Settings, error)
	GetNavigation(ctx context.Context) (Navigation, error)
	UpdateNavigation(ctx context.Context, navigation Navigation) (Navigation, error)
	ListMedia(ctx context.Context) (MediaListResult, error)
	CreateMedia(ctx context.Context, asset MediaAsset) (MediaAsset, error)
	GetStats(ctx context.Context) (Stats, error)
}

type MemoryRepository struct {
	mu         sync.RWMutex
	settings   Settings
	navigation Navigation
	media      []MediaAsset
	stats      Stats
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		settings:   seedSettings(),
		navigation: seedNavigation(),
		media:      seedMedia(),
		stats:      seedStats(),
	}
}

func (repo *MemoryRepository) GetSettings(_ context.Context) (Settings, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	return cloneSettings(repo.settings), nil
}

func (repo *MemoryRepository) UpdateSettings(_ context.Context, settings Settings) (Settings, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	settings.UpdatedAt = time.Now()
	repo.settings = cloneSettings(settings)

	return cloneSettings(repo.settings), nil
}

func (repo *MemoryRepository) GetNavigation(_ context.Context) (Navigation, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	return cloneNavigation(repo.navigation), nil
}

func (repo *MemoryRepository) UpdateNavigation(_ context.Context, navigation Navigation) (Navigation, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	navigation.UpdatedAt = time.Now()
	repo.navigation = cloneNavigation(navigation)

	return cloneNavigation(repo.navigation), nil
}

func (repo *MemoryRepository) ListMedia(_ context.Context) (MediaListResult, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	items := append([]MediaAsset{}, repo.media...)
	return MediaListResult{
		Items: items,
		Total: len(items),
	}, nil
}

func (repo *MemoryRepository) CreateMedia(_ context.Context, asset MediaAsset) (MediaAsset, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	repo.media = append([]MediaAsset{asset}, repo.media...)
	return asset, nil
}

func (repo *MemoryRepository) GetStats(_ context.Context) (Stats, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	return cloneStats(repo.stats), nil
}

func cloneSettings(settings Settings) Settings {
	settings.BlockedWords = append([]string{}, settings.BlockedWords...)
	return settings
}

func cloneNavigation(navigation Navigation) Navigation {
	navigation.TopItems = append([]NavItem{}, navigation.TopItems...)
	navigation.FooterItems = append([]NavItem{}, navigation.FooterItems...)
	navigation.Redirects = append([]RedirectRule{}, navigation.Redirects...)
	return navigation
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
		FromEmail:               "newsletter@example.com",
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
			{ID: "nav_top_4", Label: "关于", URL: "/about", Order: 4},
		},
		FooterItems: []NavItem{
			{ID: "nav_footer_1", Label: "RSS", URL: "/rss.xml", Order: 1},
			{ID: "nav_footer_2", Label: "友情链接", URL: "/friends", Order: 2},
			{ID: "nav_footer_3", Label: "隐私政策", URL: "/privacy", Order: 3},
		},
		MobileCollapse:         true,
		ExternalLinksNewWindow: true,
		ShowLoginEntry:         true,
		GitHubURL:              "https://github.com/example",
		ContactEmail:           "hello@example.com",
		RSSURL:                 "/rss.xml",
		Redirects: []RedirectRule{
			{From: "/old-blog-design", To: "/posts/blog-system-design", Code: 301},
			{From: "/newsletter", To: "/topics/content-growth", Code: 302},
		},
		UpdatedAt: time.Now(),
	}
}

func seedMedia() []MediaAsset {
	now := time.Now()
	return []MediaAsset{
		{ID: "media_001", FileName: "cover-code-desk.jpg", URL: "https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=700&q=80", Alt: "代码编辑器封面图", Type: "image", Category: "封面图", SizeLabel: "1.2 MB", Width: 1400, Height: 930, UsageCount: 8, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -1)},
		{ID: "media_002", FileName: "server-room.jpg", URL: "https://images.unsplash.com/photo-1558494949-ef010cbdcc31?auto=format&fit=crop&w=700&q=80", Alt: "服务器设备图片", Type: "image", Category: "架构", SizeLabel: "980 KB", Width: 1400, Height: 930, UsageCount: 3, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -2)},
		{ID: "media_003", FileName: "writing-notes.jpg", URL: "https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=700&q=80", Alt: "写作桌面图片", Type: "image", Category: "写作", SizeLabel: "860 KB", Width: 1400, Height: 930, UsageCount: 5, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -3)},
		{ID: "media_004", FileName: "product-dashboard.jpg", URL: "https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=700&q=80", Alt: "产品分析界面", Type: "image", Category: "产品设计", SizeLabel: "1.4 MB", Width: 1400, Height: 930, UsageCount: 2, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -4)},
		{ID: "media_005", FileName: "javascript-editor.jpg", URL: "https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=700&q=80", Alt: "代码窗口", Type: "image", Category: "Vue3", SizeLabel: "740 KB", Width: 1400, Height: 930, UsageCount: 4, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -5)},
		{ID: "media_006", FileName: "quiet-workspace.jpg", URL: "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=700&q=80", Alt: "自然光工作区", Type: "image", Category: "备用", SizeLabel: "690 KB", Width: 1400, Height: 930, UsageCount: 0, UploadedBy: "管理员", UploadedAt: now.AddDate(0, 0, -6)},
	}
}

func seedStats() Stats {
	return Stats{
		Metrics: []Metric{
			{Label: "PV", Value: "86.4k", Delta: "较上期 +18%"},
			{Label: "UV", Value: "24.7k", Delta: "较上期 +9%"},
			{Label: "平均阅读", Value: "4:18", Delta: "提升 32 秒"},
			{Label: "RSS 访问", Value: "418", Delta: "转化率 2.6%"},
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
			{Title: "Vue3 内容站的缓存与 SEO 边界", Views: "12,420", Bookmarks: 312, Comments: 48, RSSRate: "3.8%"},
			{Title: "如何设计一个内容长期增长的博客系统", Views: "9,884", Bookmarks: 286, Comments: 34, RSSRate: "4.1%"},
			{Title: "让旧文章继续被搜索引擎找到", Views: "7,209", Bookmarks: 190, Comments: 19, RSSRate: "2.4%"},
		},
		Sources: []BarPoint{
			{Label: "搜索", Value: "46%", Percent: 72},
			{Label: "直接", Value: "22%", Percent: 34, Tone: "rust"},
			{Label: "社交", Value: "18%", Percent: 28, Tone: "amber"},
			{Label: "订阅", Value: "14%", Percent: 20},
		},
		SearchTerms: []SearchTerm{
			{Term: "Vue3 SEO", Count: 1284},
			{Term: "博客系统设计", Count: 936},
			{Term: "Markdown 编辑器", Count: 642},
		},
		Suggestions: []ContentSuggestion{
			{Title: "搜索词“评论审核”增长明显", Body: "可以补一篇用户评论系统设计。"},
			{Title: "专题页带来 18% 收藏", Body: "建议继续完善专题导航。"},
		},
	}
}
