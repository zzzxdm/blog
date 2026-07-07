package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blog/api/internal/config"
	"blog/api/internal/database"
	"blog/api/internal/modules/adminposts"
	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/comments"
	"blog/api/internal/modules/messages"
	"blog/api/internal/modules/operations"
	"blog/api/internal/modules/posts"
	"blog/api/internal/modules/reactions"
	"blog/api/internal/modules/submissions"
	"blog/api/internal/modules/taxonomies"
	"blog/api/internal/modules/topics"
	"blog/api/internal/modules/users"
	appserver "blog/api/internal/server"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, repositories, err := setupRepositories(ctx, cfg)
	if err != nil {
		slog.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	router := appserver.NewRouterWithRepositories(cfg, repositories)

	schedulerCtx, stopScheduler := context.WithCancel(context.Background())
	defer stopScheduler()
	if publisher, ok := repositories.PostRepo.(posts.Publisher); ok {
		adminposts.StartScheduledPublisher(schedulerCtx, repositories.AdminPostRepo, publisher, time.Minute)
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
	stopScheduler()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("api server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("api server stopped")
}

func setupRepositories(ctx context.Context, cfg config.Config) (*sql.DB, appserver.Repositories, error) {
	dbType, err := database.NormalizeDBType(cfg.DBType)
	if err != nil {
		return nil, appserver.Repositories{}, err
	}

	db, err := database.Open(ctx, cfg)
	if err != nil {
		return nil, appserver.Repositories{}, fmt.Errorf("open %s database: %w", dbType, err)
	}

	repositories, initErr := initializeSQLRepositories(ctx, db, dbType == database.DBTypeSQLite)
	if initErr != nil {
		_ = db.Close()
		return nil, appserver.Repositories{}, fmt.Errorf("initialize %s database: %w", dbType, initErr)
	}

	if dbType == database.DBTypeSQLite {
		slog.Info("using sqlite database", "path", cfg.SQLitePath)
	} else {
		slog.Info("using postgres database")
	}
	return db, repositories, nil
}

func initializeSQLRepositories(ctx context.Context, db *sql.DB, sqlite bool) (appserver.Repositories, error) {
	migrateCtx, cancelMigrate := context.WithTimeout(ctx, 20*time.Second)
	var migrateErr error
	if sqlite {
		migrateErr = database.MigrateSQLite(migrateCtx, db)
	} else {
		migrateErr = database.Migrate(migrateCtx, db)
	}
	cancelMigrate()
	if migrateErr != nil {
		return appserver.Repositories{}, migrateErr
	}

	setupCtx, cancelSetup := context.WithTimeout(ctx, 30*time.Second)
	defer cancelSetup()

	authStore, authErr := auth.NewSQLStore(setupCtx, db)
	commentRepo, commentErr := comments.NewSQLRepository(setupCtx, db)
	reactionRepo, reactionErr := reactions.NewSQLRepository(setupCtx, db)
	messageRepo, messageErr := messages.NewSQLRepository(setupCtx, db)
	submissionRepo, submissionErr := submissions.NewSQLRepository(setupCtx, db)
	operationsRepo, operationsErr := operations.NewSQLRepository(setupCtx, db)
	userRepo, userErr := users.NewSQLRepository(setupCtx, db)
	adminPostRepo, adminPostErr := adminposts.NewSQLRepository(setupCtx, db)
	initErr := errors.Join(authErr, commentErr, reactionErr, messageErr, submissionErr, operationsErr, userErr, adminPostErr)
	if initErr != nil {
		return appserver.Repositories{}, initErr
	}

	return appserver.Repositories{
		AuthStore:      authStore,
		PostRepo:       posts.NewSQLRepository(db),
		CommentRepo:    commentRepo,
		ReactionRepo:   reactionRepo,
		MessageRepo:    messageRepo,
		SubmissionRepo: submissionRepo,
		OperationsRepo: operationsRepo,
		UserRepo:       userRepo,
		AdminPostRepo:  adminPostRepo,
		TaxonomyRepo:   taxonomies.NewSQLRepository(db),
		TopicRepo:      topics.NewSQLRepository(db),
	}, nil
}
