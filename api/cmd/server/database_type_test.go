package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"blog/api/internal/config"
	"blog/api/internal/database"
	"blog/api/internal/modules/adminposts"
	"blog/api/internal/modules/comments"
	"blog/api/internal/modules/messages"
	"blog/api/internal/modules/posts"
	"blog/api/internal/modules/submissions"
	"blog/api/internal/modules/topics"
	"blog/api/internal/modules/users"
)

func TestSetupRepositoriesUsesSQLiteByDefault(t *testing.T) {
	cfg := config.Config{
		AppEnv:     "development",
		SQLitePath: filepath.Join(t.TempDir(), "blog.sqlite"),
	}

	db, repositories, err := setupRepositories(context.Background(), cfg)
	if err != nil {
		t.Fatalf("setupRepositories returned error: %v", err)
	}
	defer db.Close()

	if !database.IsSQLite(db) {
		t.Fatal("expected sqlite database")
	}

	if _, _, err := repositories.AuthStore.Authenticate("admin@example.com", "password"); err != nil {
		t.Fatalf("authenticate sqlite seed admin: %v", err)
	}

	postList, err := repositories.PostRepo.List(context.Background(), posts.ListQuery{Page: 1, PageSize: 5})
	if err != nil {
		t.Fatalf("list sqlite posts: %v", err)
	}
	if postList.Total == 0 {
		t.Fatal("expected sqlite seed posts")
	}

	if _, err := repositories.CommentRepo.AdminList(context.Background(), comments.ListQuery{Page: 1, PageSize: 5}); err != nil {
		t.Fatalf("list sqlite comments: %v", err)
	}
	if _, err := repositories.MessageRepo.AdminList(context.Background(), messages.ListQuery{Page: 1, PageSize: 5}); err != nil {
		t.Fatalf("list sqlite messages: %v", err)
	}
	if _, err := repositories.SubmissionRepo.AdminList(context.Background(), submissions.ListQuery{Page: 1, PageSize: 5}); err != nil {
		t.Fatalf("list sqlite submissions: %v", err)
	}
	if _, err := repositories.UserRepo.List(context.Background(), users.ListQuery{Page: 1, PageSize: 5}); err != nil {
		t.Fatalf("list sqlite users: %v", err)
	}
	if _, err := repositories.AdminPostRepo.List(context.Background(), adminposts.ListQuery{Page: 1, PageSize: 5}); err != nil {
		t.Fatalf("list sqlite admin posts: %v", err)
	}
	if _, err := repositories.OperationsRepo.GetSettings(context.Background()); err != nil {
		t.Fatalf("load sqlite settings: %v", err)
	}
	if categories, err := repositories.TaxonomyRepo.ListCategories(context.Background()); err != nil {
		t.Fatalf("list sqlite categories: %v", err)
	} else if len(categories) == 0 {
		t.Fatal("expected sqlite seed categories")
	}
	if result, err := repositories.TopicRepo.List(context.Background(), topics.ListQuery{Page: 1, PageSize: 5}); err != nil {
		t.Fatalf("list sqlite topics: %v", err)
	} else if result.Total == 0 {
		t.Fatal("expected sqlite seed topics")
	}
}

func TestSetupRepositoriesDoesNotUseSQLiteWhenPostgresIsConfigured(t *testing.T) {
	sqlitePath := filepath.Join(t.TempDir(), "blog.sqlite")
	cfg := config.Config{
		AppEnv:      "development",
		DBType:      "postgres",
		DatabaseURL: "postgres://blog:blog@127.0.0.1:1/blog?sslmode=disable",
		SQLitePath:  sqlitePath,
	}

	db, _, err := setupRepositories(context.Background(), cfg)
	if err == nil {
		_ = db.Close()
		t.Fatal("expected postgres initialization error")
	}
	if _, statErr := os.Stat(sqlitePath); !os.IsNotExist(statErr) {
		t.Fatalf("expected no sqlite file when postgres is configured, stat error: %v", statErr)
	}
}
