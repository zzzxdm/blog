package topics

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type SQLRepository struct {
	db  *sql.DB
	now func() time.Time
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db, now: time.Now}
}

func (repo *SQLRepository) List(ctx context.Context, query ListQuery) (ListResult, error) {
	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	offset := (page - 1) * pageSize
	status := strings.TrimSpace(query.Status)
	keyword := strings.TrimSpace(query.Keyword)

	total, err := repo.count(ctx, query.All, status, query.Featured, keyword)
	if err != nil {
		return ListResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
				CAST(id AS TEXT),
			slug,
			title,
			summary,
			cover_image,
			image_alt,
			tone,
			status,
			featured,
			sort_order,
				CAST(categories AS TEXT),
				CAST(tags AS TEXT),
			created_at,
			updated_at
			FROM topics
				WHERE ($1 OR status = 'active')
				  AND ($2 = '' OR status = $2)
				  AND (NOT $3 OR featured)
				  AND (
					$4 = ''
					OR lower(slug) LIKE '%' || lower($4) || '%'
					OR lower(title) LIKE '%' || lower($4) || '%'
					OR lower(summary) LIKE '%' || lower($4) || '%'
					OR lower(image_alt) LIKE '%' || lower($4) || '%'
					OR lower(CAST(categories AS TEXT)) LIKE '%' || lower($4) || '%'
					OR lower(CAST(tags AS TEXT)) LIKE '%' || lower($4) || '%'
				  )
			ORDER BY sort_order ASC, title ASC
			LIMIT $5 OFFSET $6
		`, query.All, status, query.Featured, keyword, pageSize, offset)
	if err != nil {
		return ListResult{}, fmt.Errorf("query topics: %w", err)
	}
	defer rows.Close()

	items := make([]Topic, 0, pageSize)
	for rows.Next() {
		item, err := scanTopic(rows)
		if err != nil {
			return ListResult{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ListResult{}, fmt.Errorf("iterate topics: %w", err)
	}

	return ListResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (repo *SQLRepository) GetBySlug(ctx context.Context, slug string) (Topic, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
				CAST(id AS TEXT),
			slug,
			title,
			summary,
			cover_image,
			image_alt,
			tone,
			status,
			featured,
			sort_order,
			CAST(categories AS TEXT),
			CAST(tags AS TEXT),
			created_at,
			updated_at
		FROM topics
		WHERE lower(slug) = lower($1)
	`, strings.TrimSpace(slug))

	item, err := scanTopic(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Topic{}, ErrNotFound
		}
		return Topic{}, err
	}

	return item, nil
}

func (repo *SQLRepository) Save(ctx context.Context, id string, request SaveRequest) (Topic, error) {
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return Topic{}, ErrInvalid
	}

	slug := repo.topicSlug(request.Slug, title)
	duplicate, err := repo.duplicate(ctx, id, slug, title)
	if err != nil {
		return Topic{}, err
	}
	if duplicate {
		return Topic{}, ErrDuplicate
	}

	categoriesJSON, err := json.Marshal(normalizeStrings(request.Categories))
	if err != nil {
		return Topic{}, fmt.Errorf("marshal topic categories: %w", err)
	}
	tagsJSON, err := json.Marshal(normalizeStrings(request.Tags))
	if err != nil {
		return Topic{}, fmt.Errorf("marshal topic tags: %w", err)
	}

	if strings.TrimSpace(id) == "" {
		var newID string
		err := repo.db.QueryRowContext(ctx, `
			INSERT INTO topics (
				slug, title, summary, cover_image, image_alt, tone, status, featured, sort_order, categories, tags
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING CAST(id AS TEXT)
		`,
			slug,
			title,
			strings.TrimSpace(request.Summary),
			strings.TrimSpace(request.CoverImage),
			strings.TrimSpace(request.ImageAlt),
			normalizeTone(request.Tone),
			normalizeStatus(request.Status),
			request.Featured,
			request.SortOrder,
			string(categoriesJSON),
			string(tagsJSON),
		).Scan(&newID)
		if err != nil {
			return Topic{}, fmt.Errorf("insert topic: %w", err)
		}

		return repo.getByID(ctx, newID)
	}

	var updatedID string
	err = repo.db.QueryRowContext(ctx, `
		UPDATE topics
		SET slug = $2,
		    title = $3,
		    summary = $4,
		    cover_image = $5,
		    image_alt = $6,
		    tone = $7,
		    status = $8,
		    featured = $9,
		    sort_order = $10,
		    categories = $11,
		    tags = $12
		WHERE CAST(id AS TEXT) = $1
		RETURNING CAST(id AS TEXT)
	`,
		id,
		slug,
		title,
		strings.TrimSpace(request.Summary),
		strings.TrimSpace(request.CoverImage),
		strings.TrimSpace(request.ImageAlt),
		normalizeTone(request.Tone),
		normalizeStatus(request.Status),
		request.Featured,
		request.SortOrder,
		string(categoriesJSON),
		string(tagsJSON),
	).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Topic{}, ErrNotFound
		}
		return Topic{}, fmt.Errorf("update topic: %w", err)
	}

	return repo.getByID(ctx, updatedID)
}

func (repo *SQLRepository) Delete(ctx context.Context, id string) error {
	result, err := repo.db.ExecContext(ctx, "DELETE FROM topics WHERE CAST(id AS TEXT) = $1", id)
	if err != nil {
		return fmt.Errorf("delete topic: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete topic rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (repo *SQLRepository) count(ctx context.Context, all bool, status string, featured bool, keyword string) (int, error) {
	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM topics
		WHERE ($1 OR status = 'active')
		  AND ($2 = '' OR status = $2)
		  AND (NOT $3 OR featured)
		  AND (
			$4 = ''
			OR lower(slug) LIKE '%' || lower($4) || '%'
			OR lower(title) LIKE '%' || lower($4) || '%'
			OR lower(summary) LIKE '%' || lower($4) || '%'
			OR lower(image_alt) LIKE '%' || lower($4) || '%'
			OR lower(CAST(categories AS TEXT)) LIKE '%' || lower($4) || '%'
			OR lower(CAST(tags AS TEXT)) LIKE '%' || lower($4) || '%'
		  )
	`, all, status, featured, strings.TrimSpace(keyword)).Scan(&total); err != nil {
		return 0, fmt.Errorf("count topics: %w", err)
	}

	return total, nil
}

func (repo *SQLRepository) getByID(ctx context.Context, id string) (Topic, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
				CAST(id AS TEXT),
			slug,
			title,
			summary,
			cover_image,
			image_alt,
			tone,
			status,
			featured,
			sort_order,
			CAST(categories AS TEXT),
			CAST(tags AS TEXT),
			created_at,
			updated_at
		FROM topics
		WHERE CAST(id AS TEXT) = $1
	`, id)

	item, err := scanTopic(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Topic{}, ErrNotFound
		}
		return Topic{}, err
	}

	return item, nil
}

func (repo *SQLRepository) duplicate(ctx context.Context, id string, slug string, title string) (bool, error) {
	var duplicate bool
	if err := repo.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM topics
			WHERE (lower(slug) = lower($1) OR lower(title) = lower($2))
			  AND ($3 = '' OR CAST(id AS TEXT) <> $3)
		)
	`, slug, title, strings.TrimSpace(id)).Scan(&duplicate); err != nil {
		return false, fmt.Errorf("check topic duplicate: %w", err)
	}

	return duplicate, nil
}

func (repo *SQLRepository) topicSlug(value string, title string) string {
	slug := defaultString(slugify(value), slugify(title))
	if slug == "" {
		slug = fmt.Sprintf("topic-%d", repo.now().UnixNano())
	}

	return slug
}

type topicScanner interface {
	Scan(dest ...any) error
}

func scanTopic(scanner topicScanner) (Topic, error) {
	var item Topic
	var categoriesJSON string
	var tagsJSON string
	if err := scanner.Scan(
		&item.ID,
		&item.Slug,
		&item.Title,
		&item.Summary,
		&item.CoverImage,
		&item.ImageAlt,
		&item.Tone,
		&item.Status,
		&item.Featured,
		&item.SortOrder,
		&categoriesJSON,
		&tagsJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return Topic{}, fmt.Errorf("scan topic: %w", err)
	}

	item.Categories = decodeStringList(categoriesJSON)
	item.Tags = decodeStringList(tagsJSON)
	return item, nil
}

func decodeStringList(value string) []string {
	var items []string
	if err := json.Unmarshal([]byte(value), &items); err != nil {
		return []string{}
	}
	return normalizeStrings(items)
}
