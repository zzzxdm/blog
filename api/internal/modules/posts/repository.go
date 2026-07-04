package posts

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("post not found")
var ErrInvalidPost = errors.New("invalid post")

type Repository interface {
	List(ctx context.Context, query ListQuery) (ListResult, error)
	GetBySlug(ctx context.Context, slug string) (Post, error)
}

type MemoryRepository struct {
	mu    sync.RWMutex
	posts []Post
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{posts: seedPosts()}
}

func (repo *MemoryRepository) List(_ context.Context, query ListQuery) (ListResult, error) {
	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	filtered := make([]Post, 0, len(repo.posts))

	repo.mu.RLock()
	for _, post := range repo.posts {
		if !matches(post, query) {
			continue
		}

		filtered = append(filtered, post)
	}
	repo.mu.RUnlock()

	sortPosts(filtered, query.Sort)

	total := len(filtered)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return ListResult{
		Items:    filtered[start:end],
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (repo *MemoryRepository) GetBySlug(_ context.Context, slug string) (Post, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	for _, post := range repo.posts {
		if post.Slug == slug {
			return post, nil
		}
	}

	return Post{}, ErrNotFound
}

func (repo *MemoryRepository) RecordView(_ context.Context, slug string) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for index := range repo.posts {
		if repo.posts[index].Slug == slug {
			repo.posts[index].ViewCount++
			return repo.posts[index], nil
		}
	}

	return Post{}, ErrNotFound
}

func (repo *MemoryRepository) Publish(_ context.Context, input PublishInput) (Post, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if title == "" || content == "" {
		return Post{}, ErrInvalidPost
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	slug := repo.uniqueSlugLocked(defaultString(slugify(input.Slug), slugify(title)))
	if slug == "" {
		slug = repo.uniqueSlugLocked(fmt.Sprintf("post-%03d", len(repo.posts)+1))
	}

	post := Post{
		ID:           fmt.Sprintf("post_memory_%03d", len(repo.posts)+1),
		Slug:         slug,
		Title:        title,
		Summary:      strings.TrimSpace(input.Summary),
		Content:      content,
		Category:     defaultString(strings.TrimSpace(input.Category), "投稿"),
		Tags:         normalizeTags(input.Tags),
		CoverImage:   defaultString(strings.TrimSpace(input.CoverImage), "https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=1400&q=80"),
		AuthorName:   defaultString(strings.TrimSpace(input.AuthorName), "注册用户"),
		ReadingTime:  estimateReadingTime(content),
		ViewCount:    0,
		LikeCount:    0,
		DislikeCount: 0,
		CommentCount: 0,
		PublishedAt:  time.Now(),
	}
	repo.posts = append(repo.posts, post)

	return post, nil
}

func (repo *MemoryRepository) uniqueSlugLocked(slug string) string {
	if slug == "" {
		return ""
	}

	candidate := slug
	for suffix := 2; repo.hasSlugLocked(candidate); suffix++ {
		candidate = fmt.Sprintf("%s-%d", slug, suffix)
	}

	return candidate
}

func (repo *MemoryRepository) hasSlugLocked(slug string) bool {
	for _, post := range repo.posts {
		if post.Slug == slug {
			return true
		}
	}

	return false
}

func matches(post Post, query ListQuery) bool {
	if query.Category != "" && !strings.EqualFold(post.Category, query.Category) {
		return false
	}

	if query.Tag != "" && !hasTag(post.Tags, query.Tag) {
		return false
	}

	if query.Keyword == "" {
		return true
	}

	keyword := strings.ToLower(query.Keyword)
	text := strings.ToLower(strings.Join([]string{
		post.Title,
		post.Summary,
		post.Content,
		post.Category,
		strings.Join(post.Tags, " "),
	}, " "))

	return strings.Contains(text, keyword)
}

func hasTag(tags []string, tag string) bool {
	for _, item := range tags {
		if strings.EqualFold(item, tag) {
			return true
		}
	}

	return false
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

func normalizeTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, tag := range tags {
		value := strings.TrimSpace(tag)
		if value == "" || seen[strings.ToLower(value)] {
			continue
		}

		seen[strings.ToLower(value)] = true
		result = append(result, value)
	}

	return result
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

func sortPosts(posts []Post, sortMode string) {
	sort.SliceStable(posts, func(i, j int) bool {
		switch strings.ToLower(sortMode) {
		case "views":
			return posts[i].ViewCount > posts[j].ViewCount
		case "comments":
			return posts[i].CommentCount > posts[j].CommentCount
		case "likes":
			return posts[i].LikeCount > posts[j].LikeCount
		default:
			return posts[i].PublishedAt.After(posts[j].PublishedAt)
		}
	})
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

	if pageSize > 50 {
		return 50
	}

	return pageSize
}

func mustDate(value string) time.Time {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		panic(err)
	}

	return parsed
}

func seedPosts() []Post {
	return []Post{
		{
			ID:           "post_001",
			Slug:         "blog-system-design",
			Title:        "如何设计一个内容长期增长的博客系统",
			Summary:      "博客不是文章列表加详情页。真正可持续的系统需要同时照顾写作、发布、搜索、运营、迁移和长期维护。",
			Content:      "一个现代化博客系统需要从内容资产的生命周期开始设计。文章不是一次性页面，而是会被修改、引用、搜索、迁移和长期展示的结构化内容。",
			Category:     "工程实践",
			Tags:         []string{"博客系统", "架构", "内容治理"},
			CoverImage:   "https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=1200&q=80",
			AuthorName:   "管理员",
			ReadingTime:  12,
			ViewCount:    2984,
			LikeCount:    128,
			DislikeCount: 7,
			CommentCount: 34,
			PublishedAt:  mustDate("2026-07-04"),
		},
		{
			ID:           "post_002",
			Slug:         "vue3-content-site-cache-seo",
			Title:        "Vue3 内容站的缓存与 SEO 边界",
			Summary:      "客户端渲染、接口缓存和服务端 meta 需要明确边界，避免前期开发轻松、后期收录困难。",
			Content:      "Vue3 内容站可以保持前端开发效率，同时通过 Go 输出基础 HTML、meta 和结构化数据处理文章页 SEO。",
			Category:     "Vue3",
			Tags:         []string{"Vue3", "SEO", "缓存"},
			CoverImage:   "https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=1200&q=80",
			AuthorName:   "管理员",
			ReadingTime:  8,
			ViewCount:    4120,
			LikeCount:    96,
			DislikeCount: 3,
			CommentCount: 18,
			PublishedAt:  mustDate("2026-06-25"),
		},
		{
			ID:           "post_003",
			Slug:         "postgres-redis-blog-boundary",
			Title:        "Redis 和 PostgreSQL 在博客中的分工",
			Summary:      "PostgreSQL 保存事实并承担全文搜索，Redis 负责热点读取、会话、限流和异步任务协调。",
			Content:      "个人博客早期没有必要引入专用搜索中间件。PostgreSQL 的 tsvector 和 GIN 索引足以覆盖标题、摘要、正文和标签搜索。",
			Category:     "架构",
			Tags:         []string{"PostgreSQL", "Redis", "全文搜索"},
			CoverImage:   "https://images.unsplash.com/photo-1558494949-ef010cbdcc31?auto=format&fit=crop&w=1200&q=80",
			AuthorName:   "管理员",
			ReadingTime:  14,
			ViewCount:    3019,
			LikeCount:    84,
			DislikeCount: 4,
			CommentCount: 25,
			PublishedAt:  mustDate("2026-07-01"),
		},
	}
}
