package topics

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
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

type MemoryRepository struct {
	mu     sync.RWMutex
	topics map[string]Topic
	nextID int
	now    func() time.Time
}

func NewMemoryRepository() *MemoryRepository {
	now := time.Now
	return &MemoryRepository{
		topics: seedTopics(now),
		nextID: 100,
		now:    now,
	}
}

func (repo *MemoryRepository) List(_ context.Context, query ListQuery) (ListResult, error) {
	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	filtered := make([]Topic, 0, len(repo.topics))

	repo.mu.RLock()
	for _, item := range repo.topics {
		if !matchesListQuery(item, query) {
			continue
		}
		filtered = append(filtered, cloneTopic(item))
	}
	repo.mu.RUnlock()

	sortTopics(filtered)
	total := len(filtered)
	return ListResult{
		Items:    pageItems(filtered, page, pageSize),
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (repo *MemoryRepository) GetBySlug(_ context.Context, slug string) (Topic, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	for _, item := range repo.topics {
		if strings.EqualFold(item.Slug, strings.TrimSpace(slug)) {
			return cloneTopic(item), nil
		}
	}

	return Topic{}, ErrNotFound
}

func (repo *MemoryRepository) Save(_ context.Context, id string, request SaveRequest) (Topic, error) {
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return Topic{}, ErrInvalid
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	slug := repo.topicSlug(request.Slug, title)
	if repo.duplicateLocked(id, slug, title) {
		return Topic{}, ErrDuplicate
	}

	item, ok := repo.topics[id]
	if id != "" && !ok {
		return Topic{}, ErrNotFound
	}

	now := repo.now()
	if id == "" {
		id = fmt.Sprintf("topic_%03d", repo.nextID)
		repo.nextID++
		item = Topic{ID: id, CreatedAt: now}
	}

	item.Slug = slug
	item.Title = title
	item.Summary = strings.TrimSpace(request.Summary)
	item.CoverImage = strings.TrimSpace(request.CoverImage)
	item.ImageAlt = strings.TrimSpace(request.ImageAlt)
	item.Tone = normalizeTone(request.Tone)
	item.Status = normalizeStatus(request.Status)
	item.Featured = request.Featured
	item.SortOrder = request.SortOrder
	item.Categories = normalizeStrings(request.Categories)
	item.Tags = normalizeStrings(request.Tags)
	item.UpdatedAt = now
	repo.topics[item.ID] = item

	return cloneTopic(item), nil
}

func (repo *MemoryRepository) Delete(_ context.Context, id string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, ok := repo.topics[id]; !ok {
		return ErrNotFound
	}

	delete(repo.topics, id)
	return nil
}

func (repo *MemoryRepository) topicSlug(value string, title string) string {
	slug := defaultString(slugify(value), slugify(title))
	if slug == "" {
		slug = fmt.Sprintf("topic-%03d", repo.nextID)
	}

	return slug
}

func (repo *MemoryRepository) duplicateLocked(id string, slug string, title string) bool {
	for _, item := range repo.topics {
		if item.ID == id {
			continue
		}
		if strings.EqualFold(item.Slug, slug) || strings.EqualFold(item.Title, title) {
			return true
		}
	}

	return false
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
			ID:         "topic_blog_system",
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
			ID:         "topic_vue3_content",
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
			ID:         "topic_writing_workflow",
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
			ID:         "topic_resource_list",
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
