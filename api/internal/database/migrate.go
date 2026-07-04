package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version text PRIMARY KEY,
			name text NOT NULL,
			applied_at timestamptz NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := strings.TrimSuffix(entry.Name(), ".sql")
		applied, err := migrationApplied(ctx, db, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlText, err := migrationsFS.ReadFile(path.Join("migrations", entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		if err := applyMigration(ctx, db, version, entry.Name(), string(sqlText)); err != nil {
			return err
		}
	}

	return nil
}

func migrationApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var applied bool
	err := db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied)
	if err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}

	return applied, nil
}

func applyMigration(ctx context.Context, db *sql.DB, version string, name string, sqlText string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.ExecContext(ctx, sqlText); err != nil {
		return fmt.Errorf("apply migration %s: %w", version, err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version, name) VALUES ($1, $2)", version, name); err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}
	tx = nil

	return nil
}
