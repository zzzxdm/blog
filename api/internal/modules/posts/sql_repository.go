package posts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"blog/api/internal/database"
	"blog/api/internal/idgen"
)

type SQLRepository struct {
	db     *sql.DB
	sqlite bool
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db, sqlite: database.IsSQLite(db)}
}

func (repo *SQLRepository) List(ctx context.Context, query ListQuery) (ListResult, error) {
	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	offset := (page - 1) * pageSize
	keyword := strings.TrimSpace(query.Keyword)
	category := strings.TrimSpace(query.Category)
	tag := strings.TrimSpace(query.Tag)
	author := strings.TrimSpace(query.Author)
	sortMode := strings.ToLower(strings.TrimSpace(query.Sort))
	if sortMode != "views" && sortMode != "comments" && sortMode != "likes" {
		sortMode = ""
	}

	if repo.sqlite {
		return repo.listSQLite(ctx, page, pageSize, offset, keyword, category, tag, author, sortMode)
	}

	total, err := repo.count(ctx, keyword, category, tag, author)
	if err != nil {
		return ListResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		WITH input AS (
			SELECT
				NULLIF(trim($1), '') AS keyword,
				CASE
					WHEN NULLIF(trim($1), '') IS NULL THEN NULL
					ELSE websearch_to_tsquery('simple', $1)
				END AS tsquery
		)
		SELECT
			p.id::text,
			p.slug,
				p.title,
				p.summary,
				p.content,
				p.visibility,
				c.name AS category,
				COALESCE(string_agg(t.name, ',' ORDER BY t.name), '') AS tags,
				p.cover_image,
				COALESCE(p.author_id::text, '') AS author_id,
				p.author_name,
			p.reading_time,
			p.view_count,
			p.like_count,
			p.dislike_count,
			p.comment_count,
			COALESCE(p.published_at, p.created_at) AS published_at
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
			CROSS JOIN input
			WHERE p.status = 'published'
				AND p.visibility = 'public'
				AND ($2 = '' OR lower(c.slug) = lower($2) OR lower(c.name) = lower($2))
				AND (
					$3 = ''
					OR EXISTS (
					SELECT 1
					FROM post_tags filter_pt
					JOIN tags filter_t ON filter_t.id = filter_pt.tag_id
					WHERE filter_pt.post_id = p.id
							AND (lower(filter_t.slug) = lower($3) OR lower(filter_t.name) = lower($3))
					)
				)
				AND (
					$4 = ''
					OR lower(p.author_name) = lower($4)
					OR lower(replace(p.author_name, ' ', '-')) = lower($4)
				)
				AND (
					input.keyword IS NULL
					OR p.search_vector @@ input.tsquery
				OR p.title ILIKE '%' || input.keyword || '%'
				OR p.summary ILIKE '%' || input.keyword || '%'
				OR c.name ILIKE '%' || input.keyword || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags search_pt
					JOIN tags search_t ON search_t.id = search_pt.tag_id
					WHERE search_pt.post_id = p.id
						AND search_t.name ILIKE '%' || input.keyword || '%'
				)
					OR ($8 AND p.content ILIKE '%' || input.keyword || '%')
			)
			GROUP BY p.id, c.name, input.keyword, input.tsquery
			ORDER BY
				CASE WHEN $7 = 'views' THEN p.view_count END DESC NULLS LAST,
				CASE WHEN $7 = 'comments' THEN p.comment_count END DESC NULLS LAST,
				CASE WHEN $7 = 'likes' THEN p.like_count END DESC NULLS LAST,
				CASE
					WHEN $7 <> '' OR input.keyword IS NULL OR input.tsquery IS NULL THEN 0
					ELSE ts_rank_cd(p.search_vector, input.tsquery)
				END DESC,
				p.published_at DESC NULLS LAST,
				p.created_at DESC
			LIMIT $5 OFFSET $6
		`, keyword, category, tag, author, pageSize, offset, sortMode, hasCJK(keyword))
	if err != nil {
		return ListResult{}, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	items := make([]Post, 0, pageSize)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return ListResult{}, err
		}
		items = append(items, post)
	}
	if err := rows.Err(); err != nil {
		return ListResult{}, fmt.Errorf("iterate posts: %w", err)
	}

	return ListResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (repo *SQLRepository) ListPrivate(ctx context.Context, viewer Viewer, query ListQuery) (ListResult, error) {
	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	offset := (page - 1) * pageSize
	keyword := strings.TrimSpace(query.Keyword)

	if repo.sqlite {
		return repo.listPrivateSQLite(ctx, viewer, page, pageSize, offset, keyword)
	}

	total, err := repo.countPrivate(ctx, viewer, keyword)
	if err != nil {
		return ListResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			p.id::text,
			p.slug,
			p.title,
			p.summary,
			p.content,
			p.visibility,
			c.name AS category,
			COALESCE(string_agg(t.name, ',' ORDER BY t.name), '') AS tags,
			p.cover_image,
			COALESCE(p.author_id::text, '') AS author_id,
			p.author_name,
			p.reading_time,
			p.view_count,
			p.like_count,
			p.dislike_count,
			p.comment_count,
			COALESCE(p.published_at, p.created_at) AS published_at
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.status = 'published'
			AND p.visibility = 'private'
			AND ($1 = 'admin' OR COALESCE(p.author_id::text, '') = $2)
			AND (
				$3 = ''
				OR p.title ILIKE '%' || $3 || '%'
				OR p.summary ILIKE '%' || $3 || '%'
				OR c.name ILIKE '%' || $3 || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags search_pt
					JOIN tags search_t ON search_t.id = search_pt.tag_id
					WHERE search_pt.post_id = p.id
						AND search_t.name ILIKE '%' || $3 || '%'
				)
					OR ($6 AND p.content ILIKE '%' || $3 || '%')
			)
		GROUP BY p.id, c.name
		ORDER BY COALESCE(p.published_at, p.created_at) DESC, p.created_at DESC
		LIMIT $4 OFFSET $5
	`, viewer.Role, strings.TrimSpace(viewer.ID), keyword, pageSize, offset, hasCJK(keyword))
	if err != nil {
		return ListResult{}, fmt.Errorf("list private posts: %w", err)
	}
	defer rows.Close()

	items, err := scanPostRows(rows, pageSize)
	if err != nil {
		return ListResult{}, err
	}

	return ListResult{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (repo *SQLRepository) GetBySlug(ctx context.Context, slug string) (Post, error) {
	if repo.sqlite {
		return repo.getBySlugSQLite(ctx, slug)
	}

	row := repo.db.QueryRowContext(ctx, `
		SELECT
			p.id::text,
			p.slug,
				p.title,
				p.summary,
				p.content,
				p.visibility,
				c.name AS category,
				COALESCE(string_agg(t.name, ',' ORDER BY t.name), '') AS tags,
				p.cover_image,
				COALESCE(p.author_id::text, '') AS author_id,
				p.author_name,
			p.reading_time,
			p.view_count,
			p.like_count,
			p.dislike_count,
			p.comment_count,
			COALESCE(p.published_at, p.created_at) AS published_at
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
			WHERE p.slug = $1
				AND p.status = 'published'
				AND p.visibility = 'public'
		GROUP BY p.id, c.name
	`, slug)

	post, err := scanPost(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Post{}, ErrNotFound
		}
		return Post{}, err
	}

	return post, nil
}

func (repo *SQLRepository) GetBySlugForViewer(ctx context.Context, slug string, viewer Viewer) (Post, error) {
	post, err := repo.getBySlugAnyVisibility(ctx, slug)
	if err != nil {
		return Post{}, err
	}
	if !canViewPost(post, viewer) {
		return Post{}, ErrNotFound
	}

	return post, nil
}

func (repo *SQLRepository) RecordRestrictedView(ctx context.Context, slug string, viewer Viewer) (Post, error) {
	post, err := repo.GetBySlugForViewer(ctx, slug, viewer)
	if err != nil {
		return Post{}, err
	}

	if _, err := repo.db.ExecContext(ctx, `
		UPDATE posts
		SET view_count = view_count + 1
		WHERE slug = $1
			AND status = 'published'
	`, slug); err != nil {
		return Post{}, fmt.Errorf("record post view: %w", err)
	}

	post.ViewCount++
	return post, nil
}

func (repo *SQLRepository) Stats(ctx context.Context) (SiteStats, error) {
	var stats SiteStats
	if repo.sqlite {
		err := repo.db.QueryRowContext(ctx, `
			SELECT
				count(*),
				COALESCE(sum(view_count), 0),
				COALESCE(sum(length(replace(replace(content, ' ', ''), char(10), ''))), 0)
			FROM posts
				WHERE status = 'published'
					AND visibility = 'public'
		`).Scan(&stats.PostCount, &stats.ViewCount, &stats.WordCount)
		if err != nil {
			return SiteStats{}, fmt.Errorf("query site stats: %w", err)
		}

		return stats, nil
	}

	err := repo.db.QueryRowContext(ctx, `
		SELECT
			count(*)::int,
			COALESCE(sum(view_count), 0)::int,
			COALESCE(sum(char_length(regexp_replace(content, '\s+', '', 'g'))), 0)::int
		FROM posts
			WHERE status = 'published'
				AND visibility = 'public'
	`).Scan(&stats.PostCount, &stats.ViewCount, &stats.WordCount)
	if err != nil {
		return SiteStats{}, fmt.Errorf("query site stats: %w", err)
	}

	return stats, nil
}

func (repo *SQLRepository) RecordView(ctx context.Context, slug string) (Post, error) {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE posts
		SET view_count = view_count + 1
			WHERE slug = $1
				AND status = 'published'
				AND visibility = 'public'
	`, slug)
	if err != nil {
		return Post{}, fmt.Errorf("record post view: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Post{}, fmt.Errorf("read post view rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return Post{}, ErrNotFound
	}

	return repo.GetBySlug(ctx, slug)
}

func (repo *SQLRepository) Publish(ctx context.Context, input PublishInput) (Post, error) {
	input.Visibility = VisibilityPublic
	return repo.PublishSubmission(ctx, input, "")
}

func (repo *SQLRepository) PublishAdmin(ctx context.Context, input PublishInput, existingSlug string) (Post, error) {
	return repo.publishManaged(ctx, input, existingSlug, "admin", "工程实践", "管理员", "https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=1400&q=80")
}

func (repo *SQLRepository) PublishSubmission(ctx context.Context, input PublishInput, existingSlug string) (Post, error) {
	return repo.publishManaged(ctx, input, existingSlug, "submission", "投稿", "注册用户", "https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=1400&q=80")
}

func (repo *SQLRepository) Archive(ctx context.Context, slug string) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE posts
		SET status = 'archived',
			updated_at = $2
		WHERE slug = $1
			AND status = 'published'
	`, strings.TrimSpace(slug), time.Now())
	if err != nil {
		return fmt.Errorf("archive post: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read archived post rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (repo *SQLRepository) Restore(ctx context.Context, slug string) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE posts
		SET status = 'published',
			published_at = COALESCE(published_at, $2),
			updated_at = $2
		WHERE slug = $1
			AND status = 'archived'
	`, strings.TrimSpace(slug), time.Now())
	if err != nil {
		return fmt.Errorf("restore post: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read restored post rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (repo *SQLRepository) publishManaged(ctx context.Context, input PublishInput, existingSlug string, source string, defaultCategory string, defaultAuthorName string, defaultCoverImage string) (Post, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if title == "" || content == "" {
		return Post{}, ErrInvalidPost
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Post{}, fmt.Errorf("begin publish transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	category := defaultString(strings.TrimSpace(input.Category), defaultCategory)
	categoryID, err := ensureCategory(ctx, tx, category)
	if err != nil {
		return Post{}, err
	}

	baseSlug := defaultString(slugify(input.Slug), slugify(title))
	if baseSlug == "" {
		baseSlug = "post"
	}
	visibility := normalizeVisibility(input.Visibility)

	existingID, err := findPostIDBySlug(ctx, tx, strings.TrimSpace(existingSlug))
	if err != nil {
		return Post{}, err
	}

	slug, err := uniqueSQLSlugExcept(ctx, tx, baseSlug, existingID)
	if err != nil {
		return Post{}, err
	}

	now := time.Now()
	postID := existingID
	if postID == "" {
		err = tx.QueryRowContext(ctx, `
				INSERT INTO posts (
						id, slug, title, summary, content, visibility, status, source, category_id, author_id, author_name,
						cover_image, reading_time, view_count, like_count, dislike_count, comment_count, published_at
					)
							VALUES ($1, $2, $3, $4, $5, $6, 'published', $7, $8, CAST(NULLIF($9, '') AS bigint), $10, $11, $12, 0, 0, 0, 0, $13)
					RETURNING CAST(id AS TEXT)
				`,
			idgen.NextString(),
			slug,
			title,
			strings.TrimSpace(input.Summary),
			content,
			visibility,
			source,
			categoryID,
			strings.TrimSpace(input.AuthorID),
			defaultString(strings.TrimSpace(input.AuthorName), defaultAuthorName),
			defaultString(strings.TrimSpace(input.CoverImage), defaultCoverImage),
			estimateReadingTime(content),
			now,
		).Scan(&postID)
		if err != nil {
			return Post{}, fmt.Errorf("insert published post: %w", err)
		}
	} else {
		err = tx.QueryRowContext(ctx, `
			UPDATE posts
			SET
				slug = $2,
				title = $3,
				summary = $4,
					content = $5,
					visibility = $6,
					status = 'published',
					source = $7,
					category_id = $8,
						author_id = CAST(NULLIF($9, '') AS bigint),
					author_name = $10,
					cover_image = $11,
					reading_time = $12,
					published_at = COALESCE(published_at, $13),
					updated_at = $13
				WHERE id = $1
				RETURNING CAST(id AS TEXT)
			`,
			postID,
			slug,
			title,
			strings.TrimSpace(input.Summary),
			content,
			visibility,
			source,
			categoryID,
			strings.TrimSpace(input.AuthorID),
			defaultString(strings.TrimSpace(input.AuthorName), defaultAuthorName),
			defaultString(strings.TrimSpace(input.CoverImage), defaultCoverImage),
			estimateReadingTime(content),
			now,
		).Scan(&postID)
		if err != nil {
			return Post{}, fmt.Errorf("update published post: %w", err)
		}

		if _, err := tx.ExecContext(ctx, "DELETE FROM post_tags WHERE post_id = $1", postID); err != nil {
			return Post{}, fmt.Errorf("delete post tags: %w", err)
		}
	}

	if err := savePostTags(ctx, tx, postID, normalizeTags(input.Tags)); err != nil {
		return Post{}, err
	}

	if err := tx.Commit(); err != nil {
		return Post{}, fmt.Errorf("commit publish transaction: %w", err)
	}
	committed = true

	return repo.getBySlugAnyVisibility(ctx, slug)
}

func (repo *SQLRepository) count(ctx context.Context, keyword string, category string, tag string, author string) (int, error) {
	if repo.sqlite {
		return repo.countSQLite(ctx, keyword, category, tag, author)
	}

	var total int
	err := repo.db.QueryRowContext(ctx, `
		WITH input AS (
			SELECT
				NULLIF(trim($1), '') AS keyword,
				CASE
					WHEN NULLIF(trim($1), '') IS NULL THEN NULL
					ELSE websearch_to_tsquery('simple', $1)
				END AS tsquery
		)
		SELECT count(*)
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		CROSS JOIN input
		WHERE p.status = 'published'
			AND p.visibility = 'public'
			AND ($2 = '' OR lower(c.slug) = lower($2) OR lower(c.name) = lower($2))
				AND (
					$3 = ''
					OR EXISTS (
					SELECT 1
					FROM post_tags filter_pt
					JOIN tags filter_t ON filter_t.id = filter_pt.tag_id
					WHERE filter_pt.post_id = p.id
							AND (lower(filter_t.slug) = lower($3) OR lower(filter_t.name) = lower($3))
					)
				)
				AND (
					$4 = ''
					OR lower(p.author_name) = lower($4)
					OR lower(replace(p.author_name, ' ', '-')) = lower($4)
				)
				AND (
					input.keyword IS NULL
					OR p.search_vector @@ input.tsquery
				OR p.title ILIKE '%' || input.keyword || '%'
				OR p.summary ILIKE '%' || input.keyword || '%'
				OR c.name ILIKE '%' || input.keyword || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags search_pt
					JOIN tags search_t ON search_t.id = search_pt.tag_id
					WHERE search_pt.post_id = p.id
						AND search_t.name ILIKE '%' || input.keyword || '%'
				)
					OR ($5 AND p.content ILIKE '%' || input.keyword || '%')
				)
		`, keyword, category, tag, author, hasCJK(keyword)).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count posts: %w", err)
	}

	return total, nil
}

func (repo *SQLRepository) countPrivate(ctx context.Context, viewer Viewer, keyword string) (int, error) {
	var total int
	err := repo.db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		WHERE p.status = 'published'
			AND p.visibility = 'private'
			AND ($1 = 'admin' OR COALESCE(CAST(p.author_id AS TEXT), '') = $2)
			AND (
				$3 = ''
				OR lower(p.title) LIKE '%' || lower($3) || '%'
				OR lower(p.summary) LIKE '%' || lower($3) || '%'
				OR lower(c.name) LIKE '%' || lower($3) || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags search_pt
					JOIN tags search_t ON search_t.id = search_pt.tag_id
					WHERE search_pt.post_id = p.id
						AND lower(search_t.name) LIKE '%' || lower($3) || '%'
				)
			)
	`, viewer.Role, strings.TrimSpace(viewer.ID), keyword, hasCJK(keyword)).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count private posts: %w", err)
	}

	return total, nil
}

func (repo *SQLRepository) listSQLite(ctx context.Context, page int, pageSize int, offset int, keyword string, category string, tag string, author string, sortMode string) (ListResult, error) {
	total, err := repo.countSQLite(ctx, keyword, category, tag, author)
	if err != nil {
		return ListResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			p.id,
			p.slug,
			p.title,
			p.summary,
			p.content,
			p.visibility,
			c.name AS category,
			COALESCE(group_concat(t.name, ','), '') AS tags,
			p.cover_image,
			COALESCE(CAST(p.author_id AS TEXT), '') AS author_id,
			p.author_name,
			p.reading_time,
			p.view_count,
			p.like_count,
			p.dislike_count,
			p.comment_count,
			COALESCE(p.published_at, p.created_at) AS published_at
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.status = 'published'
			AND p.visibility = 'public'
			AND ($2 = '' OR lower(c.slug) = lower($2) OR lower(c.name) = lower($2))
			AND (
				$3 = ''
				OR EXISTS (
					SELECT 1
					FROM post_tags filter_pt
					JOIN tags filter_t ON filter_t.id = filter_pt.tag_id
					WHERE filter_pt.post_id = p.id
						AND (lower(filter_t.slug) = lower($3) OR lower(filter_t.name) = lower($3))
				)
			)
			AND (
				$4 = ''
				OR lower(p.author_name) = lower($4)
				OR lower(replace(p.author_name, ' ', '-')) = lower($4)
			)
			AND (
				$1 = ''
				OR lower(p.title) LIKE '%' || lower($1) || '%'
				OR lower(p.summary) LIKE '%' || lower($1) || '%'
				OR lower(c.name) LIKE '%' || lower($1) || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags search_pt
					JOIN tags search_t ON search_t.id = search_pt.tag_id
					WHERE search_pt.post_id = p.id
						AND lower(search_t.name) LIKE '%' || lower($1) || '%'
				)
			)
		GROUP BY p.id, c.name
		ORDER BY
			CASE WHEN $7 = 'views' THEN p.view_count END DESC,
			CASE WHEN $7 = 'comments' THEN p.comment_count END DESC,
			CASE WHEN $7 = 'likes' THEN p.like_count END DESC,
			COALESCE(p.published_at, p.created_at) DESC,
			p.created_at DESC
		LIMIT $5 OFFSET $6
	`, keyword, category, tag, author, pageSize, offset, sortMode)
	if err != nil {
		return ListResult{}, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	items := make([]Post, 0, pageSize)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return ListResult{}, err
		}
		items = append(items, post)
	}
	if err := rows.Err(); err != nil {
		return ListResult{}, fmt.Errorf("iterate posts: %w", err)
	}

	return ListResult{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (repo *SQLRepository) listPrivateSQLite(ctx context.Context, viewer Viewer, page int, pageSize int, offset int, keyword string) (ListResult, error) {
	total, err := repo.countPrivate(ctx, viewer, keyword)
	if err != nil {
		return ListResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			p.id,
			p.slug,
			p.title,
			p.summary,
			p.content,
			p.visibility,
			c.name AS category,
			COALESCE(group_concat(t.name, ','), '') AS tags,
			p.cover_image,
			COALESCE(CAST(p.author_id AS TEXT), '') AS author_id,
			p.author_name,
			p.reading_time,
			p.view_count,
			p.like_count,
			p.dislike_count,
			p.comment_count,
			COALESCE(p.published_at, p.created_at) AS published_at
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.status = 'published'
			AND p.visibility = 'private'
			AND ($1 = 'admin' OR COALESCE(CAST(p.author_id AS TEXT), '') = $2)
			AND (
				$3 = ''
				OR lower(p.title) LIKE '%' || lower($3) || '%'
				OR lower(p.summary) LIKE '%' || lower($3) || '%'
				OR lower(c.name) LIKE '%' || lower($3) || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags search_pt
					JOIN tags search_t ON search_t.id = search_pt.tag_id
					WHERE search_pt.post_id = p.id
						AND lower(search_t.name) LIKE '%' || lower($3) || '%'
				)
			)
		GROUP BY p.id, c.name
		ORDER BY COALESCE(p.published_at, p.created_at) DESC, p.created_at DESC
		LIMIT $4 OFFSET $5
	`, viewer.Role, strings.TrimSpace(viewer.ID), keyword, pageSize, offset)
	if err != nil {
		return ListResult{}, fmt.Errorf("list private posts: %w", err)
	}
	defer rows.Close()

	items, err := scanPostRows(rows, pageSize)
	if err != nil {
		return ListResult{}, err
	}

	return ListResult{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (repo *SQLRepository) getBySlugSQLite(ctx context.Context, slug string) (Post, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
			p.id,
			p.slug,
			p.title,
			p.summary,
			p.content,
			p.visibility,
			c.name AS category,
			COALESCE(group_concat(t.name, ','), '') AS tags,
			p.cover_image,
			COALESCE(CAST(p.author_id AS TEXT), '') AS author_id,
			p.author_name,
			p.reading_time,
			p.view_count,
			p.like_count,
			p.dislike_count,
			p.comment_count,
			COALESCE(p.published_at, p.created_at) AS published_at
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.slug = $1
			AND p.status = 'published'
			AND p.visibility = 'public'
		GROUP BY p.id, c.name
	`, slug)

	post, err := scanPost(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Post{}, ErrNotFound
		}
		return Post{}, err
	}

	return post, nil
}

func (repo *SQLRepository) getBySlugAnyVisibility(ctx context.Context, slug string) (Post, error) {
	idExpr := "p.id::text"
	authorIDExpr := "COALESCE(p.author_id::text, '')"
	tagsExpr := "COALESCE(string_agg(t.name, ',' ORDER BY t.name), '')"
	if repo.sqlite {
		idExpr = "p.id"
		authorIDExpr = "COALESCE(CAST(p.author_id AS TEXT), '')"
		tagsExpr = "COALESCE(group_concat(t.name, ','), '')"
	}

	row := repo.db.QueryRowContext(ctx, `
		SELECT
			`+idExpr+`,
			p.slug,
			p.title,
			p.summary,
			p.content,
			p.visibility,
			c.name AS category,
			`+tagsExpr+` AS tags,
			p.cover_image,
			`+authorIDExpr+` AS author_id,
			p.author_name,
			p.reading_time,
			p.view_count,
			p.like_count,
			p.dislike_count,
			p.comment_count,
			COALESCE(p.published_at, p.created_at) AS published_at
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.slug = $1
			AND p.status = 'published'
		GROUP BY p.id, c.name
	`, slug)

	post, err := scanPost(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Post{}, ErrNotFound
		}
		return Post{}, err
	}

	return post, nil
}

func (repo *SQLRepository) countSQLite(ctx context.Context, keyword string, category string, tag string, author string) (int, error) {
	var total int
	err := repo.db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		WHERE p.status = 'published'
			AND p.visibility = 'public'
			AND ($2 = '' OR lower(c.slug) = lower($2) OR lower(c.name) = lower($2))
			AND (
				$3 = ''
				OR EXISTS (
					SELECT 1
					FROM post_tags filter_pt
					JOIN tags filter_t ON filter_t.id = filter_pt.tag_id
					WHERE filter_pt.post_id = p.id
						AND (lower(filter_t.slug) = lower($3) OR lower(filter_t.name) = lower($3))
				)
			)
			AND (
				$4 = ''
				OR lower(p.author_name) = lower($4)
				OR lower(replace(p.author_name, ' ', '-')) = lower($4)
			)
			AND (
				$1 = ''
				OR lower(p.title) LIKE '%' || lower($1) || '%'
				OR lower(p.summary) LIKE '%' || lower($1) || '%'
				OR lower(c.name) LIKE '%' || lower($1) || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags search_pt
					JOIN tags search_t ON search_t.id = search_pt.tag_id
					WHERE search_pt.post_id = p.id
						AND lower(search_t.name) LIKE '%' || lower($1) || '%'
				)
			)
	`, keyword, category, tag, author).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count posts: %w", err)
	}

	return total, nil
}

func ensureCategory(ctx context.Context, tx *sql.Tx, category string) (string, error) {
	var categoryID string
	err := tx.QueryRowContext(ctx, `
		SELECT CAST(id AS TEXT)
		FROM categories
		WHERE lower(name) = lower($1) OR lower(slug) = lower($1)
		LIMIT 1
	`, category).Scan(&categoryID)
	if err == nil {
		return categoryID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("query category: %w", err)
	}

	slug := slugify(category)
	if slug == "" {
		slug = fmt.Sprintf("category-%d", time.Now().UnixNano())
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO categories (id, slug, name)
		VALUES ($1, $2, $3)
		RETURNING CAST(id AS TEXT)
	`, idgen.NextString(), slug, category).Scan(&categoryID)
	if err != nil {
		return "", fmt.Errorf("insert category: %w", err)
	}

	return categoryID, nil
}

func ensureTag(ctx context.Context, tx *sql.Tx, tag string) (string, error) {
	var tagID string
	err := tx.QueryRowContext(ctx, `
		SELECT CAST(id AS TEXT)
		FROM tags
		WHERE lower(name) = lower($1) OR lower(slug) = lower($1)
		LIMIT 1
	`, tag).Scan(&tagID)
	if err == nil {
		return tagID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("query tag: %w", err)
	}

	slug := slugify(tag)
	if slug == "" {
		slug = fmt.Sprintf("tag-%d", time.Now().UnixNano())
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO tags (id, slug, name)
		VALUES ($1, $2, $3)
		RETURNING CAST(id AS TEXT)
	`, idgen.NextString(), slug, tag).Scan(&tagID)
	if err != nil {
		return "", fmt.Errorf("insert tag: %w", err)
	}

	return tagID, nil
}

func savePostTags(ctx context.Context, tx *sql.Tx, postID string, tags []string) error {
	for _, tag := range tags {
		tagID, err := ensureTag(ctx, tx, tag)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO post_tags (post_id, tag_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, postID, tagID); err != nil {
			return fmt.Errorf("insert post tag: %w", err)
		}
	}

	return nil
}

func findPostIDBySlug(ctx context.Context, tx *sql.Tx, slug string) (string, error) {
	if slug == "" {
		return "", nil
	}

	var postID string
	err := tx.QueryRowContext(ctx, `
		SELECT CAST(id AS TEXT)
		FROM posts
		WHERE slug = $1
		LIMIT 1
	`, slug).Scan(&postID)
	if err == nil {
		return postID, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return "", fmt.Errorf("query existing post: %w", err)
}

func uniqueSQLSlug(ctx context.Context, tx *sql.Tx, slug string) (string, error) {
	return uniqueSQLSlugExcept(ctx, tx, slug, "")
}

func uniqueSQLSlugExcept(ctx context.Context, tx *sql.Tx, slug string, exceptPostID string) (string, error) {
	candidate := slug
	for suffix := 2; ; suffix++ {
		var exists bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM posts
				WHERE slug = $1
					AND ($2 = '' OR CAST(id AS TEXT) <> $2)
			)
		`, candidate, exceptPostID).Scan(&exists); err != nil {
			return "", fmt.Errorf("query slug: %w", err)
		}
		if !exists {
			return candidate, nil
		}

		candidate = fmt.Sprintf("%s-%d", slug, suffix)
	}
}

type postScanner interface {
	Scan(dest ...any) error
}

type postRows interface {
	postScanner
	Next() bool
	Err() error
}

func scanPostRows(rows postRows, capacity int) ([]Post, error) {
	items := make([]Post, 0, capacity)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}

	return items, nil
}

func scanPost(scanner postScanner) (Post, error) {
	var post Post
	var tagsCSV string
	var publishedAt database.FlexibleTime

	if err := scanner.Scan(
		&post.ID,
		&post.Slug,
		&post.Title,
		&post.Summary,
		&post.Content,
		&post.Visibility,
		&post.Category,
		&tagsCSV,
		&post.CoverImage,
		&post.AuthorID,
		&post.AuthorName,
		&post.ReadingTime,
		&post.ViewCount,
		&post.LikeCount,
		&post.DislikeCount,
		&post.CommentCount,
		&publishedAt,
	); err != nil {
		return Post{}, fmt.Errorf("scan post: %w", err)
	}

	post.Visibility = normalizeVisibility(post.Visibility)
	post.PublishedAt = publishedAt.Time
	post.Tags = splitTags(tagsCSV)
	return post, nil
}

func splitTags(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags
}

func hasCJK(s string) bool {
	for _, r := range s {
		if (r >= 0x4E00 && r <= 0x9FFF) || (r >= 0x3400 && r <= 0x4DBF) {
			return true
		}
	}
	return false
}
