package config

import "os"

type Config struct {
	AppEnv      string
	HTTPAddr    string
	WebOrigin   string
	DatabaseURL string
	RedisAddr   string
}

func Load() Config {
	return Config{
		AppEnv:      getenv("APP_ENV", "development"),
		HTTPAddr:    getenv("API_HTTP_ADDR", ":8080"),
		WebOrigin:   getenv("WEB_ORIGIN", "http://localhost:5173"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://blog:blog@localhost:5432/blog?sslmode=disable"),
		RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
