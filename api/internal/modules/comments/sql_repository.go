package comments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"blog/api/internal/modules/auth"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(ctx context.Context, db *sql.DB) (*SQLRepository, error) {
	repo := &SQLRepository{db: db}
	if err := repo.ensureSeedComments(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *SQLRepository) List(ctx context.Context, postSlug string, viewerID string) (ListResult, error) {
	items, err := repo.queryComments(ctx, `
		WHERE c.post_slug = $1
			AND (c.status = 'approved' OR c.author_id = $2)
		ORDER BY c.created_at ASC
	`, postSlug, viewerID)
	if err != nil {
		return ListResult{}, err
	}

	for index := range items {
		items[index].IsMine = viewerID != "" && items[index].AuthorID == viewerID
		items[index].Liked, err = repo.likedByUser(ctx, items[index].ID, viewerID)
		if err != nil {
			return ListResult{}, err
		}
	}

	return ListResult{
		Items: items,
		Total: len(items),
	}, nil
}

func (repo *SQLRepository) Create(ctx context.Context, postSlug string, request CreateRequest, user auth.User) (Comment, error) {
	body := trimString(request.Body)
	if body == "" {
		return Comment{}, ErrEmptyBody
	}

	id := fmt.Sprintf("comment_%d", time.Now().UnixNano())
	row := repo.db.QueryRowContext(ctx, `
		INSERT INTO comments (id, post_slug, parent_id, author_id, body, status, like_count, is_author)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, 'pending', 0, $6)
		RETURNING id
	`, id, postSlug, trimString(request.ParentID), user.ID, body, user.Role == "admin" || user.Role == "author")
	if err := row.Scan(&id); err != nil {
		return Comment{}, fmt.Errorf("insert comment: %w", err)
	}

	comment, err := repo.getByID(ctx, id)
	if err != nil {
		return Comment{}, err
	}
	comment.IsMine = true

	return comment, nil
}

func (repo *SQLRepository) CreateReply(ctx context.Context, parentID string, request CreateRequest, user auth.User) (Comment, error) {
	body := trimString(request.Body)
	if body == "" {
		return Comment{}, ErrEmptyBody
	}

	var postSlug string
	if err := repo.db.QueryRowContext(ctx, "SELECT post_slug FROM comments WHERE id = $1", parentID).Scan(&postSlug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Comment{}, ErrCommentNotFound
		}
		return Comment{}, fmt.Errorf("load parent comment: %w", err)
	}

	id := fmt.Sprintf("comment_%d", time.Now().UnixNano())
	row := repo.db.QueryRowContext(ctx, `
		INSERT INTO comments (id, post_slug, parent_id, author_id, body, status, like_count, is_author)
		VALUES ($1, $2, $3, $4, $5, 'pending', 0, $6)
		RETURNING id
	`, id, postSlug, parentID, user.ID, body, user.Role == "admin" || user.Role == "author")
	if err := row.Scan(&id); err != nil {
		return Comment{}, fmt.Errorf("insert reply: %w", err)
	}

	comment, err := repo.getByID(ctx, id)
	if err != nil {
		return Comment{}, err
	}
	comment.IsMine = true

	return comment, nil
}

func (repo *SQLRepository) ListByAuthor(ctx context.Context, userID string, query ListQuery) (ManageListResult, error) {
	items, err := repo.queryComments(ctx, `
		WHERE c.author_id = $1
			AND ($2 = '' OR $2 = 'all' OR c.status = $2)
		ORDER BY c.created_at DESC
	`, userID, normalizeStatusFilter(query.Status))
	if err != nil {
		return ManageListResult{}, err
	}

	for index := range items {
		items[index].IsMine = true
	}

	stats, err := repo.stats(ctx, "WHERE c.author_id = $1", userID)
	if err != nil {
		return ManageListResult{}, err
	}

	items = filterComments(items, query)
	sortComments(items, query.Sort)

	return pagedManageResult(items, stats, query), nil
}

func (repo *SQLRepository) AdminList(ctx context.Context, query ListQuery) (ManageListResult, error) {
	items, err := repo.queryComments(ctx, `
		WHERE ($1 = '' OR $1 = 'all' OR c.status = $1)
		ORDER BY c.created_at DESC
	`, normalizeStatusFilter(query.Status))
	if err != nil {
		return ManageListResult{}, err
	}

	stats, err := repo.stats(ctx, "", nil)
	if err != nil {
		return ManageListResult{}, err
	}

	items = filterComments(items, query)
	sortComments(items, query.Sort)

	return pagedManageResult(items, stats, query), nil
}

func (repo *SQLRepository) UpdateStatus(ctx context.Context, commentID string, status string) (Comment, error) {
	status = normalizeStatusFilter(status)
	if !isValidStatus(status) {
		return Comment{}, ErrInvalidStatus
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Comment{}, fmt.Errorf("begin comment status transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var id string
	var postSlug string
	var previousStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT id, post_slug, status
		FROM comments
		WHERE id = $1
	`, commentID).Scan(&id, &postSlug, &previousStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Comment{}, ErrCommentNotFound
		}
		return Comment{}, fmt.Errorf("load comment status: %w", err)
	}

	result, err := tx.ExecContext(ctx, "UPDATE comments SET status = $2 WHERE id = $1", commentID, status)
	if err != nil {
		return Comment{}, fmt.Errorf("update comment status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Comment{}, fmt.Errorf("read comment status rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return Comment{}, ErrCommentNotFound
	}

	if delta := approvedCommentDelta(previousStatus, status); delta != 0 {
		if _, err := tx.ExecContext(ctx, `
			UPDATE posts
			SET comment_count = CASE WHEN comment_count + $2 < 0 THEN 0 ELSE comment_count + $2 END
			WHERE slug = $1
		`, postSlug, delta); err != nil {
			return Comment{}, fmt.Errorf("update post comment count: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Comment{}, fmt.Errorf("commit comment status transaction: %w", err)
	}
	committed = true

	return repo.getByID(ctx, id)
}

func (repo *SQLRepository) DeleteByAuthor(ctx context.Context, commentID string, userID string) (Comment, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Comment{}, fmt.Errorf("begin comment delete transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var id string
	var postSlug string
	var previousStatus string
	var authorID string
	err = tx.QueryRowContext(ctx, `
		SELECT id, post_slug, status, author_id
		FROM comments
		WHERE id = $1
	`, commentID).Scan(&id, &postSlug, &previousStatus, &authorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Comment{}, ErrCommentNotFound
		}
		return Comment{}, fmt.Errorf("load comment for delete: %w", err)
	}

	if authorID != userID {
		return Comment{}, ErrForbidden
	}

	result, err := tx.ExecContext(ctx, "UPDATE comments SET status = 'deleted' WHERE id = $1 AND author_id = $2", commentID, userID)
	if err != nil {
		return Comment{}, fmt.Errorf("delete comment: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Comment{}, fmt.Errorf("read delete comment rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return Comment{}, ErrCommentNotFound
	}

	if delta := approvedCommentDelta(previousStatus, "deleted"); delta != 0 {
		if _, err := tx.ExecContext(ctx, `
			UPDATE posts
			SET comment_count = CASE WHEN comment_count + $2 < 0 THEN 0 ELSE comment_count + $2 END
			WHERE slug = $1
		`, postSlug, delta); err != nil {
			return Comment{}, fmt.Errorf("update post comment count: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Comment{}, fmt.Errorf("commit comment delete transaction: %w", err)
	}
	committed = true

	return repo.getByID(ctx, id)
}

func (repo *SQLRepository) ToggleLike(ctx context.Context, commentID string, userID string) (Comment, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Comment{}, fmt.Errorf("begin comment like transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	var exists bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM comments WHERE id = $1)", commentID).Scan(&exists); err != nil {
		return Comment{}, fmt.Errorf("check comment exists: %w", err)
	}
	if !exists {
		return Comment{}, ErrCommentNotFound
	}

	var liked bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM comment_likes WHERE comment_id = $1 AND user_id = $2)", commentID, userID).Scan(&liked); err != nil {
		return Comment{}, fmt.Errorf("check comment like: %w", err)
	}

	if liked {
		if _, err := tx.ExecContext(ctx, "DELETE FROM comment_likes WHERE comment_id = $1 AND user_id = $2", commentID, userID); err != nil {
			return Comment{}, fmt.Errorf("delete comment like: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "UPDATE comments SET like_count = CASE WHEN like_count - 1 < 0 THEN 0 ELSE like_count - 1 END WHERE id = $1", commentID); err != nil {
			return Comment{}, fmt.Errorf("decrement comment like count: %w", err)
		}
	} else {
		if _, err := tx.ExecContext(ctx, "INSERT INTO comment_likes (comment_id, user_id) VALUES ($1, $2)", commentID, userID); err != nil {
			return Comment{}, fmt.Errorf("insert comment like: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "UPDATE comments SET like_count = like_count + 1 WHERE id = $1", commentID); err != nil {
			return Comment{}, fmt.Errorf("increment comment like count: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Comment{}, fmt.Errorf("commit comment like transaction: %w", err)
	}
	tx = nil

	comment, err := repo.getByID(ctx, commentID)
	if err != nil {
		return Comment{}, err
	}
	comment.IsMine = comment.AuthorID == userID
	comment.Liked = !liked

	return comment, nil
}

func (repo *SQLRepository) Report(ctx context.Context, commentID string, user auth.User, request ReportRequest) error {
	var exists bool
	if err := repo.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM comments WHERE id = $1)", commentID).Scan(&exists); err != nil {
		return fmt.Errorf("check comment exists: %w", err)
	}
	if !exists {
		return ErrCommentNotFound
	}

	reason := trimString(request.Reason)
	if reason == "" {
		reason = "用户举报"
	}

	id := fmt.Sprintf("comment_report_%d", time.Now().UnixNano())
	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO comment_reports (id, comment_id, reporter_id, reason, status)
		VALUES ($1, $2, $3, $4, 'pending')
		ON CONFLICT (comment_id, reporter_id)
		DO UPDATE SET reason = EXCLUDED.reason, status = 'pending'
	`, id, commentID, user.ID, reason); err != nil {
		return fmt.Errorf("upsert comment report: %w", err)
	}

	return nil
}

func (repo *SQLRepository) ListReports(ctx context.Context, status string) (ReportListResult, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	query := `
		SELECT id, comment_id, reporter_id, reason, status, created_at
		FROM comment_reports
	`
	args := []any{}
	if status != "" && status != "all" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC, id DESC"

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return ReportListResult{}, fmt.Errorf("query comment reports: %w", err)
	}
	defer rows.Close()

	items := make([]CommentReport, 0)
	for rows.Next() {
		var report CommentReport
		if err := rows.Scan(&report.ID, &report.CommentID, &report.ReporterID, &report.Reason, &report.Status, &report.CreatedAt); err != nil {
			return ReportListResult{}, fmt.Errorf("scan comment report: %w", err)
		}
		items = append(items, report)
	}
	if err := rows.Err(); err != nil {
		return ReportListResult{}, fmt.Errorf("iterate comment reports: %w", err)
	}

	return ReportListResult{Items: items, Total: len(items)}, nil
}

func (repo *SQLRepository) UpdateReportStatus(ctx context.Context, id string, status string) (CommentReport, error) {
	status = normalizeReportStatus(status)
	var report CommentReport
	err := repo.db.QueryRowContext(ctx, `
		UPDATE comment_reports
		SET status = $2
		WHERE id = $1
		RETURNING id, comment_id, reporter_id, reason, status, created_at
	`, id, status).Scan(&report.ID, &report.CommentID, &report.ReporterID, &report.Reason, &report.Status, &report.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CommentReport{}, ErrCommentNotFound
		}
		return CommentReport{}, fmt.Errorf("update comment report: %w", err)
	}

	return report, nil
}

func (repo *SQLRepository) getByID(ctx context.Context, id string) (Comment, error) {
	items, err := repo.queryComments(ctx, `WHERE c.id = $1`, id)
	if err != nil {
		return Comment{}, err
	}
	if len(items) == 0 {
		return Comment{}, ErrCommentNotFound
	}

	return items[0], nil
}

func (repo *SQLRepository) likedByUser(ctx context.Context, commentID string, userID string) (bool, error) {
	if userID == "" {
		return false, nil
	}

	var liked bool
	if err := repo.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM comment_likes
			WHERE comment_id = $1 AND user_id = $2
		)
	`, commentID, userID).Scan(&liked); err != nil {
		return false, fmt.Errorf("check comment liked: %w", err)
	}

	return liked, nil
}

func (repo *SQLRepository) queryComments(ctx context.Context, whereAndOrder string, args ...any) ([]Comment, error) {
	query := `
		SELECT
			c.id,
			c.post_slug,
			p.title,
			COALESCE(c.parent_id, ''),
			u.id,
			u.display_name,
			u.avatar_text,
			c.body,
			c.status,
			c.like_count,
			c.is_author,
			c.created_at,
			(
				SELECT count(*)
				FROM comments replies
				WHERE replies.parent_id = c.id
			) AS reply_count
		FROM comments c
		JOIN users u ON u.id = c.author_id
		JOIN posts p ON p.slug = c.post_slug
		` + whereAndOrder

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query comments: %w", err)
	}
	defer rows.Close()

	items := make([]Comment, 0)
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.PostSlug,
			&comment.PostTitle,
			&comment.ParentID,
			&comment.AuthorID,
			&comment.AuthorName,
			&comment.AvatarText,
			&comment.Body,
			&comment.Status,
			&comment.LikeCount,
			&comment.IsAuthor,
			&comment.CreatedAt,
			&comment.ReplyCount,
		); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comment.RiskLevel = riskLevel(comment.Body)
		items = append(items, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}

	return items, nil
}

func (repo *SQLRepository) stats(ctx context.Context, where string, arg any) (ManageStats, error) {
	query := `
		SELECT
			count(*),
			COALESCE(sum(CASE WHEN c.status = 'pending' THEN 1 ELSE 0 END), 0),
			COALESCE(sum(CASE WHEN c.status = 'approved' THEN 1 ELSE 0 END), 0),
			COALESCE(sum(CASE WHEN c.status = 'rejected' THEN 1 ELSE 0 END), 0),
			COALESCE(sum(CASE WHEN c.status = 'spam' THEN 1 ELSE 0 END), 0),
			COALESCE(sum(CASE WHEN c.status = 'deleted' THEN 1 ELSE 0 END), 0),
			COALESCE(sum(c.like_count), 0)
		FROM comments c
		` + where

	var row *sql.Row
	if where == "" {
		row = repo.db.QueryRowContext(ctx, query)
	} else {
		row = repo.db.QueryRowContext(ctx, query, arg)
	}

	var stats ManageStats
	if err := row.Scan(&stats.Total, &stats.Pending, &stats.Approved, &stats.Rejected, &stats.Spam, &stats.Deleted, &stats.Likes); err != nil {
		return ManageStats{}, fmt.Errorf("scan comment stats: %w", err)
	}

	repliesQuery := `
		SELECT count(*)
		FROM comments replies
		JOIN comments c ON c.id = replies.parent_id
		` + where
	if where == "" {
		if err := repo.db.QueryRowContext(ctx, repliesQuery).Scan(&stats.Replies); err != nil {
			return ManageStats{}, fmt.Errorf("scan reply stats: %w", err)
		}
	} else if err := repo.db.QueryRowContext(ctx, repliesQuery, arg).Scan(&stats.Replies); err != nil {
		return ManageStats{}, fmt.Errorf("scan reply stats: %w", err)
	}

	return stats, nil
}

func (repo *SQLRepository) ensureSeedComments(ctx context.Context) error {
	seeds := seedComments()
	for _, comment := range seeds {
		if _, err := repo.db.ExecContext(ctx, `
			INSERT INTO comments (id, post_slug, parent_id, author_id, body, status, like_count, is_author, created_at)
			VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO NOTHING
		`, comment.ID, comment.PostSlug, comment.ParentID, comment.AuthorID, comment.Body, comment.Status, comment.LikeCount, comment.IsAuthor, comment.CreatedAt); err != nil {
			return fmt.Errorf("seed comment %s: %w", comment.ID, err)
		}
	}

	return nil
}

func normalizeStatusFilter(status string) string {
	return strings.ToLower(trimString(status))
}

func trimString(value string) string {
	return strings.TrimSpace(value)
}
