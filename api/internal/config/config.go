package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                 string
	HTTPAddr               string
	WebOrigin              string
	PublicURL              string
	DBType                 string
	DatabaseURL            string
	SQLitePath             string
	RedisAddr              string
	RedisPassword          string
	UploadDir              string
	MediaStorage           string
	MinIOEndpoint          string
	MinIOAccessKey         string
	MinIOSecretKey         string
	MinIOBucket            string
	MinIOUseSSL            bool
	MinIOPublicURL         string
	SMTPHost               string
	SMTPPort               string
	SMTPUsername           string
	SMTPPassword           string
	SMTPFrom               string
	BootstrapAdminEmail    string
	BootstrapAdminPassword string
	BootstrapAdminName     string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppEnv:                 getenv("APP_ENV", "development"),
		HTTPAddr:               getenv("API_HTTP_ADDR", ":8080"),
		WebOrigin:              getenv("WEB_ORIGIN", "http://localhost:5173"),
		PublicURL:              getenv("PUBLIC_URL", getenv("WEB_ORIGIN", "http://localhost:5173")),
		DBType:                 getenv("DB_TYPE", "sqlite"),
		DatabaseURL:            getenv("DATABASE_URL", "postgres://blog:blog@localhost:5432/blog?sslmode=disable"),
		SQLitePath:             getenv("SQLITE_PATH", "data/blog.sqlite"),
		RedisAddr:              getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:          getenv("REDIS_PASSWORD", ""),
		UploadDir:              getenv("UPLOAD_DIR", "uploads"),
		MediaStorage:           getenv("MEDIA_STORAGE", "local"),
		MinIOEndpoint:          getenv("MINIO_ENDPOINT", ""),
		MinIOAccessKey:         getenv("MINIO_ACCESS_KEY", ""),
		MinIOSecretKey:         getenv("MINIO_SECRET_KEY", ""),
		MinIOBucket:            getenv("MINIO_BUCKET", "blog-media"),
		MinIOUseSSL:            getenv("MINIO_USE_SSL", "false") == "true",
		MinIOPublicURL:         getenv("MINIO_PUBLIC_URL", ""),
		SMTPHost:               getenv("SMTP_HOST", ""),
		SMTPPort:               getenv("SMTP_PORT", "587"),
		SMTPUsername:           getenv("SMTP_USERNAME", ""),
		SMTPPassword:           getenv("SMTP_PASSWORD", ""),
		SMTPFrom:               getenv("SMTP_FROM", ""),
		BootstrapAdminEmail:    getenv("BOOTSTRAP_ADMIN_EMAIL", ""),
		BootstrapAdminPassword: getenv("BOOTSTRAP_ADMIN_PASSWORD", ""),
		BootstrapAdminName:     getenv("BOOTSTRAP_ADMIN_NAME", "管理员"),
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
