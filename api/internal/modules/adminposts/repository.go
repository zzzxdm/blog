package adminposts

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"blog/api/internal/modules/posts"
)

var (
	ErrPostNotFound     = errors.New("admin post not found")
	ErrRevisionNotFound = errors.New("admin post revision not found")
	ErrInvalidPost      = errors.New("invalid admin post")
	ErrPostNotPublic    = errors.New("admin post is not public")
)

type Repository interface {
	List(ctx context.Context, query ListQuery) (ListResult, error)
	Get(ctx context.Context, id string) (AdminPost, error)
	Save(ctx context.Context, id string, request SaveRequest) (AdminPost, error)
	Delete(ctx context.Context, id string) (AdminPost, error)
	Publish(ctx context.Context, id string, publisher posts.Publisher) (AdminPost, error)
	PublishDue(ctx context.Context, publisher posts.Publisher, now time.Time) (int, error)
	ListRevisions(ctx context.Context, id string) (RevisionListResult, error)
	RestoreRevision(ctx context.Context, id string, revisionID string) (AdminPost, error)
}

func countStats(items []AdminPost) Stats {
	return countStatsAt(items, time.Now())
}

func filterAdminPosts(items []AdminPost, query ListQuery) []AdminPost {
	keyword := strings.ToLower(strings.TrimSpace(query.Keyword))
	status := strings.ToLower(strings.TrimSpace(query.Status))
	filtered := make([]AdminPost, 0, len(items))

	for _, item := range items {
		if status != "" && status != "all" && item.Status != status {
			continue
		}
		if keyword != "" && !adminPostContains(item, keyword) {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered
}

func adminPostContains(item AdminPost, keyword string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		item.Title,
		item.Summary,
		item.AuthorName,
		item.Category,
		item.Slug,
		item.Visibility,
		item.Status,
		strings.Join(item.Tags, " "),
	}, " "))
	return strings.Contains(haystack, keyword)
}

func sortAdminPosts(items []AdminPost, mode string) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "views":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].ViewCount > items[j].ViewCount
		})
	case "scheduled":
		sort.SliceStable(items, func(i, j int) bool {
			left := time.Time{}
			right := time.Time{}
			if items[i].ScheduledAt != nil {
				left = *items[i].ScheduledAt
			}
			if items[j].ScheduledAt != nil {
				right = *items[j].ScheduledAt
			}
			if left.IsZero() {
				return false
			}
			if right.IsZero() {
				return true
			}
			return left.Before(right)
		})
	default:
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].UpdatedAt.After(items[j].UpdatedAt)
		})
	}
}

func pagedPostResult(items []AdminPost, stats Stats, query ListQuery) ListResult {
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

	return ListResult{
		Items:    paged,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Stats:    stats,
	}
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return 10
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

func countStatsAt(items []AdminPost, now time.Time) Stats {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := monthStart.AddDate(0, 1, 0)

	stats := Stats{Total: len(items)}
	monthlyViews := 0
	for _, item := range items {
		switch item.Status {
		case StatusPublished:
			stats.Published++
			if postPublishedInRange(item, monthStart, nextMonth) {
				monthlyViews += item.ViewCount
			}
		case StatusDraft:
			stats.Draft++
		case StatusReview:
			stats.Review++
		case StatusScheduled:
			stats.Scheduled++
		}
	}
	stats.MonthlyViews = formatCompactCount(monthlyViews)

	return stats
}

func postPublishedInRange(item AdminPost, start time.Time, end time.Time) bool {
	if item.PublishedAt != nil {
		return !item.PublishedAt.Before(start) && item.PublishedAt.Before(end)
	}

	if item.UpdatedAt.IsZero() {
		return false
	}

	return !item.UpdatedAt.Before(start) && item.UpdatedAt.Before(end)
}

func formatCompactCount(value int) string {
	if value < 1000 {
		return fmt.Sprintf("%d", value)
	}
	if value < 1000000 {
		whole := value / 1000
		decimal := (value % 1000) / 100
		if decimal == 0 {
			return fmt.Sprintf("%dk", whole)
		}
		return fmt.Sprintf("%d.%dk", whole, decimal)
	}

	whole := value / 1000000
	decimal := (value % 1000000) / 100000
	if decimal == 0 {
		return fmt.Sprintf("%dm", whole)
	}

	return fmt.Sprintf("%d.%dm", whole, decimal)
}

func clonePost(item AdminPost) AdminPost {
	item.Visibility = normalizeVisibility(item.Visibility)
	item.Tags = append([]string{}, item.Tags...)
	item.Revisions = cloneRevisions(item.Revisions)
	return item
}

func snapshotRevision(item AdminPost, createdAt time.Time) Revision {
	return Revision{
		ID:             fmt.Sprintf("%s_rev_%d", item.ID, item.Version),
		Version:        item.Version,
		Slug:           item.Slug,
		Title:          item.Title,
		Summary:        item.Summary,
		Content:        item.Content,
		Status:         item.Status,
		Visibility:     normalizeVisibility(item.Visibility),
		Category:       item.Category,
		Tags:           append([]string{}, item.Tags...),
		CoverImage:     item.CoverImage,
		SEOtitle:       item.SEOtitle,
		SEODescription: item.SEODescription,
		AuthorName:     item.AuthorName,
		CreatedAt:      createdAt,
	}
}

func appendRevision(revisions []Revision, revision Revision) []Revision {
	filtered := make([]Revision, 0, len(revisions)+1)
	for _, item := range revisions {
		if item.ID == revision.ID || item.Version == revision.Version {
			continue
		}

		filtered = append(filtered, cloneRevision(item))
	}

	filtered = append(filtered, cloneRevision(revision))
	return filtered
}

func sortedRevisions(item AdminPost) []Revision {
	revisions := cloneRevisions(item.Revisions)
	if len(revisions) == 0 && item.Version > 0 {
		revisions = append(revisions, snapshotRevision(item, item.UpdatedAt))
	}

	sort.SliceStable(revisions, func(i, j int) bool {
		return revisions[i].Version > revisions[j].Version
	})

	return revisions
}

func findRevision(item AdminPost, revisionID string) (Revision, bool) {
	for _, revision := range sortedRevisions(item) {
		if revision.ID == revisionID {
			return cloneRevision(revision), true
		}
	}

	return Revision{}, false
}

func restoreFromRevision(item AdminPost, revision Revision) AdminPost {
	item.Slug = defaultString(strings.TrimSpace(revision.Slug), item.Slug)
	item.Title = revision.Title
	item.Summary = revision.Summary
	item.Content = revision.Content
	item.Visibility = normalizeVisibility(revision.Visibility)
	item.Category = defaultString(strings.TrimSpace(revision.Category), item.Category)
	item.Tags = normalizeTags(revision.Tags)
	item.CoverImage = revision.CoverImage
	item.SEOtitle = defaultString(strings.TrimSpace(revision.SEOtitle), revision.Title)
	item.SEODescription = defaultString(strings.TrimSpace(revision.SEODescription), revision.Summary)
	return item
}

func cloneRevisions(revisions []Revision) []Revision {
	result := make([]Revision, 0, len(revisions))
	for _, revision := range revisions {
		result = append(result, cloneRevision(revision))
	}

	return result
}

func cloneRevision(revision Revision) Revision {
	revision.Visibility = normalizeVisibility(revision.Visibility)
	revision.Tags = append([]string{}, revision.Tags...)
	return revision
}

func normalizeStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case StatusPublished, StatusReview, StatusScheduled, StatusArchived:
		return status
	default:
		return StatusDraft
	}
}

func parseScheduledAt(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, ErrInvalidPost
	}

	return &parsed, nil
}

func isDueScheduledPost(item AdminPost, now time.Time) bool {
	return item.Status == StatusScheduled &&
		item.ScheduledAt != nil &&
		!item.ScheduledAt.After(now) &&
		normalizeVisibility(item.Visibility) == VisibilityPublic
}

func normalizeVisibility(visibility string) string {
	visibility = strings.ToLower(strings.TrimSpace(visibility))
	switch visibility {
	case VisibilityPrivate, VisibilityMembers:
		return visibility
	default:
		return VisibilityPublic
	}
}

func normalizeTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, tag := range tags {
		value := strings.TrimSpace(tag)
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}

	return result
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, item := range value {
		isLetter := item >= 'a' && item <= 'z'
		isDigit := item >= '0' && item <= '9'
		if isLetter || isDigit {
			builder.WriteRune(item)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func estimateReadingTime(content string) int {
	runes := len([]rune(strings.TrimSpace(content)))
	if runes == 0 {
		return 1
	}
	minutes := runes / 500
	if runes%500 != 0 {
		minutes++
	}
	if minutes < 1 {
		return 1
	}

	return minutes
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func seedAdminPosts() map[string]AdminPost {
	now := time.Now()
	publishedAt := now.Add(-2 * time.Hour)
	scheduledAt := now.Add(6 * time.Hour)
	items := []AdminPost{
		{
			ID:                "admin_post_001",
			Slug:              "blog-system-design",
			Title:             "如何设计一个内容长期增长的博客系统",
			Summary:           "博客不是文章列表加详情页。真正可持续的系统需要同时照顾写作、发布、搜索、运营、迁移和长期维护。",
			Content:           "一个现代化博客系统需要从内容资产的生命周期开始设计。",
			Status:            StatusPublished,
			Visibility:        VisibilityPublic,
			Category:          "工程实践",
			Tags:              []string{"博客系统", "架构", "内容治理"},
			CoverImage:        "https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=1400&q=80",
			AuthorName:        "管理员",
			ReadingTime:       12,
			ViewCount:         2984,
			CommentCount:      34,
			SEOtitle:          "如何设计一个现代化博客系统",
			SEODescription:    "从内容模型、发布流程、SEO、缓存和运营能力设计一个可长期维护的博客系统。",
			Version:           4,
			PublishedPostSlug: "blog-system-design",
			PublishedAt:       &publishedAt,
			UpdatedAt:         now.Add(-2 * time.Hour),
		},
		{
			ID:          "admin_post_002",
			Slug:        "vue3-content-site-cache-seo",
			Title:       "Vue3 内容站的缓存与 SEO 边界",
			Summary:     "客户端渲染、接口缓存和服务端 meta 需要明确边界。",
			Content:     "Vue3 内容站可以保持前端开发效率，同时通过 Go 输出基础 HTML。",
			Status:      StatusScheduled,
			Visibility:  VisibilityPublic,
			Category:    "Vue3",
			Tags:        []string{"Vue3", "SEO", "缓存"},
			CoverImage:  "https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=1400&q=80",
			AuthorName:  "管理员",
			Version:     2,
			ScheduledAt: &scheduledAt,
			UpdatedAt:   now.Add(-8 * time.Hour),
		},
		{
			ID:           "admin_post_003",
			Slug:         "post-version-history",
			Title:        "为什么博客后台需要文章版本历史",
			Summary:      "版本记录不是复杂功能，而是内容资产的基本保险。",
			Content:      "文章会被持续修订，后台需要记录版本历史、修改人、变更摘要和回滚能力。",
			Status:       StatusReview,
			Visibility:   VisibilityPublic,
			Category:     "内容治理",
			Tags:         []string{"版本历史", "内容治理"},
			CoverImage:   "https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=1400&q=80",
			AuthorName:   "管理员",
			ViewCount:    1988,
			CommentCount: 12,
			Version:      3,
			UpdatedAt:    now.AddDate(0, 0, -1),
		},
		{
			ID:         "admin_post_004",
			Slug:       "markdown-writing-experience",
			Title:      "把 Markdown 写作体验做到顺手",
			Summary:    "编辑器、预览、封面和 SEO 字段要服务写作流程。",
			Content:    "Markdown 编辑器需要稳定的草稿保存、预览、图片插入、代码块处理和 SEO 字段编辑。",
			Status:     StatusDraft,
			Visibility: VisibilityPublic,
			Category:   "编辑器",
			Tags:       []string{"Markdown", "写作工作流"},
			AuthorName: "管理员",
			Version:    1,
			UpdatedAt:  now.Add(-2 * time.Minute),
		},
	}

	result := map[string]AdminPost{}
	for _, item := range items {
		result[item.ID] = item
	}

	return result
}
