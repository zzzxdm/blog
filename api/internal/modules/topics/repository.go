package topics

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
)

var (
	ErrNotFound  = errors.New("topic not found")
	ErrInvalid   = errors.New("invalid topic")
	ErrDuplicate = errors.New("topic duplicate")
)

type Repository interface {
	List(ctx context.Context, query ListQuery) (ListResult, error)
	GetBySlug(ctx context.Context, slug string) (Topic, error)
	Save(ctx context.Context, id string, request SaveRequest) (Topic, error)
	Delete(ctx context.Context, id string) error
}

func matchesListQuery(item Topic, query ListQuery) bool {
	if !query.All && item.Status != "active" {
		return false
	}
	if query.Status != "" && !strings.EqualFold(item.Status, query.Status) {
		return false
	}
	if query.Featured && !item.Featured {
		return false
	}
	if !matchesKeyword(item, query.Keyword) {
		return false
	}
	return true
}

func matchesKeyword(item Topic, keyword string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}

	text := strings.ToLower(strings.Join([]string{
		item.Slug,
		item.Title,
		item.Summary,
		item.ImageAlt,
		item.Tone,
		item.Status,
		strings.Join(item.Categories, " "),
		strings.Join(item.Tags, " "),
	}, " "))
	return strings.Contains(text, keyword)
}

func sortTopics(items []Topic) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].SortOrder == items[j].SortOrder {
			return items[i].Title < items[j].Title
		}
		return items[i].SortOrder < items[j].SortOrder
	})
}

func pageItems[T any](items []T, page int, pageSize int) []T {
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
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

func normalizeStatus(value string) string {
	status := strings.ToLower(strings.TrimSpace(value))
	if status == "draft" || status == "active" {
		return status
	}
	return "active"
}

func normalizeTone(value string) string {
	tone := strings.ToLower(strings.TrimSpace(value))
	if tone == "rust" || tone == "amber" || tone == "gray" {
		return tone
	}
	return ""
}

func normalizeStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, item := range values {
		value := strings.TrimSpace(item)
		if value == "" || seen[strings.ToLower(value)] {
			continue
		}
		seen[strings.ToLower(value)] = true
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

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func cloneTopic(item Topic) Topic {
	item.Categories = append([]string{}, item.Categories...)
	item.Tags = append([]string{}, item.Tags...)
	return item
}

func seedTopics(now func() time.Time) map[string]Topic {
	createdAt := now()
	items := []Topic{
		{
			ID:         "4001",
			Slug:       "blog-system",
			Title:      "现代化博客系统",
			Summary:    "从产品功能、技术架构、用户系统、评论、搜索和后台管理完整设计一个博客系统。",
			CoverImage: "https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=900&q=80",
			ImageAlt:   "代码编辑器和开发设备",
			Status:     "active",
			Featured:   true,
			SortOrder:  10,
			Categories: []string{"工程实践", "产品设计", "用户系统", "内容治理"},
			Tags:       []string{"博客系统", "架构", "内容治理", "评论"},
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		},
		{
			ID:         "4002",
			Slug:       "vue3-content",
			Title:      "Vue3 内容站",
			Summary:    "路由、状态管理、接口缓存、SEO meta、图片优化和部署策略。",
			CoverImage: "https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=900&q=80",
			ImageAlt:   "代码编辑器中的程序文件",
			Tone:       "rust",
			Status:     "active",
			Featured:   true,
			SortOrder:  20,
			Categories: []string{"Vue3"},
			Tags:       []string{"Vue3", "SEO", "缓存"},
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		},
		{
			ID:         "4003",
			Slug:       "writing-workflow",
			Title:      "写作工作流",
			Summary:    "草稿、版本历史、编辑器、发布审批和长期内容维护。",
			CoverImage: "https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=900&q=80",
			ImageAlt:   "笔记本和写作草稿",
			Tone:       "amber",
			Status:     "active",
			Featured:   true,
			SortOrder:  30,
			Categories: []string{"写作工作流"},
			Tags:       []string{"工作流", "写作工作流", "Markdown"},
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		},
		{
			ID:         "4004",
			Slug:       "resource-list",
			Title:      "资源清单",
			Summary:    "把工具、部署、数据库和内容运营资料整理成可持续更新的阅读路线。",
			CoverImage: "https://images.unsplash.com/photo-1484480974693-6ca0a78fb36b?auto=format&fit=crop&w=900&q=80",
			ImageAlt:   "桌面上的计划清单和电脑",
			Status:     "active",
			Featured:   true,
			SortOrder:  40,
			Categories: []string{"架构", "运营"},
			Tags:       []string{"PostgreSQL", "Redis", "全文搜索", "SEO"},
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		},
	}

	result := map[string]Topic{}
	for _, item := range items {
		result[item.ID] = item
	}
	return result
}
