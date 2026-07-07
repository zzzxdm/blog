package submissions

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
)

var (
	ErrSubmissionNotFound = errors.New("submission not found")
	ErrForbidden          = errors.New("submission forbidden")
	ErrInvalidSubmission  = errors.New("invalid submission")
	ErrInvalidReview      = errors.New("invalid review")
)

type Repository interface {
	ListByAuthor(ctx context.Context, userID string, query ListQuery) (ListResult, error)
	CountSubmittedSince(ctx context.Context, userID string, since time.Time, excludeID string) (int, error)
	Create(ctx context.Context, request SaveRequest, user auth.User) (Submission, error)
	Update(ctx context.Context, submissionID string, userID string, request SaveRequest) (Submission, error)
	Submit(ctx context.Context, submissionID string, userID string) (Submission, error)
	DeleteByAuthor(ctx context.Context, submissionID string, userID string) (Submission, error)
	AdminList(ctx context.Context, query ListQuery) (ListResult, error)
	Get(ctx context.Context, submissionID string) (Submission, error)
	AdminUpdate(ctx context.Context, submissionID string, request SaveRequest) (Submission, error)
	Review(ctx context.Context, submissionID string, reviewer auth.User, request ReviewRequest, publishedPostSlug string) (Submission, error)
}

func countStatus(stats Stats, status string) Stats {
	stats.Total++
	switch status {
	case StatusDraft:
		stats.Draft++
	case StatusSubmitted:
		stats.Submitted++
	case StatusReturned:
		stats.Returned++
	case StatusRejected:
		stats.Rejected++
	case StatusPublished:
		stats.Published++
	}

	return stats
}

func pagedSubmissionResult(items []Submission, stats Stats, query ListQuery) ListResult {
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

func applySave(submission Submission, request SaveRequest) Submission {
	submission.Title = strings.TrimSpace(request.Title)
	submission.Summary = strings.TrimSpace(request.Summary)
	submission.Content = strings.TrimSpace(request.Content)
	submission.Category = defaultString(strings.TrimSpace(request.Category), "工程实践")
	submission.Tags = normalizeTags(request.Tags)
	submission.CoverImage = defaultString(strings.TrimSpace(request.CoverImage), "https://images.unsplash.com/photo-1519389950473-47ba0277781c?auto=format&fit=crop&w=1400&q=80")
	submission.Slug = strings.TrimSpace(request.Slug)

	return normalizeSubmission(submission)
}

func normalizeSubmission(submission Submission) Submission {
	submission.Tags = append([]string{}, submission.Tags...)
	submission.WordCount = len([]rune(strings.TrimSpace(submission.Content)))
	submission.RiskLevel = riskLevel(submission.Content)
	return submission
}

func validateSave(request SaveRequest, submit bool) error {
	if strings.TrimSpace(request.Title) == "" {
		return ErrInvalidSubmission
	}
	if submit && strings.TrimSpace(request.Content) == "" {
		return ErrInvalidSubmission
	}

	return nil
}

func validateSubmissionReady(submission Submission) error {
	if strings.TrimSpace(submission.Title) == "" || strings.TrimSpace(submission.Content) == "" {
		return ErrInvalidSubmission
	}

	return nil
}

func matchesStatus(status string, queryStatus string) bool {
	queryStatus = strings.ToLower(strings.TrimSpace(queryStatus))
	if queryStatus == "" || queryStatus == "all" {
		return true
	}

	return status == queryStatus
}

func filterSubmissions(items []Submission, query ListQuery) []Submission {
	keyword := strings.ToLower(strings.TrimSpace(query.Keyword))
	if keyword == "" {
		return items
	}

	filtered := make([]Submission, 0, len(items))
	for _, item := range items {
		if submissionContainsKeyword(item, keyword) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func submissionContainsKeyword(item Submission, keyword string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		item.Title,
		item.Summary,
		item.AuthorName,
		item.AuthorID,
		item.Category,
		item.Slug,
		item.RiskLevel,
		strings.Join(item.Tags, " "),
	}, " "))
	return strings.Contains(haystack, keyword)
}

func sortSubmissions(items []Submission, mode string) {
	sort.SliceStable(items, func(i, j int) bool {
		if mode == "risk" {
			return riskRank(items[i].RiskLevel) > riskRank(items[j].RiskLevel)
		}
		if mode == "quality" {
			return items[i].WordCount > items[j].WordCount
		}

		left := items[i].UpdatedAt
		right := items[j].UpdatedAt
		if items[i].SubmittedAt != nil {
			left = *items[i].SubmittedAt
		}
		if items[j].SubmittedAt != nil {
			right = *items[j].SubmittedAt
		}

		return left.After(right)
	})
}

func riskRank(value string) int {
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

func riskLevel(content string) string {
	if strings.Count(strings.ToLower(content), "http://")+strings.Count(strings.ToLower(content), "https://") >= 2 {
		return "中"
	}

	return "低"
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func seedSubmissions() []Submission {
	now := time.Now()
	submittedToday := now.Add(-5 * time.Hour)
	submittedYesterday := now.Add(-28 * time.Hour)
	reviewedYesterday := now.Add(-22 * time.Hour)
	publishedAt := now.AddDate(0, 0, -16)

	return []Submission{
		normalizeSubmission(Submission{
			ID:           "submission_001",
			AuthorID:     "user_linyi",
			AuthorName:   "林一",
			AuthorAvatar: "林",
			Title:        "用户评论系统应该怎么设计",
			Summary:      "从登录用户评论、审核、举报、通知和禁言机制出发，设计一个可维护的评论系统。",
			Content:      "登录用户评论、审核、举报、通知和禁言机制，是开放内容站点的基础能力。\n\n评论需要和用户系统、通知系统、反垃圾策略一起设计。",
			Category:     "用户系统",
			Tags:         []string{"评论", "用户系统", "审核"},
			CoverImage:   "https://images.unsplash.com/photo-1519389950473-47ba0277781c?auto=format&fit=crop&w=700&q=80",
			Slug:         "user-comment-system-design",
			Status:       StatusDraft,
			Version:      1,
			CreatedAt:    now.Add(-2 * time.Hour),
			UpdatedAt:    now.Add(-20 * time.Minute),
		}),
		normalizeSubmission(Submission{
			ID:           "submission_002",
			AuthorID:     "user_linyi",
			AuthorName:   "林一",
			AuthorAvatar: "林",
			Title:        "开放投稿入口后如何做反垃圾",
			Summary:      "开放投稿后，需要从注册、提交、审核和站内信反馈几个环节降低垃圾内容风险。",
			Content:      "开放投稿入口后，反垃圾策略不能只依赖审核按钮。注册门槛、提交频率、链接数量、敏感词和用户历史都需要参与风险判断。",
			Category:     "内容治理",
			Tags:         []string{"投稿", "反垃圾", "内容治理"},
			CoverImage:   "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=700&q=80",
			Slug:         "submission-anti-spam",
			Status:       StatusSubmitted,
			Version:      2,
			CreatedAt:    now.Add(-30 * time.Hour),
			UpdatedAt:    submittedToday,
			SubmittedAt:  &submittedToday,
		}),
		normalizeSubmission(Submission{
			ID:           "submission_003",
			AuthorID:     "user_linyi",
			AuthorName:   "林一",
			AuthorAvatar: "林",
			Title:        "如何写一篇可维护的技术文章",
			Summary:      "可维护的技术文章应该清楚说明问题、上下文、约束和可复用结论。",
			Content:      "一篇技术文章需要解释问题背景、方案取舍、关键实现和限制。只有代码片段不够，读者需要知道它为什么这样写。",
			Category:     "写作工作流",
			Tags:         []string{"Markdown", "写作工作流"},
			CoverImage:   "https://images.unsplash.com/photo-1499750310107-5fef28a66643?auto=format&fit=crop&w=700&q=80",
			Slug:         "maintainable-technical-writing",
			Status:       StatusReturned,
			ReviewNote:   "摘要过短，建议明确文章解决的问题；正文中有一段代码没有解释上下文；封面图缺少 alt 文本。",
			ReviewerID:   "user_admin",
			ReviewerName: "管理员",
			Version:      2,
			CreatedAt:    now.Add(-3 * 24 * time.Hour),
			UpdatedAt:    reviewedYesterday,
			SubmittedAt:  &submittedYesterday,
			ReviewedAt:   &reviewedYesterday,
		}),
		normalizeSubmission(Submission{
			ID:                "submission_004",
			AuthorID:          "user_linyi",
			AuthorName:        "林一",
			AuthorAvatar:      "林",
			Title:             "版本历史如何保护内容资产",
			Summary:           "版本记录可以保护长期运营中的内容资产，降低误改和误删风险。",
			Content:           "文章会被持续修订，后台需要记录版本历史、修改人、变更摘要和回滚能力。",
			Category:          "内容治理",
			Tags:              []string{"版本历史", "内容治理"},
			CoverImage:        "https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=700&q=80",
			Slug:              "post-version-history",
			Status:            StatusPublished,
			ReviewNote:        "审核通过",
			ReviewerID:        "user_admin",
			ReviewerName:      "管理员",
			PublishedPostSlug: "post-version-history",
			Version:           1,
			CreatedAt:         now.AddDate(0, 0, -18),
			UpdatedAt:         publishedAt,
			SubmittedAt:       &publishedAt,
			ReviewedAt:        &publishedAt,
			PublishedAt:       &publishedAt,
		}),
		normalizeSubmission(Submission{
			ID:           "submission_005",
			AuthorID:     "user_chen",
			AuthorName:   "陈默",
			AuthorAvatar: "陈",
			Title:        "从读者路径看博客首页",
			Summary:      "首页不是文章堆叠，而是把最新、专题和归档入口组织成读者路径。",
			Content:      "博客首页需要承担发现、分流和返回阅读的职责。读者从首页进入文章详情后，应该能通过返回按钮或面包屑回到上下文。",
			Category:     "产品设计",
			Tags:         []string{"信息架构", "首页"},
			CoverImage:   "https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=700&q=80",
			Slug:         "blog-home-reader-path",
			Status:       StatusSubmitted,
			Version:      1,
			CreatedAt:    now.Add(-8 * time.Hour),
			UpdatedAt:    now.Add(-7 * time.Hour),
			SubmittedAt:  &now,
		}),
	}
}
