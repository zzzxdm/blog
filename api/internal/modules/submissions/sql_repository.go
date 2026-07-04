package submissions

import (
	"context"
	"database/sql"
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
	if err := repo.ensureSeedSubmissions(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *SQLRepository) ListByAuthor(ctx context.Context, userID string, query ListQuery) (ListResult, error) {
	items, err := repo.querySubmissions(ctx, `
		WHERE s.author_id = $1
			AND ($2 = '' OR $2 = 'all' OR s.status = $2)
		ORDER BY COALESCE(s.submitted_at, s.updated_at) DESC
	`, userID, normalizeStatus(query.Status))
	if err != nil {
		return ListResult{}, err
	}

	stats, err := repo.stats(ctx, "WHERE author_id = $1", userID)
	if err != nil {
		return ListResult{}, err
	}

	return ListResult{
		Items: items,
		Total: len(items),
		Stats: stats,
	}, nil
}

func (repo *SQLRepository) Create(ctx context.Context, request SaveRequest, user auth.User) (Submission, error) {
	if err := validateSave(request, request.Submit); err != nil {
		return Submission{}, err
	}

	now := time.Now()
	status := StatusDraft
	var submittedAt *time.Time
	if request.Submit {
		status = StatusSubmitted
		submittedAt = &now
	}

	submission := applySave(Submission{
		ID:           fmt.Sprintf("submission_%d", now.UnixNano()),
		AuthorID:     user.ID,
		AuthorName:   user.DisplayName,
		AuthorAvatar: user.AvatarText,
		Status:       status,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
		SubmittedAt:  submittedAt,
	}, request)

	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO submissions (
			id, author_id, title, summary, content, category, tags, cover_image, slug,
			status, review_note, version, created_at, updated_at, submitted_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, '', $11, $12, $13, $14)
	`, submission.ID,
		submission.AuthorID,
		submission.Title,
		submission.Summary,
		submission.Content,
		submission.Category,
		submission.Tags,
		submission.CoverImage,
		submission.Slug,
		submission.Status,
		submission.Version,
		submission.CreatedAt,
		submission.UpdatedAt,
		submission.SubmittedAt,
	); err != nil {
		return Submission{}, fmt.Errorf("insert submission: %w", err)
	}

	return repo.Get(ctx, submission.ID)
}

func (repo *SQLRepository) Update(ctx context.Context, submissionID string, userID string, request SaveRequest) (Submission, error) {
	if err := validateSave(request, request.Submit); err != nil {
		return Submission{}, err
	}

	current, err := repo.Get(ctx, submissionID)
	if err != nil {
		return Submission{}, err
	}
	if current.AuthorID != userID {
		return Submission{}, ErrForbidden
	}
	if current.Status == StatusPublished {
		return Submission{}, ErrForbidden
	}

	updated := applySave(current, request)
	status := current.Status
	reviewNote := current.ReviewNote
	var reviewedAt any = current.ReviewedAt
	var submittedAt any = current.SubmittedAt
	if request.Submit {
		status = StatusSubmitted
		reviewNote = ""
		reviewedAt = nil
		submittedAt = time.Now()
	}

	if _, err := repo.db.ExecContext(ctx, `
		UPDATE submissions
		SET title = $2,
			summary = $3,
			content = $4,
			category = $5,
			tags = $6,
			cover_image = $7,
			slug = $8,
			status = $9,
			review_note = $10,
			submitted_at = $11,
			reviewed_at = $12,
			version = version + 1
		WHERE id = $1
	`, submissionID,
		updated.Title,
		updated.Summary,
		updated.Content,
		updated.Category,
		updated.Tags,
		updated.CoverImage,
		updated.Slug,
		status,
		reviewNote,
		submittedAt,
		reviewedAt,
	); err != nil {
		return Submission{}, fmt.Errorf("update submission: %w", err)
	}

	return repo.Get(ctx, submissionID)
}

func (repo *SQLRepository) Submit(ctx context.Context, submissionID string, userID string) (Submission, error) {
	current, err := repo.Get(ctx, submissionID)
	if err != nil {
		return Submission{}, err
	}
	if current.AuthorID != userID {
		return Submission{}, ErrForbidden
	}
	if current.Status == StatusPublished {
		return Submission{}, ErrForbidden
	}
	if err := validateSubmissionReady(current); err != nil {
		return Submission{}, err
	}

	if _, err := repo.db.ExecContext(ctx, `
		UPDATE submissions
		SET status = 'submitted',
			review_note = '',
			submitted_at = now(),
			reviewed_at = NULL,
			version = version + 1
		WHERE id = $1
	`, submissionID); err != nil {
		return Submission{}, fmt.Errorf("submit submission: %w", err)
	}

	return repo.Get(ctx, submissionID)
}

func (repo *SQLRepository) AdminList(ctx context.Context, query ListQuery) (ListResult, error) {
	items, err := repo.querySubmissions(ctx, `
		WHERE ($1 = '' OR $1 = 'all' OR s.status = $1)
		ORDER BY COALESCE(s.submitted_at, s.updated_at) DESC
	`, normalizeStatus(query.Status))
	if err != nil {
		return ListResult{}, err
	}

	stats, err := repo.stats(ctx, "", nil)
	if err != nil {
		return ListResult{}, err
	}

	return ListResult{
		Items: items,
		Total: len(items),
		Stats: stats,
	}, nil
}

func (repo *SQLRepository) Get(ctx context.Context, submissionID string) (Submission, error) {
	items, err := repo.querySubmissions(ctx, `WHERE s.id = $1`, submissionID)
	if err != nil {
		return Submission{}, err
	}
	if len(items) == 0 {
		return Submission{}, ErrSubmissionNotFound
	}

	return items[0], nil
}

func (repo *SQLRepository) Review(ctx context.Context, submissionID string, reviewer auth.User, request ReviewRequest, publishedPostSlug string) (Submission, error) {
	action := strings.ToLower(strings.TrimSpace(request.Action))
	if action != ActionApprove && action != ActionReturn && action != ActionReject {
		return Submission{}, ErrInvalidReview
	}
	if action == ActionApprove && strings.TrimSpace(publishedPostSlug) == "" {
		return Submission{}, ErrInvalidReview
	}

	current, err := repo.Get(ctx, submissionID)
	if err != nil {
		return Submission{}, err
	}
	if current.Status != StatusSubmitted && current.Status != StatusReturned {
		return Submission{}, ErrInvalidReview
	}

	status := StatusRejected
	var publishedAt any
	publishedSlug := ""
	switch action {
	case ActionApprove:
		status = StatusPublished
		publishedAt = time.Now()
		publishedSlug = strings.TrimSpace(publishedPostSlug)
	case ActionReturn:
		status = StatusReturned
	case ActionReject:
		status = StatusRejected
	}

	if _, err := repo.db.ExecContext(ctx, `
		UPDATE submissions
		SET status = $2,
			review_note = $3,
			reviewer_id = $4,
			reviewed_at = now(),
			published_post_slug = NULLIF($5, ''),
			published_at = $6
		WHERE id = $1
	`, submissionID, status, strings.TrimSpace(request.Note), reviewer.ID, publishedSlug, publishedAt); err != nil {
		return Submission{}, fmt.Errorf("review submission: %w", err)
	}

	return repo.Get(ctx, submissionID)
}

func (repo *SQLRepository) querySubmissions(ctx context.Context, whereAndOrder string, args ...any) ([]Submission, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			s.id,
			s.author_id,
			author.display_name,
			author.avatar_text,
			s.title,
			s.summary,
			s.content,
			s.category,
			s.tags,
			s.cover_image,
			s.slug,
			s.status,
			s.review_note,
			COALESCE(s.reviewer_id, ''),
			COALESCE(reviewer.display_name, ''),
			COALESCE(s.published_post_slug, ''),
			s.version,
			s.created_at,
			s.updated_at,
			s.submitted_at,
			s.reviewed_at,
			s.published_at
		FROM submissions s
		JOIN users author ON author.id = s.author_id
		LEFT JOIN users reviewer ON reviewer.id = s.reviewer_id
		`+whereAndOrder, args...)
	if err != nil {
		return nil, fmt.Errorf("query submissions: %w", err)
	}
	defer rows.Close()

	items := make([]Submission, 0)
	for rows.Next() {
		var submission Submission
		var submittedAt sql.NullTime
		var reviewedAt sql.NullTime
		var publishedAt sql.NullTime
		if err := rows.Scan(
			&submission.ID,
			&submission.AuthorID,
			&submission.AuthorName,
			&submission.AuthorAvatar,
			&submission.Title,
			&submission.Summary,
			&submission.Content,
			&submission.Category,
			&submission.Tags,
			&submission.CoverImage,
			&submission.Slug,
			&submission.Status,
			&submission.ReviewNote,
			&submission.ReviewerID,
			&submission.ReviewerName,
			&submission.PublishedPostSlug,
			&submission.Version,
			&submission.CreatedAt,
			&submission.UpdatedAt,
			&submittedAt,
			&reviewedAt,
			&publishedAt,
		); err != nil {
			return nil, fmt.Errorf("scan submission: %w", err)
		}
		if submittedAt.Valid {
			submission.SubmittedAt = &submittedAt.Time
		}
		if reviewedAt.Valid {
			submission.ReviewedAt = &reviewedAt.Time
		}
		if publishedAt.Valid {
			submission.PublishedAt = &publishedAt.Time
		}
		items = append(items, normalizeSubmission(submission))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate submissions: %w", err)
	}

	return items, nil
}

func (repo *SQLRepository) stats(ctx context.Context, where string, arg any) (Stats, error) {
	query := `
		SELECT
			count(*)::int,
			count(*) FILTER (WHERE status = 'draft')::int,
			count(*) FILTER (WHERE status = 'submitted')::int,
			count(*) FILTER (WHERE status = 'returned')::int,
			count(*) FILTER (WHERE status = 'rejected')::int,
			count(*) FILTER (WHERE status = 'published')::int
		FROM submissions
		` + where

	var row *sql.Row
	if where == "" {
		row = repo.db.QueryRowContext(ctx, query)
	} else {
		row = repo.db.QueryRowContext(ctx, query, arg)
	}

	var stats Stats
	if err := row.Scan(&stats.Total, &stats.Draft, &stats.Submitted, &stats.Returned, &stats.Rejected, &stats.Published); err != nil {
		return Stats{}, fmt.Errorf("scan submission stats: %w", err)
	}

	return stats, nil
}

func (repo *SQLRepository) ensureSeedSubmissions(ctx context.Context) error {
	for _, submission := range seedSubmissions() {
		if _, err := repo.db.ExecContext(ctx, `
			INSERT INTO submissions (
				id, author_id, title, summary, content, category, tags, cover_image, slug,
				status, review_note, reviewer_id, published_post_slug, version,
				created_at, updated_at, submitted_at, reviewed_at, published_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NULLIF($12, ''), NULLIF($13, ''), $14, $15, $16, $17, $18, $19)
			ON CONFLICT (id) DO NOTHING
		`,
			submission.ID,
			submission.AuthorID,
			submission.Title,
			submission.Summary,
			submission.Content,
			submission.Category,
			submission.Tags,
			submission.CoverImage,
			submission.Slug,
			submission.Status,
			submission.ReviewNote,
			submission.ReviewerID,
			submission.PublishedPostSlug,
			submission.Version,
			submission.CreatedAt,
			submission.UpdatedAt,
			submission.SubmittedAt,
			submission.ReviewedAt,
			submission.PublishedAt,
		); err != nil {
			return fmt.Errorf("seed submission %s: %w", submission.ID, err)
		}
	}

	return nil
}

func normalizeStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}
