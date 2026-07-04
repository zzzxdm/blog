package posts

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
)

var ErrNotFound = errors.New("post not found")

type Repository interface {
	List(ctx context.Context, query ListQuery) (ListResult, error)
	GetBySlug(ctx context.Context, slug string) (Post, error)
}

type MemoryRepository struct {
	posts []Post
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{posts: seedPosts()}
}

func (repo *MemoryRepository) List(_ context.Context, query ListQuery) (ListResult, error) {
	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	filtered := make([]Post, 0, len(repo.posts))

	for _, post := range repo.posts {
		if !matches(post, query) {
			continue
		}

		filtered = append(filtered, post)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].PublishedAt.After(filtered[j].PublishedAt)
	})

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
	for _, post := range repo.posts {
		if post.Slug == slug {
			return post, nil
		}
	}

	return Post{}, ErrNotFound
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
