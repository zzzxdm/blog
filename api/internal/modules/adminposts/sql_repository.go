package adminposts

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
	"blog/api/internal/modules/posts"
)

type SQLRepository struct {
	db     *sql.DB
	now    func() time.Time
	sqlite bool
}

func NewSQLRepository(ctx context.Context, db *sql.DB) (*SQLRepository, error) {
	repo := &SQLRepository{
		db:     db,
		now:    time.Now,
		sqlite: database.IsSQLite(db),
	}
	if err := repo.ensureSeedPosts(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *SQLRepository) List(ctx context.Context, query ListQuery) (ListResult, error) {
	publishedSlugExpr := "admin_posts.data->>'publishedPostSlug'"
	if repo.sqlite {
		publishedSlugExpr = "json_extract(admin_posts.data, '$.publishedPostSlug')"
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			admin_posts.data,
			posts.view_count,
			posts.comment_count
		FROM admin_posts
		LEFT JOIN posts
			ON posts.slug = COALESCE(NULLIF(`+publishedSlugExpr+`, ''), admin_posts.slug)
			AND posts.status = 'published'
		ORDER BY admin_posts.updated_at DESC, admin_posts.id DESC
	`)
	if err != nil {
		return ListResult{}, fmt.Errorf("query admin posts: %w", err)
	}
	defer rows.Close()

	items := make([]AdminPost, 0)
	for rows.Next() {
		item, err := scanAdminPostListItem(rows)
		if err != nil {
			return ListResult{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ListResult{}, fmt.Errorf("iterate admin posts: %w", err)
	}

	stats := countStats(items)
	items = filterAdminPosts(items, query)
	sortAdminPosts(items, query.Sort)

	return pagedPostResult(items, stats, query), nil
}

func (repo *SQLRepository) Get(ctx context.Context, id string) (AdminPost, error) {
	var data []byte
	err := repo.db.QueryRowContext(ctx, "SELECT data FROM admin_posts WHERE id = $1", id).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return AdminPost{}, ErrPostNotFound
		}
		return AdminPost{}, fmt.Errorf("load admin post: %w", err)
	}

	return decodeAdminPost(data)
}

func (repo *SQLRepository) Save(ctx context.Context, id string, request SaveRequest, actor auth.User) (AdminPost, error) {
	title := strings.TrimSpace(request.Title)
	content := strings.TrimSpace(request.Content)
	if title == "" {
		return AdminPost{}, ErrInvalidPost
	}
	scheduledAt, err := parseScheduledAt(request.ScheduledAt)
	if err != nil {
		return AdminPost{}, err
	}

	now := repo.now()
	item := AdminPost{
		ID:         idgen.NextString(),
		AuthorName: "管理员",
		Status:     StatusDraft,
		Visibility: VisibilityPublic,
		Version:    0,
		UpdatedAt:  now,
	}

	var previous AdminPost
	isCreate := strings.TrimSpace(id) == ""
	if !isCreate {
		existing, err := repo.Get(ctx, id)
		if err != nil {
			return AdminPost{}, err
		}
		item = existing
		previous = existing
	}

	item.Slug = defaultString(slugify(request.Slug), slugify(title))
	item.Title = title
	item.Summary = strings.TrimSpace(request.Summary)
	item.Content = content
	item.Category = defaultString(strings.TrimSpace(request.Category), "工程实践")
	item.Tags = normalizeTags(request.Tags)
	item.CoverImage = defaultString(strings.TrimSpace(request.CoverImage), "https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=1400&q=80")
	if strings.TrimSpace(item.AuthorID) == "" {
		item.AuthorID = strings.TrimSpace(actor.ID)
	}
	if strings.TrimSpace(item.AuthorName) == "" && strings.TrimSpace(actor.DisplayName) != "" {
		item.AuthorName = strings.TrimSpace(actor.DisplayName)
	}
	item.SEOtitle = defaultString(strings.TrimSpace(request.SEOtitle), title)
	item.SEODescription = defaultString(strings.TrimSpace(request.SEODescription), item.Summary)
	item.ScheduledAt = scheduledAt
	item.ReadingTime = estimateReadingTime(content)
	if request.Status != "" {
		item.Status = normalizeStatus(request.Status)
	}
	if item.Status == "" {
		item.Status = StatusDraft
	}
	if request.Visibility != "" {
		item.Visibility = normalizeVisibility(request.Visibility)
	}
	if item.Visibility == "" {
		item.Visibility = VisibilityPublic
	}
	if item.Status == StatusScheduled {
		if item.ScheduledAt == nil || strings.TrimSpace(item.Content) == "" {
			return AdminPost{}, ErrInvalidPost
		}
		if item.Visibility == VisibilityMembers {
			return AdminPost{}, ErrPostNotPublic
		}
	}

	item.UpdatedAt = now
	if isCreate || !postContentEqual(previous, item) || previous.Status != item.Status || !scheduledAtEqual(previous.ScheduledAt, item.ScheduledAt) {
		item.Version++
		item.Revisions = appendRevision(item.Revisions, snapshotRevision(item, now))
	}

	if err := repo.savePost(ctx, item, false); err != nil {
		return AdminPost{}, err
	}

	return clonePost(item), nil
}

func (repo *SQLRepository) Delete(ctx context.Context, id string) (AdminPost, error) {
	item, err := repo.Get(ctx, id)
	if err != nil {
		return AdminPost{}, err
	}

	result, err := repo.db.ExecContext(ctx, "DELETE FROM admin_posts WHERE id = $1", id)
	if err != nil {
		return AdminPost{}, fmt.Errorf("delete admin post: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return AdminPost{}, fmt.Errorf("read admin post delete rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return AdminPost{}, ErrPostNotFound
	}

	return item, nil
}

func scanAdminPostListItem(scanner interface{ Scan(dest ...any) error }) (AdminPost, error) {
	var data []byte
	var viewCount sql.NullInt64
	var commentCount sql.NullInt64
	if err := scanner.Scan(&data, &viewCount, &commentCount); err != nil {
		return AdminPost{}, fmt.Errorf("scan admin post: %w", err)
	}

	item, err := decodeAdminPost(data)
	if err != nil {
		return AdminPost{}, err
	}

	return applyPublicPostStats(item, viewCount, commentCount), nil
}

func applyPublicPostStats(item AdminPost, viewCount sql.NullInt64, commentCount sql.NullInt64) AdminPost {
	if viewCount.Valid {
		item.ViewCount = int(viewCount.Int64)
	}
	if commentCount.Valid {
		item.CommentCount = int(commentCount.Int64)
	}

	return item
}

func (repo *SQLRepository) Publish(ctx context.Context, id string, publisher posts.AdminPublisher, actor auth.User) (AdminPost, error) {
	item, err := repo.Get(ctx, id)
	if err != nil {
		return AdminPost{}, err
	}
	if strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.Content) == "" {
		return AdminPost{}, ErrInvalidPost
	}
	item.Visibility = normalizeVisibility(item.Visibility)
	if item.Visibility == VisibilityMembers {
		return AdminPost{}, ErrPostNotPublic
	}
	if publisher == nil {
		return AdminPost{}, ErrInvalidPost
	}
	if strings.TrimSpace(item.AuthorID) == "" {
		item.AuthorID = strings.TrimSpace(actor.ID)
	}
	if strings.TrimSpace(item.AuthorName) == "" && strings.TrimSpace(actor.DisplayName) != "" {
		item.AuthorName = strings.TrimSpace(actor.DisplayName)
	}

	published, err := publisher.PublishAdmin(ctx, posts.PublishInput{
		Slug:       item.Slug,
		Title:      item.Title,
		Summary:    item.Summary,
		Content:    item.Content,
		Visibility: item.Visibility,
		Category:   item.Category,
		Tags:       item.Tags,
		CoverImage: item.CoverImage,
		AuthorID:   item.AuthorID,
		AuthorName: item.AuthorName,
	}, item.PublishedPostSlug)
	if err != nil {
		return AdminPost{}, err
	}

	now := repo.now()
	item.Status = StatusPublished
	item.PublishedPostSlug = published.Slug
	item.PublishedAt = &now
	item.UpdatedAt = now
	item.ViewCount = published.ViewCount
	item.CommentCount = published.CommentCount
	// Publishing already-saved content should not create a second history entry.
	if !hasMatchingContentRevision(item) {
		item.Version++
		item.Revisions = appendRevision(item.Revisions, snapshotRevision(item, now))
	} else {
		item.Revisions = updateMatchingRevisionStatus(item, StatusPublished)
	}

	if err := repo.savePost(ctx, item, false); err != nil {
		return AdminPost{}, err
	}

	return clonePost(item), nil
}

func (repo *SQLRepository) Archive(ctx context.Context, id string, archiver posts.Archiver) (AdminPost, error) {
	item, err := repo.Get(ctx, id)
	if err != nil {
		return AdminPost{}, err
	}
	if archiver == nil {
		return AdminPost{}, ErrInvalidPost
	}

	targetSlug := strings.TrimSpace(item.PublishedPostSlug)
	if targetSlug == "" {
		targetSlug = strings.TrimSpace(item.Slug)
	}
	if targetSlug != "" && item.Status == StatusPublished {
		if err := archiver.Archive(ctx, targetSlug); err != nil {
			return AdminPost{}, err
		}
	}

	now := repo.now()
	item.Status = StatusArchived
	item.UpdatedAt = now
	item.Version++
	item.Revisions = appendRevision(item.Revisions, snapshotRevision(item, now))

	if err := repo.savePost(ctx, item, false); err != nil {
		return AdminPost{}, err
	}

	return clonePost(item), nil
}

func (repo *SQLRepository) PublishDue(ctx context.Context, publisher posts.AdminPublisher, now time.Time) (int, error) {
	if publisher == nil {
		return 0, ErrInvalidPost
	}

	rows, err := repo.db.QueryContext(ctx, "SELECT data FROM admin_posts WHERE status = $1 ORDER BY updated_at ASC, id ASC", StatusScheduled)
	if err != nil {
		return 0, fmt.Errorf("query scheduled admin posts: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		item, err := scanAdminPost(rows)
		if err != nil {
			return 0, err
		}
		if isDueScheduledPost(item, now) {
			ids = append(ids, item.ID)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate scheduled admin posts: %w", err)
	}

	publishedCount := 0
	var firstErr error
	for _, id := range ids {
		item, err := repo.Get(ctx, id)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if !isDueScheduledPost(item, now) {
			continue
		}

		actor := auth.User{ID: item.AuthorID, DisplayName: item.AuthorName, Role: "admin"}
		if _, err := repo.Publish(ctx, id, publisher, actor); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		publishedCount++
	}

	return publishedCount, firstErr
}

func (repo *SQLRepository) ListRevisions(ctx context.Context, id string) (RevisionListResult, error) {
	item, err := repo.Get(ctx, id)
	if err != nil {
		return RevisionListResult{}, err
	}

	revisions := sortedRevisions(item)
	return RevisionListResult{
		Items: revisions,
		Total: len(revisions),
	}, nil
}

func (repo *SQLRepository) RestoreRevision(ctx context.Context, id string, revisionID string) (AdminPost, error) {
	item, err := repo.Get(ctx, id)
	if err != nil {
		return AdminPost{}, err
	}

	revision, ok := findRevision(item, revisionID)
	if !ok {
		return AdminPost{}, ErrRevisionNotFound
	}

	now := repo.now()
	item = restoreFromRevision(item, revision)
	item.Version++
	item.UpdatedAt = now
	item.ReadingTime = estimateReadingTime(item.Content)
	item.Revisions = appendRevision(item.Revisions, snapshotRevision(item, now))

	if err := repo.savePost(ctx, item, false); err != nil {
		return AdminPost{}, err
	}

	return clonePost(item), nil
}

func (repo *SQLRepository) DeleteRevision(ctx context.Context, id string, revisionID string) (AdminPost, error) {
	item, err := repo.Get(ctx, id)
	if err != nil {
		return AdminPost{}, err
	}

	revisionID = strings.TrimSpace(revisionID)
	if revisionID == "" {
		return AdminPost{}, ErrRevisionNotFound
	}

	// Ensure current content has a revision snapshot so deleting history never loses current version.
	item.Revisions = ensureCurrentRevision(item)

	if item.Version > 0 {
		currentID := fmt.Sprintf("%s_rev_%d", item.ID, item.Version)
		if revisionID == currentID {
			return AdminPost{}, ErrInvalidPost
		}
	}

	revisions, found := removeRevision(item.Revisions, revisionID)
	if !found {
		return AdminPost{}, ErrRevisionNotFound
	}

	item.Revisions = revisions
	item.UpdatedAt = repo.now()
	if err := repo.savePost(ctx, item, false); err != nil {
		return AdminPost{}, err
	}

	return clonePost(item), nil
}

func (repo *SQLRepository) ensureSeedPosts(ctx context.Context) error {
	for _, item := range seedAdminPosts() {
		if err := repo.savePost(ctx, item, true); err != nil {
			return err
		}
	}

	return nil
}

func (repo *SQLRepository) savePost(ctx context.Context, item AdminPost, ignoreConflict bool) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal admin post %s: %w", item.ID, err)
	}

	query := `
		INSERT INTO admin_posts (id, slug, title, status, updated_at, data)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if ignoreConflict {
		query += " ON CONFLICT (id) DO NOTHING"
	} else {
		query += `
			ON CONFLICT (id)
			DO UPDATE SET
				slug = EXCLUDED.slug,
				title = EXCLUDED.title,
				status = EXCLUDED.status,
				updated_at = EXCLUDED.updated_at,
				data = EXCLUDED.data
		`
	}

	if _, err := repo.db.ExecContext(ctx, query, item.ID, item.Slug, item.Title, item.Status, item.UpdatedAt, data); err != nil {
		return fmt.Errorf("save admin post %s: %w", item.ID, err)
	}

	return nil
}

func scanAdminPost(scanner interface{ Scan(dest ...any) error }) (AdminPost, error) {
	var data []byte
	if err := scanner.Scan(&data); err != nil {
		return AdminPost{}, fmt.Errorf("scan admin post: %w", err)
	}

	return decodeAdminPost(data)
}

func decodeAdminPost(data []byte) (AdminPost, error) {
	var item AdminPost
	if err := json.Unmarshal(data, &item); err != nil {
		return AdminPost{}, fmt.Errorf("decode admin post: %w", err)
	}

	return clonePost(item), nil
}
