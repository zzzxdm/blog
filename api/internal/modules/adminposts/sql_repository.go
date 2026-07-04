package adminposts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"blog/api/internal/modules/posts"
)

type SQLRepository struct {
	db  *sql.DB
	now func() time.Time
}

func NewSQLRepository(ctx context.Context, db *sql.DB) (*SQLRepository, error) {
	repo := &SQLRepository{
		db:  db,
		now: time.Now,
	}
	if err := repo.ensureSeedPosts(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *SQLRepository) List(ctx context.Context) (ListResult, error) {
	rows, err := repo.db.QueryContext(ctx, "SELECT data FROM admin_posts ORDER BY updated_at DESC, id DESC")
	if err != nil {
		return ListResult{}, fmt.Errorf("query admin posts: %w", err)
	}
	defer rows.Close()

	items := make([]AdminPost, 0)
	for rows.Next() {
		item, err := scanAdminPost(rows)
		if err != nil {
			return ListResult{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ListResult{}, fmt.Errorf("iterate admin posts: %w", err)
	}

	return ListResult{
		Items: items,
		Total: len(items),
		Stats: countStats(items),
	}, nil
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

func (repo *SQLRepository) Save(ctx context.Context, id string, request SaveRequest) (AdminPost, error) {
	title := strings.TrimSpace(request.Title)
	content := strings.TrimSpace(request.Content)
	if title == "" {
		return AdminPost{}, ErrInvalidPost
	}

	now := repo.now()
	item := AdminPost{
		ID:         fmt.Sprintf("admin_post_%d", now.UnixNano()),
		AuthorName: "管理员",
		Status:     StatusDraft,
		Version:    0,
		UpdatedAt:  now,
	}

	if strings.TrimSpace(id) != "" {
		existing, err := repo.Get(ctx, id)
		if err != nil {
			return AdminPost{}, err
		}
		item = existing
	}

	item.Slug = defaultString(slugify(request.Slug), slugify(title))
	item.Title = title
	item.Summary = strings.TrimSpace(request.Summary)
	item.Content = content
	item.Category = defaultString(strings.TrimSpace(request.Category), "工程实践")
	item.Tags = normalizeTags(request.Tags)
	item.CoverImage = defaultString(strings.TrimSpace(request.CoverImage), "https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=1400&q=80")
	item.SEOtitle = defaultString(strings.TrimSpace(request.SEOtitle), title)
	item.SEODescription = defaultString(strings.TrimSpace(request.SEODescription), item.Summary)
	item.ReadingTime = estimateReadingTime(content)
	item.UpdatedAt = now
	item.Version++
	if request.Status != "" {
		item.Status = normalizeStatus(request.Status)
	}
	if item.Status == "" {
		item.Status = StatusDraft
	}
	item.Revisions = appendRevision(item.Revisions, snapshotRevision(item, now))

	if err := repo.savePost(ctx, item, false); err != nil {
		return AdminPost{}, err
	}

	return clonePost(item), nil
}

func (repo *SQLRepository) Publish(ctx context.Context, id string, publisher posts.Publisher) (AdminPost, error) {
	item, err := repo.Get(ctx, id)
	if err != nil {
		return AdminPost{}, err
	}
	if strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.Content) == "" {
		return AdminPost{}, ErrInvalidPost
	}
	if item.Status == StatusPublished && item.PublishedPostSlug != "" {
		return clonePost(item), nil
	}
	if publisher == nil {
		return AdminPost{}, ErrInvalidPost
	}

	published, err := publisher.Publish(ctx, posts.PublishInput{
		Slug:       item.Slug,
		Title:      item.Title,
		Summary:    item.Summary,
		Content:    item.Content,
		Category:   item.Category,
		Tags:       item.Tags,
		CoverImage: item.CoverImage,
		AuthorName: item.AuthorName,
	})
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
	item.Version++
	item.Revisions = appendRevision(item.Revisions, snapshotRevision(item, now))

	if err := repo.savePost(ctx, item, false); err != nil {
		return AdminPost{}, err
	}

	return clonePost(item), nil
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
