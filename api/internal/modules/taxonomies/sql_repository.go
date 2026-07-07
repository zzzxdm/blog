package taxonomies

import (
	"context"
	"database/sql"
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
	return &SQLRepository{
		db:  db,
		now: time.Now,
	}
}

func (repo *SQLRepository) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
				CAST(c.id AS TEXT),
			c.slug,
			c.name,
			c.description,
			c.sort_order,
			COUNT(p.id) AS post_count
		FROM categories c
		LEFT JOIN posts p ON p.category_id = c.id AND p.status = 'published'
		GROUP BY c.id
		ORDER BY c.sort_order ASC, c.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		item, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}

func (repo *SQLRepository) SaveCategory(ctx context.Context, id string, request SaveCategoryRequest) (Category, error) {
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return Category{}, ErrInvalid
	}

	slug := repo.categorySlug(request.Slug, name)
	duplicate, err := repo.categoryDuplicate(ctx, id, slug, name)
	if err != nil {
		return Category{}, err
	}
	if duplicate {
		return Category{}, ErrDuplicate
	}

	if strings.TrimSpace(id) == "" {
		var newID string
		err := repo.db.QueryRowContext(ctx, `
			INSERT INTO categories (slug, name, description, sort_order)
			VALUES ($1, $2, $3, $4)
				RETURNING CAST(id AS TEXT)
		`, slug, name, strings.TrimSpace(request.Description), request.SortOrder).Scan(&newID)
		if err != nil {
			return Category{}, fmt.Errorf("insert category: %w", err)
		}

		return repo.getCategory(ctx, newID)
	}

	var updatedID string
	err = repo.db.QueryRowContext(ctx, `
		UPDATE categories
		SET slug = $2,
		    name = $3,
		    description = $4,
		    sort_order = $5
			WHERE CAST(id AS TEXT) = $1
			RETURNING CAST(id AS TEXT)
	`, id, slug, name, strings.TrimSpace(request.Description), request.SortOrder).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Category{}, ErrNotFound
		}

		return Category{}, fmt.Errorf("update category: %w", err)
	}

	return repo.getCategory(ctx, updatedID)
}

func (repo *SQLRepository) DeleteCategory(ctx context.Context, id string) error {
	var postCount int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM posts WHERE CAST(category_id AS TEXT) = $1", id).Scan(&postCount); err != nil {
		return fmt.Errorf("count category posts: %w", err)
	}
	if postCount > 0 {
		return ErrTaxonomyInUse
	}

	result, err := repo.db.ExecContext(ctx, "DELETE FROM categories WHERE CAST(id AS TEXT) = $1", id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete category rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (repo *SQLRepository) ListTags(ctx context.Context) ([]Tag, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
				CAST(t.id AS TEXT),
			t.slug,
			t.name,
			COUNT(p.id) AS post_count
		FROM tags t
		LEFT JOIN post_tags pt ON pt.tag_id = t.id
		LEFT JOIN posts p ON p.id = pt.post_id AND p.status = 'published'
		GROUP BY t.id
		ORDER BY post_count DESC, t.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	tags := make([]Tag, 0)
	for rows.Next() {
		item, err := scanTag(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}

	return tags, nil
}

func (repo *SQLRepository) SaveTag(ctx context.Context, id string, request SaveTagRequest) (Tag, error) {
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return Tag{}, ErrInvalid
	}

	slug := repo.tagSlug(request.Slug, name)
	duplicate, err := repo.tagDuplicate(ctx, id, slug, name)
	if err != nil {
		return Tag{}, err
	}
	if duplicate {
		return Tag{}, ErrDuplicate
	}

	if strings.TrimSpace(id) == "" {
		var newID string
		err := repo.db.QueryRowContext(ctx, `
			INSERT INTO tags (slug, name)
			VALUES ($1, $2)
				RETURNING CAST(id AS TEXT)
		`, slug, name).Scan(&newID)
		if err != nil {
			return Tag{}, fmt.Errorf("insert tag: %w", err)
		}

		return repo.getTag(ctx, newID)
	}

	var updatedID string
	err = repo.db.QueryRowContext(ctx, `
		UPDATE tags
		SET slug = $2,
		    name = $3
			WHERE CAST(id AS TEXT) = $1
			RETURNING CAST(id AS TEXT)
	`, id, slug, name).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tag{}, ErrNotFound
		}

		return Tag{}, fmt.Errorf("update tag: %w", err)
	}

	return repo.getTag(ctx, updatedID)
}

func (repo *SQLRepository) DeleteTag(ctx context.Context, id string) error {
	var postCount int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM post_tags WHERE CAST(tag_id AS TEXT) = $1", id).Scan(&postCount); err != nil {
		return fmt.Errorf("count tag posts: %w", err)
	}
	if postCount > 0 {
		return ErrTaxonomyInUse
	}

	result, err := repo.db.ExecContext(ctx, "DELETE FROM tags WHERE CAST(id AS TEXT) = $1", id)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete tag rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (repo *SQLRepository) getCategory(ctx context.Context, id string) (Category, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
				CAST(c.id AS TEXT),
			c.slug,
			c.name,
			c.description,
			c.sort_order,
			COUNT(p.id) AS post_count
		FROM categories c
		LEFT JOIN posts p ON p.category_id = c.id AND p.status = 'published'
		WHERE CAST(c.id AS TEXT) = $1
		GROUP BY c.id
	`, id)

	item, err := scanCategory(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Category{}, ErrNotFound
		}

		return Category{}, err
	}

	return item, nil
}

func (repo *SQLRepository) getTag(ctx context.Context, id string) (Tag, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
				CAST(t.id AS TEXT),
			t.slug,
			t.name,
			COUNT(p.id) AS post_count
		FROM tags t
		LEFT JOIN post_tags pt ON pt.tag_id = t.id
		LEFT JOIN posts p ON p.id = pt.post_id AND p.status = 'published'
		WHERE CAST(t.id AS TEXT) = $1
		GROUP BY t.id
	`, id)

	item, err := scanTag(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tag{}, ErrNotFound
		}

		return Tag{}, err
	}

	return item, nil
}

func (repo *SQLRepository) categoryDuplicate(ctx context.Context, id string, slug string, name string) (bool, error) {
	var duplicate bool
	if err := repo.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM categories
			WHERE (lower(slug) = lower($1) OR lower(name) = lower($2))
				  AND ($3 = '' OR CAST(id AS TEXT) <> $3)
		)
	`, slug, name, strings.TrimSpace(id)).Scan(&duplicate); err != nil {
		return false, fmt.Errorf("check category duplicate: %w", err)
	}

	return duplicate, nil
}

func (repo *SQLRepository) tagDuplicate(ctx context.Context, id string, slug string, name string) (bool, error) {
	var duplicate bool
	if err := repo.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM tags
			WHERE (lower(slug) = lower($1) OR lower(name) = lower($2))
				  AND ($3 = '' OR CAST(id AS TEXT) <> $3)
		)
	`, slug, name, strings.TrimSpace(id)).Scan(&duplicate); err != nil {
		return false, fmt.Errorf("check tag duplicate: %w", err)
	}

	return duplicate, nil
}

func (repo *SQLRepository) categorySlug(value string, name string) string {
	slug := defaultString(slugify(value), slugify(name))
	if slug == "" {
		slug = fmt.Sprintf("category-%d", repo.now().UnixNano())
	}

	return slug
}

func (repo *SQLRepository) tagSlug(value string, name string) string {
	slug := defaultString(slugify(value), slugify(name))
	if slug == "" {
		slug = fmt.Sprintf("tag-%d", repo.now().UnixNano())
	}

	return slug
}

func scanCategory(scanner interface{ Scan(dest ...any) error }) (Category, error) {
	var item Category
	if err := scanner.Scan(&item.ID, &item.Slug, &item.Name, &item.Description, &item.SortOrder, &item.PostCount); err != nil {
		return Category{}, fmt.Errorf("scan category: %w", err)
	}

	return item, nil
}

func scanTag(scanner interface{ Scan(dest ...any) error }) (Tag, error) {
	var item Tag
	if err := scanner.Scan(&item.ID, &item.Slug, &item.Name, &item.PostCount); err != nil {
		return Tag{}, fmt.Errorf("scan tag: %w", err)
	}

	return item, nil
}
