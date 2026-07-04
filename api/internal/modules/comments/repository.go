package comments

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"blog/api/internal/modules/auth"
)

var ErrEmptyBody = errors.New("comment body is empty")
var ErrCommentNotFound = errors.New("comment not found")
var ErrInvalidStatus = errors.New("invalid comment status")

type Repository interface {
	List(ctx context.Context, postSlug string, viewerID string) (ListResult, error)
	Create(ctx context.Context, postSlug string, request CreateRequest, user auth.User) (Comment, error)
	ListByAuthor(ctx context.Context, userID string, query ListQuery) (ManageListResult, error)
	AdminList(ctx context.Context, query ListQuery) (ManageListResult, error)
	UpdateStatus(ctx context.Context, commentID string, status string) (Comment, error)
}

type MemoryRepository struct {
	mu       sync.RWMutex
	comments []Comment
	nextID   int
	now      func() time.Time
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		comments: seedComments(),
		nextID:   100,
		now:      time.Now,
	}
}

func (repo *MemoryRepository) List(_ context.Context, postSlug string, viewerID string) (ListResult, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	items := make([]Comment, 0)
	for _, comment := range repo.comments {
		if comment.PostSlug != postSlug {
			continue
		}

		if comment.Status != "approved" && comment.AuthorID != viewerID {
			continue
		}

		item := comment
		item = repo.enrichLocked(item)
		item.IsMine = viewerID != "" && comment.AuthorID == viewerID
		items = append(items, item)
	}

	return ListResult{
		Items: items,
		Total: len(items),
	}, nil
}

func (repo *MemoryRepository) Create(_ context.Context, postSlug string, request CreateRequest, user auth.User) (Comment, error) {
	body := strings.TrimSpace(request.Body)
	if body == "" {
		return Comment{}, ErrEmptyBody
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	comment := Comment{
		ID:         fmt.Sprintf("comment_%03d", repo.nextID),
		PostSlug:   postSlug,
		PostTitle:  titleForSlug(postSlug),
		ParentID:   strings.TrimSpace(request.ParentID),
		AuthorID:   user.ID,
		AuthorName: user.DisplayName,
		AvatarText: user.AvatarText,
		Body:       body,
		Status:     "pending",
		LikeCount:  0,
		RiskLevel:  riskLevel(body),
		IsMine:     true,
		CreatedAt:  repo.now(),
	}
	repo.nextID++
	repo.comments = append(repo.comments, comment)

	return repo.enrichLocked(comment), nil
}

func (repo *MemoryRepository) ListByAuthor(_ context.Context, userID string, query ListQuery) (ManageListResult, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	items := make([]Comment, 0)
	for _, comment := range repo.comments {
		if comment.AuthorID != userID || !matchesStatus(comment.Status, query.Status) {
			continue
		}

		item := repo.enrichLocked(comment)
		item.IsMine = true
		items = append(items, item)
	}

	sortComments(items)

	return ManageListResult{
		Items: items,
		Total: len(items),
		Stats: repo.statsByAuthorLocked(userID),
	}, nil
}

func (repo *MemoryRepository) AdminList(_ context.Context, query ListQuery) (ManageListResult, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	items := make([]Comment, 0, len(repo.comments))
	for _, comment := range repo.comments {
		if !matchesStatus(comment.Status, query.Status) {
			continue
		}

		items = append(items, repo.enrichLocked(comment))
	}

	sortComments(items)

	return ManageListResult{
		Items: items,
		Total: len(items),
		Stats: repo.adminStatsLocked(),
	}, nil
}

func (repo *MemoryRepository) UpdateStatus(_ context.Context, commentID string, status string) (Comment, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if !isValidStatus(status) {
		return Comment{}, ErrInvalidStatus
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	for index := range repo.comments {
		if repo.comments[index].ID != commentID {
			continue
		}

		repo.comments[index].Status = status
		return repo.enrichLocked(repo.comments[index]), nil
	}

	return Comment{}, ErrCommentNotFound
}

func (repo *MemoryRepository) enrichLocked(comment Comment) Comment {
	comment.PostTitle = titleForSlug(comment.PostSlug)
	comment.RiskLevel = riskLevel(comment.Body)
	comment.ReplyCount = 0
	for _, item := range repo.comments {
		if item.ParentID == comment.ID {
			comment.ReplyCount++
		}
	}

	return comment
}

func (repo *MemoryRepository) statsByAuthorLocked(userID string) ManageStats {
	stats := ManageStats{}
	for _, comment := range repo.comments {
		if comment.AuthorID != userID {
			continue
		}
		stats = countCommentStats(stats, repo.enrichLocked(comment))
	}

	return stats
}

func (repo *MemoryRepository) adminStatsLocked() ManageStats {
	stats := ManageStats{}
	for _, comment := range repo.comments {
		stats = countCommentStats(stats, repo.enrichLocked(comment))
	}

	return stats
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

func sortComments(items []Comment) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
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
