package posts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (repo *SQLRepository) List(ctx context.Context, query ListQuery) (ListResult, error) {
	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	offset := (page - 1) * pageSize
	keyword := strings.TrimSpace(query.Keyword)
	category := strings.TrimSpace(query.Category)
	tag := strings.TrimSpace(query.Tag)
	sortMode := strings.ToLower(strings.TrimSpace(query.Sort))
	if sortMode != "views" && sortMode != "comments" && sortMode != "likes" {
		sortMode = ""
	}

	total, err := repo.count(ctx, keyword, category, tag)
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
			c.name AS category,
			COALESCE(string_agg(t.name, ',' ORDER BY t.name), '') AS tags,
			p.cover_image,
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
				input.keyword IS NULL
				OR p.search_vector @@ input.tsquery
				OR p.title ILIKE '%' || input.keyword || '%'
				OR p.summary ILIKE '%' || input.keyword || '%'
				OR p.content ILIKE '%' || input.keyword || '%'
				OR c.name ILIKE '%' || input.keyword || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags search_pt
					JOIN tags search_t ON search_t.id = search_pt.tag_id
					WHERE search_pt.post_id = p.id
						AND search_t.name ILIKE '%' || input.keyword || '%'
				)
			)
		GROUP BY p.id, c.name, input.keyword, input.tsquery
		ORDER BY
			CASE WHEN $6 = 'views' THEN p.view_count END DESC NULLS LAST,
			CASE WHEN $6 = 'comments' THEN p.comment_count END DESC NULLS LAST,
			CASE WHEN $6 = 'likes' THEN p.like_count END DESC NULLS LAST,
			CASE
				WHEN $6 <> '' OR input.keyword IS NULL OR input.tsquery IS NULL THEN 0
				ELSE ts_rank_cd(p.search_vector, input.tsquery)
			END DESC,
			p.published_at DESC NULLS LAST,
			p.created_at DESC
		LIMIT $4 OFFSET $5
	`, keyword, category, tag, pageSize, offset, sortMode)
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

func (repo *SQLRepository) GetBySlug(ctx context.Context, slug string) (Post, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
			p.id::text,
			p.slug,
			p.title,
			p.summary,
			p.content,
			c.name AS category,
			COALESCE(string_agg(t.name, ',' ORDER BY t.name), '') AS tags,
			p.cover_image,
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

func (repo *SQLRepository) RecordView(ctx context.Context, slug string) (Post, error) {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE posts
		SET view_count = view_count + 1
		WHERE slug = $1
			AND status = 'published'
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

	category := defaultString(strings.TrimSpace(input.Category), "投稿")
	categoryID, err := ensureCategory(ctx, tx, category)
	if err != nil {
		return Post{}, err
	}

	baseSlug := defaultString(slugify(input.Slug), slugify(title))
	if baseSlug == "" {
		baseSlug = "post"
	}

	slug, err := uniqueSQLSlug(ctx, tx, baseSlug)
	if err != nil {
		return Post{}, err
	}

	var postID string
	publishedAt := time.Now()
	err = tx.QueryRowContext(ctx, `
		INSERT INTO posts (
			slug, title, summary, content, status, source, category_id, author_name,
			cover_image, reading_time, view_count, like_count, dislike_count, comment_count, published_at
		)
		VALUES ($1, $2, $3, $4, 'published', 'submission', $5, $6, $7, $8, 0, 0, 0, 0, $9)
		RETURNING id::text
	`,
		slug,
		title,
		strings.TrimSpace(input.Summary),
		content,
		categoryID,
		defaultString(strings.TrimSpace(input.AuthorName), "注册用户"),
		defaultString(strings.TrimSpace(input.CoverImage), "https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=1400&q=80"),
		estimateReadingTime(content),
		publishedAt,
	).Scan(&postID)
	if err != nil {
		return Post{}, fmt.Errorf("insert published post: %w", err)
	}

	for _, tag := range normalizeTags(input.Tags) {
		tagID, err := ensureTag(ctx, tx, tag)
		if err != nil {
			return Post{}, err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO post_tags (post_id, tag_id)
			VALUES ($1::uuid, $2::uuid)
			ON CONFLICT DO NOTHING
		`, postID, tagID); err != nil {
			return Post{}, fmt.Errorf("insert post tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Post{}, fmt.Errorf("commit publish transaction: %w", err)
	}
	committed = true

	return repo.GetBySlug(ctx, slug)
}

func (repo *SQLRepository) count(ctx context.Context, keyword string, category string, tag string) (int, error) {
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
				input.keyword IS NULL
				OR p.search_vector @@ input.tsquery
				OR p.title ILIKE '%' || input.keyword || '%'
				OR p.summary ILIKE '%' || input.keyword || '%'
				OR p.content ILIKE '%' || input.keyword || '%'
				OR c.name ILIKE '%' || input.keyword || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags search_pt
					JOIN tags search_t ON search_t.id = search_pt.tag_id
					WHERE search_pt.post_id = p.id
						AND search_t.name ILIKE '%' || input.keyword || '%'
				)
			)
	`, keyword, category, tag).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count posts: %w", err)
	}

	return total, nil
}

func ensureCategory(ctx context.Context, tx *sql.Tx, category string) (string, error) {
	var categoryID string
	err := tx.QueryRowContext(ctx, `
		SELECT id::text
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
		INSERT INTO categories (slug, name)
		VALUES ($1, $2)
		RETURNING id::text
	`, slug, category).Scan(&categoryID)
	if err != nil {
		return "", fmt.Errorf("insert category: %w", err)
	}

	return categoryID, nil
}

func ensureTag(ctx context.Context, tx *sql.Tx, tag string) (string, error) {
	var tagID string
	err := tx.QueryRowContext(ctx, `
		SELECT id::text
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
		INSERT INTO tags (slug, name)
		VALUES ($1, $2)
		RETURNING id::text
	`, slug, tag).Scan(&tagID)
	if err != nil {
		return "", fmt.Errorf("insert tag: %w", err)
	}

	return tagID, nil
}

func uniqueSQLSlug(ctx context.Context, tx *sql.Tx, slug string) (string, error) {
	candidate := slug
	for suffix := 2; ; suffix++ {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM posts WHERE slug = $1)`, candidate).Scan(&exists); err != nil {
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

func scanPost(scanner postScanner) (Post, error) {
	var post Post
	var tagsCSV string

	if err := scanner.Scan(
		&post.ID,
		&post.Slug,
		&post.Title,
		&post.Summary,
		&post.Content,
		&post.Category,
		&tagsCSV,
		&post.CoverImage,
		&post.AuthorName,
		&post.ReadingTime,
		&post.ViewCount,
		&post.LikeCount,
		&post.DislikeCount,
		&post.CommentCount,
		&post.PublishedAt,
	); err != nil {
		return Post{}, fmt.Errorf("scan post: %w", err)
	}

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
