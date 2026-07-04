package operations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	settingsDocumentKey   = "settings"
	navigationDocumentKey = "navigation"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(ctx context.Context, db *sql.DB) (*SQLRepository, error) {
	repo := &SQLRepository{db: db}
	if err := repo.ensureSeedData(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *SQLRepository) GetSettings(ctx context.Context) (Settings, error) {
	var settings Settings
	if err := repo.getDocument(ctx, settingsDocumentKey, &settings); err != nil {
		return Settings{}, err
	}

	return normalizeSettings(settings), nil
}

func (repo *SQLRepository) UpdateSettings(ctx context.Context, settings Settings) (Settings, error) {
	settings = settingsForUpdate(settings, time.Now())
	if err := repo.saveDocument(ctx, settingsDocumentKey, settings); err != nil {
		return Settings{}, err
	}

	return cloneSettings(settings), nil
}

func (repo *SQLRepository) SendTestMail(ctx context.Context) (TestMailResult, error) {
	settings, err := repo.GetSettings(ctx)
	if err != nil {
		return TestMailResult{}, err
	}

	return testMailResult(settings, time.Now()), nil
}

func (repo *SQLRepository) RunBackup(ctx context.Context) (BackupResult, error) {
	settings, err := repo.GetSettings(ctx)
	if err != nil {
		return BackupResult{}, err
	}

	now := time.Now()
	settings = normalizeSettings(settings)
	settings.LastBackupAt = now
	settings.UpdatedAt = now
	if err := repo.saveDocument(ctx, settingsDocumentKey, settings); err != nil {
		return BackupResult{}, err
	}

	return backupResult(settings, now), nil
}

func (repo *SQLRepository) GetNavigation(ctx context.Context) (Navigation, error) {
	var navigation Navigation
	if err := repo.getDocument(ctx, navigationDocumentKey, &navigation); err != nil {
		return Navigation{}, err
	}

	return normalizeNavigation(navigation), nil
}

func (repo *SQLRepository) UpdateNavigation(ctx context.Context, navigation Navigation) (Navigation, error) {
	navigation = navigationForUpdate(navigation, time.Now())
	if err := repo.saveDocument(ctx, navigationDocumentKey, navigation); err != nil {
		return Navigation{}, err
	}

	return cloneNavigation(navigation), nil
}

func (repo *SQLRepository) ListMedia(ctx context.Context) (MediaListResult, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, file_name, url, alt, type, category, size_label, width, height, usage_count, uploaded_by, uploaded_at
		FROM media_assets
		ORDER BY uploaded_at DESC, id DESC
	`)
	if err != nil {
		return MediaListResult{}, fmt.Errorf("query media assets: %w", err)
	}
	defer rows.Close()

	items := make([]MediaAsset, 0)
	for rows.Next() {
		var asset MediaAsset
		if err := rows.Scan(
			&asset.ID,
			&asset.FileName,
			&asset.URL,
			&asset.Alt,
			&asset.Type,
			&asset.Category,
			&asset.SizeLabel,
			&asset.Width,
			&asset.Height,
			&asset.UsageCount,
			&asset.UploadedBy,
			&asset.UploadedAt,
		); err != nil {
			return MediaListResult{}, fmt.Errorf("scan media asset: %w", err)
		}
		items = append(items, asset)
	}
	if err := rows.Err(); err != nil {
		return MediaListResult{}, fmt.Errorf("iterate media assets: %w", err)
	}

	return MediaListResult{
		Items: items,
		Total: len(items),
	}, nil
}

func (repo *SQLRepository) CreateMedia(ctx context.Context, asset MediaAsset) (MediaAsset, error) {
	if err := repo.insertMedia(ctx, asset, false); err != nil {
		return MediaAsset{}, err
	}

	return asset, nil
}

func (repo *SQLRepository) GetMedia(ctx context.Context, id string) (MediaAsset, error) {
	asset, err := repo.getMedia(ctx, id)
	if err != nil {
		return MediaAsset{}, err
	}

	return asset, nil
}

func (repo *SQLRepository) UpdateMedia(ctx context.Context, id string, request MediaUpdateRequest) (MediaAsset, error) {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE media_assets
		SET alt = $2, category = $3
		WHERE id = $1
	`, id, request.Alt, request.Category)
	if err != nil {
		return MediaAsset{}, fmt.Errorf("update media asset %s: %w", id, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return MediaAsset{}, fmt.Errorf("read media update rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return MediaAsset{}, ErrMediaNotFound
	}

	return repo.getMedia(ctx, id)
}

func (repo *SQLRepository) DeleteMedia(ctx context.Context, id string) (MediaAsset, error) {
	asset, err := repo.getMedia(ctx, id)
	if err != nil {
		return MediaAsset{}, err
	}
	if asset.UsageCount > 0 {
		return MediaAsset{}, ErrMediaInUse
	}

	if _, err := repo.db.ExecContext(ctx, "DELETE FROM media_assets WHERE id = $1", id); err != nil {
		return MediaAsset{}, fmt.Errorf("delete media asset %s: %w", id, err)
	}

	return asset, nil
}

func (repo *SQLRepository) GetStats(ctx context.Context, rangeKey string) (Stats, error) {
	spec := newStatsRangeSpec(rangeKey, time.Now())
	totals, err := repo.queryStatsTotals(ctx, spec)
	if err != nil {
		return Stats{}, err
	}

	trend, err := repo.queryStatsTrend(ctx, spec)
	if err != nil {
		return Stats{}, err
	}
	topPosts, err := repo.queryTopPosts(ctx, spec)
	if err != nil {
		return Stats{}, err
	}
	sources, err := repo.queryStatsSources(ctx, spec)
	if err != nil {
		return Stats{}, err
	}
	terms, err := repo.queryTopTags(ctx, spec)
	if err != nil {
		return Stats{}, err
	}
	suggestions, err := repo.queryStatsSuggestions(ctx, totals)
	if err != nil {
		return Stats{}, err
	}

	return Stats{
		Range:      spec.Key,
		RangeLabel: spec.Label,
		Metrics: []Metric{
			{Label: "阅读量", Value: formatStatNumber(totals.ViewCount), Delta: "当前范围累计"},
			{Label: "发布文章", Value: formatStatNumber(totals.PostCount), Delta: "已公开"},
			{Label: "平均阅读", Value: formatReadingMinutes(totals.AvgReading), Delta: "按阅读时长估算"},
			{Label: "互动数", Value: formatStatNumber(totals.InteractionCount()), Delta: fmt.Sprintf("评论 %s / 收藏 %s", formatStatNumber(totals.CommentCount), formatStatNumber(totals.BookmarkCount))},
		},
		Trend:       trend,
		TopPosts:    topPosts,
		Sources:     sources,
		SearchTerms: terms,
		Suggestions: suggestions,
	}, nil
}

type sqlStatsTotals struct {
	PostCount     int
	ViewCount     int
	AvgReading    float64
	CommentCount  int
	LikeCount     int
	DislikeCount  int
	BookmarkCount int
}

func (totals sqlStatsTotals) InteractionCount() int {
	return totals.CommentCount + totals.LikeCount + totals.DislikeCount + totals.BookmarkCount
}

func (repo *SQLRepository) queryStatsTotals(ctx context.Context, spec statsRangeSpec) (sqlStatsTotals, error) {
	var totals sqlStatsTotals
	err := repo.db.QueryRowContext(ctx, `
		SELECT
			count(*)::int,
			COALESCE(sum(p.view_count), 0)::int,
			COALESCE(avg(p.reading_time), 0)::float8,
			COALESCE(sum(p.comment_count), 0)::int,
			COALESCE(sum(p.like_count), 0)::int,
			COALESCE(sum(p.dislike_count), 0)::int,
			COALESCE(sum(COALESCE(interactions.bookmark_count, 0)), 0)::int
		FROM posts p
		LEFT JOIN post_interaction_stats interactions ON interactions.post_slug = p.slug
		WHERE p.status = 'published'
			AND COALESCE(p.published_at, p.created_at) >= $1
			AND COALESCE(p.published_at, p.created_at) < $2
	`, spec.Start, spec.End).Scan(
		&totals.PostCount,
		&totals.ViewCount,
		&totals.AvgReading,
		&totals.CommentCount,
		&totals.LikeCount,
		&totals.DislikeCount,
		&totals.BookmarkCount,
	)
	if err != nil {
		return sqlStatsTotals{}, fmt.Errorf("query stats totals: %w", err)
	}

	return totals, nil
}

func (repo *SQLRepository) queryStatsTrend(ctx context.Context, spec statsRangeSpec) ([]BarPoint, error) {
	bucket := spec.trendBucket()
	rows, err := repo.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT
			date_trunc('%s', COALESCE(p.published_at, p.created_at)) AS bucket,
			COALESCE(sum(p.view_count), 0)::int AS views
		FROM posts p
		WHERE p.status = 'published'
			AND COALESCE(p.published_at, p.created_at) >= $1
			AND COALESCE(p.published_at, p.created_at) < $2
		GROUP BY bucket
		ORDER BY bucket ASC
	`, bucket), spec.Start, spec.End)
	if err != nil {
		return nil, fmt.Errorf("query stats trend: %w", err)
	}
	defer rows.Close()

	type trendPoint struct {
		Bucket time.Time
		Views  int
	}

	points := make([]trendPoint, 0)
	maxViews := 0
	for rows.Next() {
		var point trendPoint
		if err := rows.Scan(&point.Bucket, &point.Views); err != nil {
			return nil, fmt.Errorf("scan stats trend: %w", err)
		}
		if point.Views > maxViews {
			maxViews = point.Views
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stats trend: %w", err)
	}
	if len(points) == 0 {
		return []BarPoint{{Label: "暂无", Value: "0", Percent: 0}}, nil
	}

	result := make([]BarPoint, 0, len(points))
	for _, point := range points {
		result = append(result, BarPoint{
			Label:   formatTrendLabel(point.Bucket, spec.Key),
			Value:   formatStatNumber(point.Views),
			Percent: barPercent(point.Views, maxViews),
		})
	}

	return result, nil
}

func (repo *SQLRepository) queryTopPosts(ctx context.Context, spec statsRangeSpec) ([]TopPost, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			p.title,
			p.view_count,
			COALESCE(interactions.bookmark_count, 0)::int AS bookmark_count,
			p.comment_count,
			(
				p.comment_count +
				p.like_count +
				p.dislike_count +
				COALESCE(interactions.bookmark_count, 0)
			)::int AS interaction_count
		FROM posts p
		LEFT JOIN post_interaction_stats interactions ON interactions.post_slug = p.slug
		WHERE p.status = 'published'
			AND COALESCE(p.published_at, p.created_at) >= $1
			AND COALESCE(p.published_at, p.created_at) < $2
		ORDER BY p.view_count DESC, interaction_count DESC, COALESCE(p.published_at, p.created_at) DESC
		LIMIT 5
	`, spec.Start, spec.End)
	if err != nil {
		return nil, fmt.Errorf("query top posts: %w", err)
	}
	defer rows.Close()

	items := make([]TopPost, 0)
	for rows.Next() {
		var title string
		var views int
		var bookmarks int
		var comments int
		var interactions int
		if err := rows.Scan(&title, &views, &bookmarks, &comments, &interactions); err != nil {
			return nil, fmt.Errorf("scan top post: %w", err)
		}

		items = append(items, TopPost{
			Title:          title,
			Views:          formatStatNumber(views),
			Bookmarks:      bookmarks,
			Comments:       comments,
			EngagementRate: formatRate(interactions, views),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top posts: %w", err)
	}

	return items, nil
}

func (repo *SQLRepository) queryStatsSources(ctx context.Context, spec statsRangeSpec) ([]BarPoint, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			p.source,
			count(*)::int AS post_count,
			COALESCE(sum(p.view_count), 0)::int AS views
		FROM posts p
		WHERE p.status = 'published'
			AND COALESCE(p.published_at, p.created_at) >= $1
			AND COALESCE(p.published_at, p.created_at) < $2
		GROUP BY p.source
		ORDER BY views DESC, post_count DESC
	`, spec.Start, spec.End)
	if err != nil {
		return nil, fmt.Errorf("query stats sources: %w", err)
	}
	defer rows.Close()

	type sourcePoint struct {
		Source    string
		PostCount int
		Views     int
	}

	points := make([]sourcePoint, 0)
	totalViews := 0
	totalPosts := 0
	for rows.Next() {
		var point sourcePoint
		if err := rows.Scan(&point.Source, &point.PostCount, &point.Views); err != nil {
			return nil, fmt.Errorf("scan stats source: %w", err)
		}
		totalViews += point.Views
		totalPosts += point.PostCount
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stats sources: %w", err)
	}
	if len(points) == 0 {
		return []BarPoint{{Label: "暂无内容", Value: "0%", Percent: 0}}, nil
	}

	total := totalViews
	if total == 0 {
		total = totalPosts
	}

	result := make([]BarPoint, 0, len(points))
	for _, point := range points {
		value := point.Views
		if totalViews == 0 {
			value = point.PostCount
		}

		result = append(result, BarPoint{
			Label:   statsSourceLabel(point.Source),
			Value:   formatRate(value, total),
			Percent: barPercent(value, total),
		})
	}

	return result, nil
}

func (repo *SQLRepository) queryTopTags(ctx context.Context, spec statsRangeSpec) ([]SearchTerm, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT t.name, count(*)::int AS post_count
		FROM posts p
		JOIN post_tags pt ON pt.post_id = p.id
		JOIN tags t ON t.id = pt.tag_id
		WHERE p.status = 'published'
			AND COALESCE(p.published_at, p.created_at) >= $1
			AND COALESCE(p.published_at, p.created_at) < $2
		GROUP BY t.name
		ORDER BY post_count DESC, t.name ASC
		LIMIT 5
	`, spec.Start, spec.End)
	if err != nil {
		return nil, fmt.Errorf("query top tags: %w", err)
	}
	defer rows.Close()

	items := make([]SearchTerm, 0)
	for rows.Next() {
		var item SearchTerm
		if err := rows.Scan(&item.Term, &item.Count); err != nil {
			return nil, fmt.Errorf("scan top tag: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top tags: %w", err)
	}

	return items, nil
}

func (repo *SQLRepository) queryStatsSuggestions(ctx context.Context, totals sqlStatsTotals) ([]ContentSuggestion, error) {
	pendingComments, err := repo.countRows(ctx, "SELECT count(*)::int FROM comments WHERE status = 'pending'")
	if err != nil {
		return nil, fmt.Errorf("count pending comments: %w", err)
	}
	pendingSubmissions, err := repo.countRows(ctx, "SELECT count(*)::int FROM submissions WHERE status = 'submitted'")
	if err != nil {
		return nil, fmt.Errorf("count pending submissions: %w", err)
	}

	suggestions := make([]ContentSuggestion, 0, 3)
	if pendingSubmissions > 0 {
		suggestions = append(suggestions, ContentSuggestion{
			Title: fmt.Sprintf("有 %s 篇投稿待审核", formatStatNumber(pendingSubmissions)),
			Body:  "优先处理投稿可以缩短用户从提交到发布的等待时间。",
		})
	}
	if pendingComments > 0 {
		suggestions = append(suggestions, ContentSuggestion{
			Title: fmt.Sprintf("有 %s 条评论待审核", formatStatNumber(pendingComments)),
			Body:  "及时审核评论可以让文章讨论保持连续。",
		})
	}
	if totals.PostCount == 0 {
		suggestions = append(suggestions, ContentSuggestion{
			Title: "当前范围没有新发布文章",
			Body:  "可以检查排期或把已完成草稿推进到发布流程。",
		})
	}
	if totals.ViewCount > 0 {
		suggestions = append(suggestions, ContentSuggestion{
			Title: "复盘阅读量最高的内容",
			Body:  "把高阅读文章补充到首页专题或相关链接中，延长长尾访问。",
		})
	}
	if len(suggestions) == 0 {
		suggestions = append(suggestions, ContentSuggestion{
			Title: "暂无待处理事项",
			Body:  "当前内容发布和互动状态平稳。",
		})
	}
	if len(suggestions) > 3 {
		return suggestions[:3], nil
	}

	return suggestions, nil
}

func (repo *SQLRepository) countRows(ctx context.Context, query string) (int, error) {
	var total int
	if err := repo.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

func statsSourceLabel(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "submission":
		return "用户投稿"
	case "admin":
		return "后台发布"
	default:
		return defaultString(source, "未知来源")
	}
}

func (repo *SQLRepository) ListAuditLogs(ctx context.Context, query AuditLogQuery) (AuditLogListResult, error) {
	query = normalizeAuditLogQuery(query)
	where := []string{"1 = 1"}
	args := make([]any, 0)

	if query.Action != "" {
		args = append(args, query.Action)
		where = append(where, fmt.Sprintf("action = $%d", len(args)))
	}
	if query.ResourceType != "" {
		args = append(args, query.ResourceType)
		where = append(where, fmt.Sprintf("resource_type = $%d", len(args)))
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT count(*) FROM audit_logs WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return AuditLogListResult{}, fmt.Errorf("count audit logs: %w", err)
	}

	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, actor_id, actor_name, action, resource_type, resource_id, resource_title, status, ip, user_agent, detail, created_at
		FROM audit_logs
		WHERE `+whereSQL+`
		ORDER BY created_at DESC, id DESC
		LIMIT $`+fmt.Sprint(len(args)-1)+` OFFSET $`+fmt.Sprint(len(args))+`
	`, args...)
	if err != nil {
		return AuditLogListResult{}, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	items := make([]AuditLog, 0)
	for rows.Next() {
		var item AuditLog
		if err := rows.Scan(
			&item.ID,
			&item.ActorID,
			&item.ActorName,
			&item.Action,
			&item.ResourceType,
			&item.ResourceID,
			&item.ResourceTitle,
			&item.Status,
			&item.IP,
			&item.UserAgent,
			&item.Detail,
			&item.CreatedAt,
		); err != nil {
			return AuditLogListResult{}, fmt.Errorf("scan audit log: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return AuditLogListResult{}, fmt.Errorf("iterate audit logs: %w", err)
	}

	return AuditLogListResult{
		Items:    items,
		Page:     query.Page,
		PageSize: query.PageSize,
		Total:    total,
	}, nil
}

func (repo *SQLRepository) RecordAuditLog(ctx context.Context, item AuditLog) error {
	return repo.insertAuditLog(ctx, normalizeAuditLog(item, time.Now()), false)
}

func (repo *SQLRepository) ensureSeedData(ctx context.Context) error {
	if err := repo.ensureDocument(ctx, settingsDocumentKey, seedSettings()); err != nil {
		return err
	}
	if err := repo.ensureDocument(ctx, navigationDocumentKey, seedNavigation()); err != nil {
		return err
	}

	for _, asset := range seedMedia() {
		if err := repo.insertMedia(ctx, asset, true); err != nil {
			return err
		}
	}
	for _, item := range seedAuditLogs() {
		if err := repo.insertAuditLog(ctx, item, true); err != nil {
			return err
		}
	}

	return nil
}

func (repo *SQLRepository) ensureDocument(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal %s document: %w", key, err)
	}

	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO operation_documents (key, data)
		VALUES ($1, $2)
		ON CONFLICT (key) DO NOTHING
	`, key, data); err != nil {
		return fmt.Errorf("seed %s document: %w", key, err)
	}

	return nil
}

func (repo *SQLRepository) getDocument(ctx context.Context, key string, target any) error {
	var data []byte
	if err := repo.db.QueryRowContext(ctx, "SELECT data FROM operation_documents WHERE key = $1", key).Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("%s document not found", key)
		}
		return fmt.Errorf("load %s document: %w", key, err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s document: %w", key, err)
	}

	return nil
}

func (repo *SQLRepository) saveDocument(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal %s document: %w", key, err)
	}

	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO operation_documents (key, data)
		VALUES ($1, $2)
		ON CONFLICT (key)
		DO UPDATE SET data = EXCLUDED.data
	`, key, data); err != nil {
		return fmt.Errorf("save %s document: %w", key, err)
	}

	return nil
}

func (repo *SQLRepository) insertMedia(ctx context.Context, asset MediaAsset, ignoreConflict bool) error {
	query := `
		INSERT INTO media_assets (
			id, file_name, url, alt, type, category, size_label, width, height, usage_count, uploaded_by, uploaded_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	if ignoreConflict {
		query += " ON CONFLICT (id) DO NOTHING"
	}

	if _, err := repo.db.ExecContext(ctx, query,
		asset.ID,
		asset.FileName,
		asset.URL,
		asset.Alt,
		asset.Type,
		asset.Category,
		asset.SizeLabel,
		asset.Width,
		asset.Height,
		asset.UsageCount,
		asset.UploadedBy,
		asset.UploadedAt,
	); err != nil {
		return fmt.Errorf("insert media asset %s: %w", asset.ID, err)
	}

	return nil
}

func (repo *SQLRepository) insertAuditLog(ctx context.Context, item AuditLog, ignoreConflict bool) error {
	query := `
		INSERT INTO audit_logs (
			id, actor_id, actor_name, action, resource_type, resource_id, resource_title,
			status, ip, user_agent, detail, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	if ignoreConflict {
		query += " ON CONFLICT (id) DO NOTHING"
	}

	if _, err := repo.db.ExecContext(ctx, query,
		item.ID,
		item.ActorID,
		item.ActorName,
		item.Action,
		item.ResourceType,
		item.ResourceID,
		item.ResourceTitle,
		item.Status,
		item.IP,
		item.UserAgent,
		item.Detail,
		item.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert audit log %s: %w", item.ID, err)
	}

	return nil
}

func (repo *SQLRepository) getMedia(ctx context.Context, id string) (MediaAsset, error) {
	var asset MediaAsset
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, file_name, url, alt, type, category, size_label, width, height, usage_count, uploaded_by, uploaded_at
		FROM media_assets
		WHERE id = $1
	`, id).Scan(
		&asset.ID,
		&asset.FileName,
		&asset.URL,
		&asset.Alt,
		&asset.Type,
		&asset.Category,
		&asset.SizeLabel,
		&asset.Width,
		&asset.Height,
		&asset.UsageCount,
		&asset.UploadedBy,
		&asset.UploadedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return MediaAsset{}, ErrMediaNotFound
		}

		return MediaAsset{}, fmt.Errorf("load media asset %s: %w", id, err)
	}

	return asset, nil
}
