package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blog/api/internal/config"
	"blog/api/internal/database"
	"blog/api/internal/modules/posts"
	appserver "blog/api/internal/server"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	router := appserver.NewRouter(cfg)

	db, err := database.Open(ctx, cfg)
	if err != nil {
		if cfg.AppEnv == "production" {
			slog.Error("database connection failed", "error", err)
			os.Exit(1)
		}

		slog.Warn("database unavailable, using in-memory repositories", "error", err)
	} else {
		defer db.Close()

		migrateCtx, cancelMigrate := context.WithTimeout(ctx, 20*time.Second)
		if err := database.Migrate(migrateCtx, db); err != nil {
			cancelMigrate()
			if cfg.AppEnv == "production" {
				slog.Error("database migration failed", "error", err)
				os.Exit(1)
			}

			slog.Warn("database migration failed, using in-memory repositories", "error", err)
		} else {
			cancelMigrate()
			router = appserver.NewRouterWithPostsRepository(cfg, posts.NewSQLRepository(db))
		}
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("api server starting", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("api server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("api server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("api server stopped")
}
