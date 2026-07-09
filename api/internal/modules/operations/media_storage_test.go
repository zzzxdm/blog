package operations

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalMediaStorageSaveAndDelete(t *testing.T) {
	root := t.TempDir()
	storage := NewLocalMediaStorage(root)

	publicURL, err := storage.Save(context.Background(), "2026/07/example.png", strings.NewReader("image-bytes"), 11, "image/png")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if publicURL != "/uploads/2026/07/example.png" {
		t.Fatalf("Save() publicURL = %q", publicURL)
	}

	target := filepath.Join(root, "2026", "07", "example.png")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("saved file read error = %v", err)
	}
	if string(data) != "image-bytes" {
		t.Fatalf("saved file = %q", data)
	}

	if err := storage.Delete(context.Background(), publicURL); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleted file still exists or unexpected stat error: %v", err)
	}
}

func TestLocalMediaStorageDeleteIgnoresEscapedPath(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "uploads")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	outside := filepath.Join(base, "outside.txt")
	if err := os.WriteFile(outside, []byte("keep"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	storage := NewLocalMediaStorage(root)
	if err := storage.Delete(context.Background(), "/uploads/../outside.txt"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	data, err := os.ReadFile(outside)
	if err != nil {
		t.Fatalf("outside file read error = %v", err)
	}
	if string(data) != "keep" {
		t.Fatalf("outside file changed to %q", data)
	}
}

func TestMinIOMediaStoragePublicObjectURL(t *testing.T) {
	storage := &MinIOMediaStorage{
		endpoint:  "localhost:9000",
		bucket:    "blog-media",
		publicURL: "https://cdn.example.com/media",
	}

	publicURL := storage.publicObjectURL("2026/07/example.png")
	if publicURL != "https://cdn.example.com/media/2026/07/example.png" {
		t.Fatalf("publicObjectURL() = %q", publicURL)
	}
	if objectName := storage.objectNameFromURL(publicURL); objectName != "2026/07/example.png" {
		t.Fatalf("objectNameFromURL() = %q", objectName)
	}
}
