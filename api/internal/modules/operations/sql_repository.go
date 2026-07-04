package operations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

func (repo *SQLRepository) GetStats(_ context.Context) (Stats, error) {
	return cloneStats(seedStats()), nil
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
