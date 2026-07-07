package reactions

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

func NewSQLRepository(ctx context.Context, db *sql.DB) (*SQLRepository, error) {
	repo := &SQLRepository{db: db}
	if err := repo.ensureSeedInteractions(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *SQLRepository) Get(ctx context.Context, postSlug string, userID string) (Summary, error) {
	if err := repo.ensureStats(ctx, postSlug); err != nil {
		return Summary{}, err
	}

	var summary Summary
	var myReaction sql.NullString
	var bookmarked bool
	err := repo.db.QueryRowContext(ctx, `
		SELECT
			stats.post_slug,
			stats.like_count,
			stats.dislike_count,
			stats.bookmark_count,
			reaction.reaction,
			EXISTS (
				SELECT 1
				FROM post_bookmarks bookmark
				WHERE bookmark.post_slug = stats.post_slug
					AND bookmark.user_id = $2
			) AS bookmarked
		FROM post_interaction_stats stats
		LEFT JOIN post_reactions reaction
			ON reaction.post_slug = stats.post_slug
			AND reaction.user_id = $2
		WHERE stats.post_slug = $1
	`, postSlug, userID).Scan(
		&summary.PostSlug,
		&summary.LikeCount,
		&summary.DislikeCount,
		&summary.BookmarkCount,
		&myReaction,
		&bookmarked,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Summary{PostSlug: postSlug}, nil
		}
		return Summary{}, fmt.Errorf("query reaction summary: %w", err)
	}

	summary.MyReaction = myReaction.String
	summary.Bookmarked = bookmarked

	return summary, nil
}

func (repo *SQLRepository) SetReaction(ctx context.Context, postSlug string, userID string, reaction string) (Summary, error) {
	reaction = strings.ToLower(strings.TrimSpace(reaction))
	if reaction != "" && reaction != "like" && reaction != "dislike" {
		return Summary{}, ErrInvalidReaction
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Summary{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := ensureStatsTx(ctx, tx, postSlug); err != nil {
		return Summary{}, err
	}

	previous, err := currentReaction(ctx, tx, postSlug, userID)
	if err != nil {
		return Summary{}, err
	}

	if previous == reaction {
		reaction = ""
	}

	likeDelta, dislikeDelta := reactionDeltas(previous, reaction)

	if reaction == "" {
		if _, err := tx.ExecContext(ctx, `DELETE FROM post_reactions WHERE post_slug = $1 AND user_id = $2`, postSlug, userID); err != nil {
			return Summary{}, fmt.Errorf("delete reaction: %w", err)
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO post_reactions (post_slug, user_id, reaction)
			VALUES ($1, $2, $3)
			ON CONFLICT (post_slug, user_id) DO UPDATE SET
				reaction = EXCLUDED.reaction,
				updated_at = $4
		`, postSlug, userID, reaction, time.Now()); err != nil {
			return Summary{}, fmt.Errorf("upsert reaction: %w", err)
		}
	}

	if err := updateReactionCounts(ctx, tx, postSlug, likeDelta, dislikeDelta); err != nil {
		return Summary{}, err
	}

	if err := tx.Commit(); err != nil {
		return Summary{}, err
	}
	committed = true

	return repo.Get(ctx, postSlug, userID)
}

func (repo *SQLRepository) SetBookmark(ctx context.Context, postSlug string, userID string, bookmarked bool) (Summary, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Summary{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := ensureStatsTx(ctx, tx, postSlug); err != nil {
		return Summary{}, err
	}

	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM post_bookmarks WHERE post_slug = $1 AND user_id = $2
		)
	`, postSlug, userID).Scan(&exists); err != nil {
		return Summary{}, fmt.Errorf("query bookmark: %w", err)
	}

	delta := 0
	if bookmarked && !exists {
		delta = 1
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO post_bookmarks (post_slug, user_id)
			VALUES ($1, $2)
		`, postSlug, userID); err != nil {
			return Summary{}, fmt.Errorf("insert bookmark: %w", err)
		}
	}
	if !bookmarked && exists {
		delta = -1
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM post_bookmarks WHERE post_slug = $1 AND user_id = $2
		`, postSlug, userID); err != nil {
			return Summary{}, fmt.Errorf("delete bookmark: %w", err)
		}
	}

	if delta != 0 {
		if _, err := tx.ExecContext(ctx, `
			UPDATE post_interaction_stats
			SET bookmark_count = CASE WHEN bookmark_count + $2 < 0 THEN 0 ELSE bookmark_count + $2 END
			WHERE post_slug = $1
		`, postSlug, delta); err != nil {
			return Summary{}, fmt.Errorf("update bookmark count: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Summary{}, err
	}
	committed = true

	return repo.Get(ctx, postSlug, userID)
}

func (repo *SQLRepository) ListBookmarks(ctx context.Context, userID string, query BookmarkQuery) (BookmarkPage, error) {
	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	offset := (page - 1) * pageSize
	keyword := strings.TrimSpace(query.Keyword)
	category := strings.TrimSpace(query.Category)
	sortMode := strings.ToLower(strings.TrimSpace(query.Sort))
	if sortMode != "published" && sortMode != "views" {
		sortMode = "bookmarked"
	}

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM post_bookmarks bookmark
		JOIN posts post ON post.slug = bookmark.post_slug
		JOIN categories category ON category.id = post.category_id
		WHERE bookmark.user_id = $1
			AND post.status = 'published'
			AND ($2 = '' OR lower(category.slug) = lower($2) OR lower(category.name) = lower($2))
			AND (
				$3 = ''
				OR lower(post.title) LIKE '%' || lower($3) || '%'
				OR lower(post.summary) LIKE '%' || lower($3) || '%'
				OR lower(post.content) LIKE '%' || lower($3) || '%'
				OR lower(category.name) LIKE '%' || lower($3) || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags post_tag
					JOIN tags tag ON tag.id = post_tag.tag_id
					WHERE post_tag.post_id = post.id
						AND lower(tag.name) LIKE '%' || lower($3) || '%'
				)
			)
	`, userID, category, keyword).Scan(&total); err != nil {
		return BookmarkPage{}, fmt.Errorf("count bookmarks: %w", err)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT bookmark.post_slug, bookmark.created_at
		FROM post_bookmarks bookmark
		JOIN posts post ON post.slug = bookmark.post_slug
		JOIN categories category ON category.id = post.category_id
		WHERE bookmark.user_id = $1
			AND post.status = 'published'
			AND ($2 = '' OR lower(category.slug) = lower($2) OR lower(category.name) = lower($2))
			AND (
				$3 = ''
				OR lower(post.title) LIKE '%' || lower($3) || '%'
				OR lower(post.summary) LIKE '%' || lower($3) || '%'
				OR lower(post.content) LIKE '%' || lower($3) || '%'
				OR lower(category.name) LIKE '%' || lower($3) || '%'
				OR EXISTS (
					SELECT 1
					FROM post_tags post_tag
					JOIN tags tag ON tag.id = post_tag.tag_id
					WHERE post_tag.post_id = post.id
						AND lower(tag.name) LIKE '%' || lower($3) || '%'
				)
			)
		ORDER BY
			CASE WHEN $4 = 'published' THEN post.published_at END DESC,
			CASE WHEN $4 = 'views' THEN post.view_count END DESC,
			bookmark.created_at DESC
		LIMIT $5 OFFSET $6
	`, userID, category, keyword, sortMode, pageSize, offset)
	if err != nil {
		return BookmarkPage{}, fmt.Errorf("query bookmarks: %w", err)
	}
	defer rows.Close()

	items := make([]Bookmark, 0, pageSize)
	for rows.Next() {
		var bookmark Bookmark
		if err := rows.Scan(&bookmark.PostSlug, &bookmark.BookmarkedAt); err != nil {
			return BookmarkPage{}, fmt.Errorf("scan bookmark: %w", err)
		}
		items = append(items, bookmark)
	}
	if err := rows.Err(); err != nil {
		return BookmarkPage{}, fmt.Errorf("iterate bookmarks: %w", err)
	}

	return BookmarkPage{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 10
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

func (repo *SQLRepository) ensureStats(ctx context.Context, postSlug string) error {
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO post_interaction_stats (post_slug, like_count, dislike_count, bookmark_count)
		SELECT
			slug,
			like_count,
			dislike_count,
			(
					SELECT count(*)
				FROM post_bookmarks bookmark
				WHERE bookmark.post_slug = posts.slug
			)
		FROM posts
		WHERE slug = $1
		ON CONFLICT (post_slug) DO UPDATE SET
			like_count = EXCLUDED.like_count,
			dislike_count = EXCLUDED.dislike_count,
			bookmark_count = EXCLUDED.bookmark_count
	`, postSlug)
	if err != nil {
		return fmt.Errorf("ensure interaction stats: %w", err)
	}

	return nil
}

func ensureStatsTx(ctx context.Context, tx *sql.Tx, postSlug string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO post_interaction_stats (post_slug, like_count, dislike_count, bookmark_count)
		SELECT
			slug,
			like_count,
			dislike_count,
			(
					SELECT count(*)
				FROM post_bookmarks bookmark
				WHERE bookmark.post_slug = posts.slug
			)
		FROM posts
		WHERE slug = $1
		ON CONFLICT (post_slug) DO UPDATE SET
			like_count = EXCLUDED.like_count,
			dislike_count = EXCLUDED.dislike_count,
			bookmark_count = EXCLUDED.bookmark_count
	`, postSlug)
	if err != nil {
		return fmt.Errorf("ensure interaction stats: %w", err)
	}

	return nil
}

func currentReaction(ctx context.Context, tx *sql.Tx, postSlug string, userID string) (string, error) {
	var reaction string
	err := tx.QueryRowContext(ctx, `
		SELECT reaction
		FROM post_reactions
		WHERE post_slug = $1
			AND user_id = $2
	`, postSlug, userID).Scan(&reaction)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("query current reaction: %w", err)
	}

	return reaction, nil
}

func reactionDeltas(previous string, next string) (int, int) {
	likeDelta := 0
	dislikeDelta := 0

	if previous == "like" {
		likeDelta--
	}
	if previous == "dislike" {
		dislikeDelta--
	}
	if next == "like" {
		likeDelta++
	}
	if next == "dislike" {
		dislikeDelta++
	}

	return likeDelta, dislikeDelta
}

func updateReactionCounts(ctx context.Context, tx *sql.Tx, postSlug string, likeDelta int, dislikeDelta int) error {
	if likeDelta == 0 && dislikeDelta == 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
			UPDATE post_interaction_stats
			SET
				like_count = CASE WHEN like_count + $2 < 0 THEN 0 ELSE like_count + $2 END,
				dislike_count = CASE WHEN dislike_count + $3 < 0 THEN 0 ELSE dislike_count + $3 END
			WHERE post_slug = $1
	`, postSlug, likeDelta, dislikeDelta); err != nil {
		return fmt.Errorf("update interaction counts: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
			UPDATE posts
			SET
				like_count = CASE WHEN like_count + $2 < 0 THEN 0 ELSE like_count + $2 END,
				dislike_count = CASE WHEN dislike_count + $3 < 0 THEN 0 ELSE dislike_count + $3 END
			WHERE slug = $1
	`, postSlug, likeDelta, dislikeDelta); err != nil {
		return fmt.Errorf("update post counts: %w", err)
	}

	return nil
}

func (repo *SQLRepository) ensureSeedInteractions(ctx context.Context) error {
	seeds := []struct {
		PostSlug string
		UserID   string
		Reaction string
	}{
		{PostSlug: "blog-system-design", UserID: "user_linyi", Reaction: "like"},
	}

	for _, seed := range seeds {
		if _, err := repo.db.ExecContext(ctx, `
			INSERT INTO post_reactions (post_slug, user_id, reaction)
			VALUES ($1, $2, $3)
			ON CONFLICT (post_slug, user_id) DO NOTHING
		`, seed.PostSlug, seed.UserID, seed.Reaction); err != nil {
			return fmt.Errorf("seed reaction %s/%s: %w", seed.PostSlug, seed.UserID, err)
		}
	}

	now := time.Now()
	bookmarkSeeds := []struct {
		PostSlug  string
		UserID    string
		CreatedAt time.Time
	}{
		{PostSlug: "blog-system-design", UserID: "user_linyi", CreatedAt: now.Add(-2 * time.Hour)},
		{PostSlug: "vue3-content-site-cache-seo", UserID: "user_linyi", CreatedAt: now.Add(-26 * time.Hour)},
	}

	for _, seed := range bookmarkSeeds {
		if _, err := repo.db.ExecContext(ctx, `
			INSERT INTO post_bookmarks (post_slug, user_id, created_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (post_slug, user_id) DO NOTHING
		`, seed.PostSlug, seed.UserID, seed.CreatedAt); err != nil {
			return fmt.Errorf("seed bookmark %s/%s: %w", seed.PostSlug, seed.UserID, err)
		}
	}

	return nil
}
