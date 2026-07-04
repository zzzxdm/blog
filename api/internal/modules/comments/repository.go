package comments

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"blog/api/internal/modules/auth"
)

var ErrEmptyBody = errors.New("comment body is empty")

type Repository interface {
	List(ctx context.Context, postSlug string, viewerID string) (ListResult, error)
	Create(ctx context.Context, postSlug string, request CreateRequest, user auth.User) (Comment, error)
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
		ParentID:   strings.TrimSpace(request.ParentID),
		AuthorID:   user.ID,
		AuthorName: user.DisplayName,
		AvatarText: user.AvatarText,
		Body:       body,
		Status:     "pending",
		LikeCount:  0,
		IsMine:     true,
		CreatedAt:  repo.now(),
	}
	repo.nextID++
	repo.comments = append(repo.comments, comment)

	return comment, nil
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
