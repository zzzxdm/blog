package comments

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
)

var ErrEmptyBody = errors.New("comment body is empty")
var ErrCommentNotFound = errors.New("comment not found")
var ErrInvalidStatus = errors.New("invalid comment status")
var ErrForbidden = errors.New("comment forbidden")

type Repository interface {
	List(ctx context.Context, postSlug string, viewerID string) (ListResult, error)
	Create(ctx context.Context, postSlug string, request CreateRequest, user auth.User) (Comment, error)
	CreateReply(ctx context.Context, parentID string, request CreateRequest, user auth.User) (Comment, error)
	ListByAuthor(ctx context.Context, userID string, query ListQuery) (ManageListResult, error)
	AdminList(ctx context.Context, query ListQuery) (ManageListResult, error)
	UpdateStatus(ctx context.Context, commentID string, status string) (Comment, error)
	DeleteByAuthor(ctx context.Context, commentID string, userID string) (Comment, error)
	ToggleLike(ctx context.Context, commentID string, userID string) (Comment, error)
	Report(ctx context.Context, commentID string, user auth.User, request ReportRequest) error
	ListReports(ctx context.Context, status string) (ReportListResult, error)
	UpdateReportStatus(ctx context.Context, id string, status string) (CommentReport, error)
}

func countCommentStats(stats ManageStats, comment Comment) ManageStats {
	stats.Total++
	stats.Likes += comment.LikeCount
	stats.Replies += comment.ReplyCount
	switch comment.Status {
	case "approved":
		stats.Approved++
	case "pending":
		stats.Pending++
	case "rejected":
		stats.Rejected++
	case "spam":
		stats.Spam++
	case "deleted":
		stats.Deleted++
	}

	return stats
}

func matchesStatus(status string, queryStatus string) bool {
	queryStatus = strings.ToLower(strings.TrimSpace(queryStatus))
	if queryStatus == "" || queryStatus == "all" {
		return true
	}

	return status == queryStatus
}

func isValidStatus(status string) bool {
	switch status {
	case "approved", "pending", "rejected", "spam", "deleted":
		return true
	default:
		return false
	}
}

func normalizeReportStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "reviewed", "dismissed":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "pending"
	}
}

func approvedCommentDelta(previousStatus string, nextStatus string) int {
	previousApproved := normalizeStatusFilter(previousStatus) == "approved"
	nextApproved := normalizeStatusFilter(nextStatus) == "approved"
	if previousApproved == nextApproved {
		return 0
	}
	if nextApproved {
		return 1
	}

	return -1
}

func filterComments(items []Comment, query ListQuery) []Comment {
	keyword := strings.ToLower(strings.TrimSpace(query.Keyword))
	if keyword == "" {
		return items
	}

	filtered := make([]Comment, 0, len(items))
	for _, item := range items {
		if commentContainsKeyword(item, keyword) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func commentContainsKeyword(item Comment, keyword string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		item.Body,
		item.AuthorName,
		item.AuthorID,
		item.PostTitle,
		item.PostSlug,
		item.RiskLevel,
	}, " "))
	return strings.Contains(haystack, keyword)
}

func sortComments(items []Comment, mode string) {
	sort.SliceStable(items, func(i, j int) bool {
		if mode == "likes" {
			return items[i].LikeCount > items[j].LikeCount
		}
		if mode == "replies" {
			return items[i].ReplyCount > items[j].ReplyCount
		}
		if mode == "risk" {
			return commentRiskRank(items[i].RiskLevel) > commentRiskRank(items[j].RiskLevel)
		}

		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func commentRiskRank(value string) int {
	value = strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.Contains(value, "\u9ad8"), strings.Contains(value, "high"):
		return 3
	case strings.Contains(value, "\u4e2d"), strings.Contains(value, "medium"):
		return 2
	case value != "":
		return 1
	default:
		return 0
	}
}

func pagedManageResult(items []Comment, stats ManageStats, query ListQuery) ManageListResult {
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

	return ManageListResult{
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

func riskLevel(body string) string {
	lower := strings.ToLower(body)
	if strings.Count(lower, "http://")+strings.Count(lower, "https://") > 0 {
		return "高"
	}
	if len([]rune(strings.TrimSpace(body))) > 200 {
		return "中"
	}

	return "低"
}

func titleForSlug(slug string) string {
	titles := map[string]string{
		"blog-system-design":                       "如何设计一个内容长期增长的博客系统",
		"vue3-content-site-cache-seo":              "Vue3 内容站的缓存与 SEO 边界",
		"postgres-redis-blog-boundary":             "Redis 和 PostgreSQL 在博客中的分工",
		"post-version-history":                     "为什么博客后台需要文章版本历史",
		"postgres-full-text-search":                "用 PostgreSQL 做博客全文搜索够不够",
		"home-to-article-information-architecture": "从首页到文章页，博客的信息架构怎么排",
	}
	if title, ok := titles[slug]; ok {
		return title
	}

	return slug
}

func seedComments() []Comment {
	now := time.Now()
	return []Comment{
		{
			ID:         "comment_001",
			PostSlug:   "blog-system-design",
			AuthorID:   "user_chen",
			AuthorName: "陈默",
			AvatarText: "陈",
			Body:       "文章里提到“内容模型先于页面”很关键。很多博客后期难维护，就是因为一开始把文章当页面模板来处理了。",
			Status:     "approved",
			LikeCount:  18,
			CreatedAt:  now.Add(-2 * time.Hour),
		},
		{
			ID:         "comment_002",
			PostSlug:   "blog-system-design",
			ParentID:   "comment_001",
			AuthorID:   "user_admin",
			AuthorName: "管理员",
			AvatarText: "管",
			Body:       "是的，所以我会优先把 slug、SEO、状态、版本历史这些字段纳入第一版数据模型。",
			Status:     "approved",
			LikeCount:  9,
			IsAuthor:   true,
			CreatedAt:  now.Add(-1 * time.Hour),
		},
		{
			ID:         "comment_003",
			PostSlug:   "postgres-redis-blog-boundary",
			AuthorID:   "user_linyi",
			AuthorName: "林一",
			AvatarText: "林",
			Body:       "PostgreSQL 全文搜索足够覆盖多数个人博客场景。",
			Status:     "approved",
			LikeCount:  9,
			CreatedAt:  now.Add(-72 * time.Hour),
		},
	}
}
