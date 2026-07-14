package adminposts

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/posts"

	_ "modernc.org/sqlite"
)

type fakePublisher struct{}

func (fakePublisher) PublishAdmin(ctx context.Context, input posts.PublishInput, existingSlug string) (posts.Post, error) {
	slug := input.Slug
	if existingSlug != "" {
		slug = existingSlug
	}
	return posts.Post{
		Slug:       slug,
		Title:      input.Title,
		Summary:    input.Summary,
		Content:    input.Content,
		Visibility: input.Visibility,
		Category:   input.Category,
		Tags:       input.Tags,
		CoverImage: input.CoverImage,
		AuthorID:   input.AuthorID,
		AuthorName: input.AuthorName,
	}, nil
}

func openAdminPostsTestRepo(t *testing.T) *SQLRepository {
	t.Helper()

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "adminposts.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(`
		CREATE TABLE admin_posts (
			id text PRIMARY KEY,
			slug text NOT NULL,
			title text NOT NULL,
			status text NOT NULL,
			updated_at text NOT NULL,
			data text NOT NULL
		);
	`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	return &SQLRepository{
		db:     db,
		now:    time.Now,
		sqlite: true,
	}
}

func TestPublishDoesNotDuplicateRevisionWhenContentUnchanged(t *testing.T) {
	repo := openAdminPostsTestRepo(t)
	actor := auth.User{ID: "u1", DisplayName: "Admin"}

	saved, err := repo.Save(context.Background(), "", SaveRequest{
		Title:   "Hello",
		Content: "Body",
		Status:  StatusDraft,
	}, actor)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if saved.Version != 1 {
		t.Fatalf("version after save = %d, want 1", saved.Version)
	}
	if len(saved.Revisions) != 1 {
		t.Fatalf("revisions after save = %d, want 1", len(saved.Revisions))
	}

	published, err := repo.Publish(context.Background(), saved.ID, fakePublisher{}, actor)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.Version != 1 {
		t.Fatalf("version after publish = %d, want 1", published.Version)
	}
	if len(published.Revisions) != 1 {
		t.Fatalf("revisions after publish = %d, want 1", len(published.Revisions))
	}
	if published.Status != StatusPublished {
		t.Fatalf("status = %s, want published", published.Status)
	}
}

func TestDeleteRevisionRemovesHistoryButKeepsCurrent(t *testing.T) {
	repo := openAdminPostsTestRepo(t)
	actor := auth.User{ID: "u1", DisplayName: "Admin"}

	first, err := repo.Save(context.Background(), "", SaveRequest{
		Title:   "V1",
		Content: "C1",
		Status:  StatusDraft,
	}, actor)
	if err != nil {
		t.Fatalf("save1: %v", err)
	}
	second, err := repo.Save(context.Background(), first.ID, SaveRequest{
		Title:   "V2",
		Content: "C2",
		Status:  StatusDraft,
	}, actor)
	if err != nil {
		t.Fatalf("save2: %v", err)
	}
	if second.Version != 2 || len(second.Revisions) != 2 {
		t.Fatalf("after second save version=%d revisions=%d", second.Version, len(second.Revisions))
	}

	oldID := ""
	for _, rev := range second.Revisions {
		if rev.Version == 1 {
			oldID = rev.ID
		}
	}
	if oldID == "" {
		t.Fatal("missing revision v1")
	}

	updated, err := repo.DeleteRevision(context.Background(), second.ID, oldID)
	if err != nil {
		t.Fatalf("delete revision: %v", err)
	}
	if len(updated.Revisions) != 1 {
		t.Fatalf("revisions after delete = %d, want 1", len(updated.Revisions))
	}
	if updated.Version != 2 {
		t.Fatalf("current version changed to %d", updated.Version)
	}

	_, err = repo.DeleteRevision(context.Background(), second.ID, updated.Revisions[0].ID)
	if err != ErrInvalidPost {
		t.Fatalf("delete current error = %v, want ErrInvalidPost", err)
	}
}

func TestSaveSkipsRevisionWhenUnchanged(t *testing.T) {
	repo := openAdminPostsTestRepo(t)
	actor := auth.User{ID: "u1", DisplayName: "Admin"}

	saved, err := repo.Save(context.Background(), "", SaveRequest{
		Title:   "Same",
		Content: "Body",
		Status:  StatusDraft,
	}, actor)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	again, err := repo.Save(context.Background(), saved.ID, SaveRequest{
		Title:   "Same",
		Content: "Body",
		Status:  StatusDraft,
	}, actor)
	if err != nil {
		t.Fatalf("save again: %v", err)
	}
	if again.Version != saved.Version {
		t.Fatalf("version bumped from %d to %d on identical save", saved.Version, again.Version)
	}
	if len(again.Revisions) != len(saved.Revisions) {
		t.Fatalf("revisions changed on identical save: %d -> %d", len(saved.Revisions), len(again.Revisions))
	}
}
