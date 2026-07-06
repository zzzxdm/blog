package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv       string
	HTTPAddr     string
	WebOrigin    string
	PublicURL    string
	DatabaseURL  string
	RedisAddr    string
	UploadDir    string
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppEnv:       getenv("APP_ENV", "development"),
		HTTPAddr:     getenv("API_HTTP_ADDR", ":8080"),
		WebOrigin:    getenv("WEB_ORIGIN", "http://localhost:5173"),
		PublicURL:    getenv("PUBLIC_URL", getenv("WEB_ORIGIN", "http://localhost:5173")),
		DatabaseURL:  getenv("DATABASE_URL", "postgres://blog:blog@localhost:5432/blog?sslmode=disable"),
		RedisAddr:    getenv("REDIS_ADDR", "localhost:6379"),
		UploadDir:    getenv("UPLOAD_DIR", "uploads"),
		SMTPHost:     getenv("SMTP_HOST", ""),
		SMTPPort:     getenv("SMTP_PORT", "587"),
		SMTPUsername: getenv("SMTP_USERNAME", ""),
		SMTPPassword: getenv("SMTP_PASSWORD", ""),
		SMTPFrom:     getenv("SMTP_FROM", ""),
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
