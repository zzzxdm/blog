package submissions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"blog/api/internal/database"
	"blog/api/internal/idgen"
	"blog/api/internal/modules/auth"
)

type SQLRepository struct {
	db     *sql.DB
	sqlite bool
}

func NewSQLRepository(ctx context.Context, db *sql.DB) (*SQLRepository, error) {
	repo := &SQLRepository{db: db, sqlite: database.IsSQLite(db)}
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

	items = filterSubmissions(items, query)
	sortSubmissions(items, query.Sort)

	return pagedSubmissionResult(items, stats, query), nil
}

func (repo *SQLRepository) CountSubmittedSince(ctx context.Context, userID string, since time.Time, excludeID string) (int, error) {
	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT count(*)
			FROM submissions
			WHERE author_id = $1
				AND submitted_at >= $2
				AND ($3 = '' OR CAST(id AS TEXT) <> $3)
				AND visibility = 'public'
	`, userID, since, strings.TrimSpace(excludeID)).Scan(&total); err != nil {
		return 0, fmt.Errorf("count submitted submissions: %w", err)
	}

	return total, nil
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
		ID:           idgen.NextString(),
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
			id, author_id, title, summary, content, category, tags, cover_image, slug, visibility,
			status, review_note, version, created_at, updated_at, submitted_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, '', $12, $13, $14, $15)
	`, submission.ID,
		submission.AuthorID,
		submission.Title,
		submission.Summary,
		submission.Content,
		submission.Category,
		repo.tagsValue(submission.Tags),
		submission.CoverImage,
		submission.Slug,
		submission.Visibility,
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
	if (current.Status == StatusPublished || current.Status == StatusArchived) && (current.Visibility != VisibilityPrivate || !request.Submit) {
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
			visibility = $9,
			status = $10,
			review_note = $11,
			submitted_at = $12,
			reviewed_at = $13,
			version = version + 1
		WHERE id = $1
	`, submissionID,
		updated.Title,
		updated.Summary,
		updated.Content,
		updated.Category,
		repo.tagsValue(updated.Tags),
		updated.CoverImage,
		updated.Slug,
		updated.Visibility,
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
	if current.Status == StatusPublished || current.Status == StatusArchived {
		return Submission{}, ErrForbidden
	}
	if err := validateSubmissionReady(current); err != nil {
		return Submission{}, err
	}

	if _, err := repo.db.ExecContext(ctx, `
		UPDATE submissions
		SET status = 'submitted',
			review_note = '',
			submitted_at = $2,
			reviewed_at = NULL,
			version = version + 1
		WHERE id = $1
	`, submissionID, time.Now()); err != nil {
		return Submission{}, fmt.Errorf("submit submission: %w", err)
	}

	return repo.Get(ctx, submissionID)
}

func (repo *SQLRepository) MarkPublished(ctx context.Context, submissionID string, userID string, publishedPostSlug string) (Submission, error) {
	if strings.TrimSpace(publishedPostSlug) == "" {
		return Submission{}, ErrInvalidSubmission
	}

	current, err := repo.Get(ctx, submissionID)
	if err != nil {
		return Submission{}, err
	}
	if current.AuthorID != userID {
		return Submission{}, ErrForbidden
	}
	if err := validateSubmissionReady(current); err != nil {
		return Submission{}, err
	}

	now := time.Now()
	submittedAt := current.SubmittedAt
	if submittedAt == nil {
		submittedAt = &now
	}
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE submissions
		SET status = $2,
			review_note = '',
			published_post_slug = $3,
			submitted_at = $4,
			reviewed_at = NULL,
			published_at = $5,
			version = version + 1
		WHERE id = $1
	`, submissionID, StatusPublished, strings.TrimSpace(publishedPostSlug), submittedAt, now); err != nil {
		return Submission{}, fmt.Errorf("mark submission published: %w", err)
	}

	return repo.Get(ctx, submissionID)
}

func (repo *SQLRepository) DeleteByAuthor(ctx context.Context, submissionID string, userID string) (Submission, error) {
	current, err := repo.Get(ctx, submissionID)
	if err != nil {
		return Submission{}, err
	}
	if current.AuthorID != userID || current.Status == StatusPublished || current.Status == StatusArchived {
		return Submission{}, ErrForbidden
	}

	result, err := repo.db.ExecContext(ctx, "DELETE FROM submissions WHERE id = $1 AND author_id = $2 AND status NOT IN ($3, $4)", submissionID, userID, StatusPublished, StatusArchived)
	if err != nil {
		return Submission{}, fmt.Errorf("delete submission: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Submission{}, fmt.Errorf("delete submission rows affected: %w", err)
	}
	if affected == 0 {
		return Submission{}, ErrSubmissionNotFound
	}

	return current, nil
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

	items = filterSubmissions(items, query)
	sortSubmissions(items, query.Sort)

	return pagedSubmissionResult(items, stats, query), nil
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

func (repo *SQLRepository) AdminUpdate(ctx context.Context, submissionID string, request SaveRequest) (Submission, error) {
	if err := validateSave(request, false); err != nil {
		return Submission{}, err
	}

	current, err := repo.Get(ctx, submissionID)
	if err != nil {
		return Submission{}, err
	}
	if current.Status == StatusPublished || current.Status == StatusArchived {
		return Submission{}, ErrForbidden
	}

	updated := applySave(current, request)
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE submissions
		SET title = $2,
			summary = $3,
			content = $4,
			category = $5,
			tags = $6,
			cover_image = $7,
			slug = $8,
			visibility = $9,
			updated_at = $10,
			version = version + 1
		WHERE id = $1
	`, submissionID,
		updated.Title,
		updated.Summary,
		updated.Content,
		updated.Category,
		repo.tagsValue(updated.Tags),
		updated.CoverImage,
		updated.Slug,
		updated.Visibility,
		time.Now(),
	); err != nil {
		return Submission{}, fmt.Errorf("admin update submission: %w", err)
	}

	return repo.Get(ctx, submissionID)
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
	publishedSlug := current.PublishedPostSlug
	switch action {
	case ActionApprove:
		status = StatusPublished
		publishedAt = time.Now()
		publishedSlug = strings.TrimSpace(publishedPostSlug)
	case ActionReturn:
		status = StatusReturned
		publishedAt = current.PublishedAt
	case ActionReject:
		status = StatusRejected
		publishedAt = current.PublishedAt
	}

	if _, err := repo.db.ExecContext(ctx, `
			UPDATE submissions
			SET status = $2,
				review_note = $3,
				reviewer_id = CAST($4 AS bigint),
				reviewed_at = $7,
				published_post_slug = NULLIF($5, ''),
				published_at = $6
		WHERE id = $1
	`, submissionID, status, strings.TrimSpace(request.Note), reviewer.ID, publishedSlug, publishedAt, time.Now()); err != nil {
		return Submission{}, fmt.Errorf("review submission: %w", err)
	}

	return repo.Get(ctx, submissionID)
}

func (repo *SQLRepository) ArchivePublished(ctx context.Context, submissionID string, reviewer auth.User) (Submission, error) {
	current, err := repo.Get(ctx, submissionID)
	if err != nil {
		return Submission{}, err
	}
	if current.Status != StatusPublished || strings.TrimSpace(current.PublishedPostSlug) == "" {
		return Submission{}, ErrInvalidReview
	}

	if _, err := repo.db.ExecContext(ctx, `
			UPDATE submissions
			SET status = $2,
				review_note = $3,
				reviewer_id = CAST($4 AS bigint),
				reviewed_at = $5
			WHERE id = $1
	`, submissionID, StatusArchived, "管理员已下架该文章", reviewer.ID, time.Now()); err != nil {
		return Submission{}, fmt.Errorf("archive published submission: %w", err)
	}

	return repo.Get(ctx, submissionID)
}

func (repo *SQLRepository) RestorePublished(ctx context.Context, submissionID string, reviewer auth.User) (Submission, error) {
	current, err := repo.Get(ctx, submissionID)
	if err != nil {
		return Submission{}, err
	}
	if current.Status != StatusArchived || strings.TrimSpace(current.PublishedPostSlug) == "" {
		return Submission{}, ErrInvalidReview
	}

	now := time.Now()
	if _, err := repo.db.ExecContext(ctx, `
			UPDATE submissions
			SET status = $2,
				review_note = $3,
				reviewer_id = CAST($4 AS bigint),
				reviewed_at = $5,
				published_at = COALESCE(published_at, $5)
		WHERE id = $1
	`, submissionID, StatusPublished, "管理员已重新上架该文章", reviewer.ID, now); err != nil {
		return Submission{}, fmt.Errorf("restore published submission: %w", err)
	}

	return repo.Get(ctx, submissionID)
}

func (repo *SQLRepository) querySubmissions(ctx context.Context, whereAndOrder string, args ...any) ([]Submission, error) {
	tagsExpression := "array_to_string(s.tags, ',')"
	if repo.sqlite {
		tagsExpression = "s.tags"
	}

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
			`+tagsExpression+`,
			s.cover_image,
			s.slug,
			s.visibility,
			s.status,
			s.review_note,
				COALESCE(CAST(s.reviewer_id AS TEXT), ''),
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
		var tagsValue string
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
			&tagsValue,
			&submission.CoverImage,
			&submission.Slug,
			&submission.Visibility,
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
		submission.Tags = repo.decodeTags(tagsValue)
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
			count(*),
			COALESCE(sum(CASE WHEN status = 'draft' THEN 1 ELSE 0 END), 0),
			COALESCE(sum(CASE WHEN status = 'submitted' THEN 1 ELSE 0 END), 0),
				COALESCE(sum(CASE WHEN status = 'returned' THEN 1 ELSE 0 END), 0),
				COALESCE(sum(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END), 0),
				COALESCE(sum(CASE WHEN status = 'published' THEN 1 ELSE 0 END), 0),
				COALESCE(sum(CASE WHEN status = 'archived' THEN 1 ELSE 0 END), 0)
			FROM submissions
		` + where

	var row *sql.Row
	if where == "" {
		row = repo.db.QueryRowContext(ctx, query)
	} else {
		row = repo.db.QueryRowContext(ctx, query, arg)
	}

	var stats Stats
	if err := row.Scan(&stats.Total, &stats.Draft, &stats.Submitted, &stats.Returned, &stats.Rejected, &stats.Published, &stats.Archived); err != nil {
		return Stats{}, fmt.Errorf("scan submission stats: %w", err)
	}

	return stats, nil
}

func (repo *SQLRepository) ensureSeedSubmissions(ctx context.Context) error {
	for _, submission := range seedSubmissions() {
		if _, err := repo.db.ExecContext(ctx, `
				INSERT INTO submissions (
					id, author_id, title, summary, content, category, tags, cover_image, slug, visibility,
					status, review_note, reviewer_id, published_post_slug, version,
					created_at, updated_at, submitted_at, reviewed_at, published_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CAST(NULLIF($13, '') AS bigint), NULLIF($14, ''), $15, $16, $17, $18, $19, $20)
				ON CONFLICT (id) DO NOTHING
		`,
			submission.ID,
			submission.AuthorID,
			submission.Title,
			submission.Summary,
			submission.Content,
			submission.Category,
			repo.tagsValue(submission.Tags),
			submission.CoverImage,
			submission.Slug,
			submission.Visibility,
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

func (repo *SQLRepository) tagsValue(tags []string) any {
	if !repo.sqlite {
		return tags
	}

	data, err := json.Marshal(normalizeTags(tags))
	if err != nil {
		return "[]"
	}
	return string(data)
}

func (repo *SQLRepository) decodeTags(value string) []string {
	if repo.sqlite {
		var tags []string
		if err := json.Unmarshal([]byte(value), &tags); err == nil {
			return normalizeTags(tags)
		}
	}

	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	return normalizeTags(strings.Split(value, ","))
}

func normalizeStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}
