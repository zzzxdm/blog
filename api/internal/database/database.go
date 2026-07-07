package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"blog/api/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

const (
	DBTypeSQLite   = "sqlite"
	DBTypePostgres = "postgres"
)

func NormalizeDBType(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", DBTypeSQLite, "sqlite3":
		return DBTypeSQLite, nil
	case DBTypePostgres, "postgresql", "pg":
		return DBTypePostgres, nil
	default:
		return "", fmt.Errorf("unsupported database type %q", value)
	}
}

func Open(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	dbType, err := NormalizeDBType(cfg.DBType)
	if err != nil {
		return nil, err
	}

	switch dbType {
	case DBTypeSQLite:
		return OpenSQLite(ctx, cfg.SQLitePath)
	case DBTypePostgres:
		return OpenPostgres(ctx, cfg.DatabaseURL)
	default:
		return nil, fmt.Errorf("unsupported database type %q", dbType)
	}
}

func OpenPostgres(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func OpenSQLite(ctx context.Context, path string) (*sql.DB, error) {
	if path == "" {
		path = "data/blog.sqlite"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	dsn := "file:" + filepath.ToSlash(path) + "?_pragma=" + url.QueryEscape("foreign_keys(1)") + "&_pragma=" + url.QueryEscape("busy_timeout(5000)") + "&_time_format=sqlite"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func IsSQLite(db *sql.DB) bool {
	if db == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var version string
	return db.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version) == nil
}
