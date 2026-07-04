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

	return cloneSettings(settings), nil
}

func (repo *SQLRepository) UpdateSettings(ctx context.Context, settings Settings) (Settings, error) {
	settings.UpdatedAt = time.Now()
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

	return cloneNavigation(navigation), nil
}

func (repo *SQLRepository) UpdateNavigation(ctx context.Context, navigation Navigation) (Navigation, error) {
	navigation.UpdatedAt = time.Now()
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

func (repo *SQLRepository) GetStats(_ context.Context) (Stats, error) {
	return cloneStats(seedStats()), nil
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
