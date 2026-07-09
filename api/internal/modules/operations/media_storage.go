package operations

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MediaStorage interface {
	Save(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, publicURL string) error
}

type LocalMediaStorage struct {
	dir string
}

func NewLocalMediaStorage(dir string) *LocalMediaStorage {
	if dir == "" {
		dir = "uploads"
	}

	return &LocalMediaStorage{dir: dir}
}

func (storage *LocalMediaStorage) Save(_ context.Context, key string, reader io.Reader, _ int64, _ string) (string, error) {
	targetPath := filepath.Join(storage.dir, filepath.FromSlash(key))
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}

	destination, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(destination, reader); err != nil {
		_ = destination.Close()
		_ = os.Remove(targetPath)
		return "", err
	}
	if err := destination.Close(); err != nil {
		_ = os.Remove(targetPath)
		return "", err
	}

	return "/uploads/" + key, nil
}

func (storage *LocalMediaStorage) Delete(_ context.Context, publicURL string) error {
	if !strings.HasPrefix(publicURL, "/uploads/") {
		return nil
	}

	relativePath := strings.TrimPrefix(publicURL, "/uploads/")
	targetPath := filepath.Join(storage.dir, filepath.FromSlash(relativePath))
	root, err := filepath.Abs(storage.dir)
	if err != nil {
		return err
	}
	target, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	if target == root || !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return nil
	}
	if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

type MinIOMediaStorage struct {
	client    *minio.Client
	endpoint  string
	bucket    string
	useSSL    bool
	publicURL string
}

type MinIOStorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	PublicURL string
}

func NewMinIOMediaStorage(config MinIOStorageConfig) (*MinIOMediaStorage, error) {
	endpoint := strings.TrimSpace(config.Endpoint)
	accessKey := strings.TrimSpace(config.AccessKey)
	secretKey := strings.TrimSpace(config.SecretKey)
	bucket := strings.TrimSpace(config.Bucket)
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, errors.New("minio endpoint, access key, secret key and bucket are required")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOMediaStorage{
		client:    client,
		endpoint:  endpoint,
		bucket:    bucket,
		useSSL:    config.UseSSL,
		publicURL: strings.TrimRight(strings.TrimSpace(config.PublicURL), "/"),
	}, nil
}

func (storage *MinIOMediaStorage) Save(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	exists, err := storage.client.BucketExists(ctx, storage.bucket)
	if err != nil {
		return "", err
	}
	if !exists {
		if err := storage.client.MakeBucket(ctx, storage.bucket, minio.MakeBucketOptions{}); err != nil {
			return "", err
		}
	}

	objectName := path.Clean(strings.TrimPrefix(key, "/"))
	if objectName == "." || strings.HasPrefix(objectName, "../") {
		return "", errors.New("invalid object key")
	}

	_, err = storage.client.PutObject(ctx, storage.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return storage.publicObjectURL(objectName), nil
}

func (storage *MinIOMediaStorage) Delete(ctx context.Context, publicURL string) error {
	objectName := storage.objectNameFromURL(publicURL)
	if objectName == "" {
		return nil
	}

	return storage.client.RemoveObject(ctx, storage.bucket, objectName, minio.RemoveObjectOptions{})
}

func (storage *MinIOMediaStorage) publicObjectURL(objectName string) string {
	if storage.publicURL != "" {
		return storage.publicURL + "/" + objectName
	}

	scheme := "http"
	if storage.useSSL {
		scheme = "https"
	}

	return scheme + "://" + strings.TrimRight(storage.endpoint, "/") + "/" + storage.bucket + "/" + objectName
}

func (storage *MinIOMediaStorage) objectNameFromURL(publicURL string) string {
	if storage.publicURL != "" && strings.HasPrefix(publicURL, storage.publicURL+"/") {
		return strings.TrimPrefix(publicURL, storage.publicURL+"/")
	}

	parsed, err := url.Parse(publicURL)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.TrimPrefix(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != storage.bucket {
		return ""
	}

	return strings.Join(parts[1:], "/")
}
