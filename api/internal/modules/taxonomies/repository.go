package taxonomies

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrNotFound      = errors.New("taxonomy not found")
	ErrInvalid       = errors.New("invalid taxonomy")
	ErrDuplicate     = errors.New("taxonomy duplicate")
	ErrTaxonomyInUse = errors.New("taxonomy in use")
)

type Repository interface {
	ListCategories(ctx context.Context) ([]Category, error)
	SaveCategory(ctx context.Context, id string, request SaveCategoryRequest) (Category, error)
	DeleteCategory(ctx context.Context, id string) error
	ListTags(ctx context.Context) ([]Tag, error)
	SaveTag(ctx context.Context, id string, request SaveTagRequest) (Tag, error)
	DeleteTag(ctx context.Context, id string) error
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

func seedCategories() map[string]Category {
	items := []Category{
		{ID: "category_engineering", Slug: "engineering", Name: "工程实践", Description: "工程方法、架构落地和长期维护经验。", SortOrder: 10, PostCount: 3},
		{ID: "category_architecture", Slug: "architecture", Name: "架构", Description: "系统边界、数据层和基础设施设计。", SortOrder: 20, PostCount: 3},
		{ID: "category_product_design", Slug: "product-design", Name: "产品设计", Description: "信息架构、交互体验和内容产品设计。", SortOrder: 30, PostCount: 2},
		{ID: "category_operations", Slug: "operations", Name: "运营", Description: "内容运营、增长反馈和站点治理。", SortOrder: 40, PostCount: 2},
		{ID: "category_vue3", Slug: "vue3", Name: "Vue3", Description: "Vue3 内容站、前端架构和交互实现。", SortOrder: 50, PostCount: 2},
		{ID: "category_workflow", Slug: "workflow", Name: "写作工作流", Description: "投稿、审核、编辑器和内容生命周期。", SortOrder: 60, PostCount: 3},
	}

	result := map[string]Category{}
	for _, item := range items {
		result[item.ID] = item
	}

	return result
}

func seedTags() map[string]Tag {
	items := []Tag{
		{ID: "tag_blog_system", Slug: "blog-system", Name: "博客系统", PostCount: 3},
		{ID: "tag_architecture", Slug: "architecture", Name: "架构", PostCount: 2},
		{ID: "tag_content_governance", Slug: "content-governance", Name: "内容治理", PostCount: 2},
		{ID: "tag_vue3", Slug: "vue3", Name: "Vue3", PostCount: 2},
		{ID: "tag_seo", Slug: "seo", Name: "SEO", PostCount: 1},
		{ID: "tag_cache", Slug: "cache", Name: "缓存", PostCount: 1},
		{ID: "tag_postgresql", Slug: "postgresql", Name: "PostgreSQL", PostCount: 2},
		{ID: "tag_redis", Slug: "redis", Name: "Redis", PostCount: 1},
		{ID: "tag_full_text_search", Slug: "full-text-search", Name: "全文搜索", PostCount: 2},
		{ID: "tag_submission", Slug: "submission", Name: "投稿", PostCount: 1},
		{ID: "tag_message", Slug: "message", Name: "站内信", PostCount: 1},
		{ID: "tag_account", Slug: "account", Name: "账号", PostCount: 1},
	}

	result := map[string]Tag{}
	for _, item := range items {
		result[item.ID] = item
	}

	return result
}
