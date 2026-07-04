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
	"blog/api/internal/modules/adminposts"
	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/comments"
	"blog/api/internal/modules/messages"
	"blog/api/internal/modules/operations"
	"blog/api/internal/modules/posts"
	"blog/api/internal/modules/reactions"
	"blog/api/internal/modules/submissions"
	"blog/api/internal/modules/taxonomies"
	"blog/api/internal/modules/users"
	appserver "blog/api/internal/server"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	var authStore auth.Store = auth.NewMemoryStore()
	var postRepo posts.Repository = posts.NewMemoryRepository()
	var commentRepo comments.Repository = comments.NewMemoryRepository()
	var reactionRepo reactions.Repository = reactions.NewMemoryRepository()
	var messageRepo messages.Repository = messages.NewMemoryRepository()
	var submissionRepo submissions.Repository = submissions.NewMemoryRepository()
	var operationsRepo operations.Repository = operations.NewMemoryRepository()
	var userRepo users.Repository = users.NewMemoryRepository()
	var adminPostRepo adminposts.Repository = adminposts.NewMemoryRepository()
	var taxonomyRepo taxonomies.Repository = taxonomies.NewMemoryRepository()

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
			setupCtx, cancelSetup := context.WithTimeout(ctx, 10*time.Second)
			sqlAuthStore, authErr := auth.NewSQLStore(setupCtx, db)
			sqlCommentRepo, commentErr := comments.NewSQLRepository(setupCtx, db)
			sqlReactionRepo, reactionErr := reactions.NewSQLRepository(setupCtx, db)
			sqlMessageRepo, messageErr := messages.NewSQLRepository(setupCtx, db)
			sqlSubmissionRepo, submissionErr := submissions.NewSQLRepository(setupCtx, db)
			sqlOperationsRepo, operationsErr := operations.NewSQLRepository(setupCtx, db)
			sqlUserRepo, userErr := users.NewSQLRepository(setupCtx, db)
			sqlAdminPostRepo, adminPostErr := adminposts.NewSQLRepository(setupCtx, db)
			cancelSetup()

			if authErr != nil || commentErr != nil || reactionErr != nil || messageErr != nil || submissionErr != nil || operationsErr != nil || userErr != nil || adminPostErr != nil {
				if cfg.AppEnv == "production" {
					slog.Error("database repository initialization failed", "auth", authErr, "comments", commentErr, "reactions", reactionErr, "messages", messageErr, "submissions", submissionErr, "operations", operationsErr, "users", userErr, "admin_posts", adminPostErr)
					os.Exit(1)
				}

				slog.Warn("database repository initialization failed, using in-memory repositories", "auth", authErr, "comments", commentErr, "reactions", reactionErr, "messages", messageErr, "submissions", submissionErr, "operations", operationsErr, "users", userErr, "admin_posts", adminPostErr)
			} else {
				authStore = sqlAuthStore
				postRepo = posts.NewSQLRepository(db)
				commentRepo = sqlCommentRepo
				reactionRepo = sqlReactionRepo
				messageRepo = sqlMessageRepo
				submissionRepo = sqlSubmissionRepo
				operationsRepo = sqlOperationsRepo
				userRepo = sqlUserRepo
				adminPostRepo = sqlAdminPostRepo
				taxonomyRepo = taxonomies.NewSQLRepository(db)
			}
		}
	}

	router := appserver.NewRouterWithRepositories(cfg, appserver.Repositories{
		AuthStore:      authStore,
		PostRepo:       postRepo,
		CommentRepo:    commentRepo,
		ReactionRepo:   reactionRepo,
		MessageRepo:    messageRepo,
		SubmissionRepo: submissionRepo,
		OperationsRepo: operationsRepo,
		UserRepo:       userRepo,
		AdminPostRepo:  adminPostRepo,
		TaxonomyRepo:   taxonomyRepo,
	})

	schedulerCtx, stopScheduler := context.WithCancel(context.Background())
	defer stopScheduler()
	if publisher, ok := postRepo.(posts.Publisher); ok {
		adminposts.StartScheduledPublisher(schedulerCtx, adminPostRepo, publisher, time.Minute)
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
