package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	JWTExpiry      time.Duration
	AllowedOrigins []string
	StoragePath    string
	MaxUploadSize  int64

	RateLimitRequests int
	RateLimitWindow   time.Duration

	SMTPHost string
	SMTPPort int
	SMTPFrom string

	Env string
}


func Load() *Config {
	jwtExpiry, _ := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	rateLimitReqs, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "100"))
	rateLimitWindow, _ := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	maxUpload, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64) 

	origins := strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ",")

	return &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/todo?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production-use-256-bit-key"),
		JWTExpiry:         jwtExpiry,
		AllowedOrigins:    origins,
		StoragePath:       getEnv("STORAGE_PATH", "./uploads"),
		MaxUploadSize:     maxUpload,
		RateLimitRequests: rateLimitReqs,
		RateLimitWindow:   rateLimitWindow,
		SMTPHost:          getEnv("SMTP_HOST", "localhost"),
		SMTPPort:          smtpPort,
		SMTPFrom:          getEnv("SMTP_FROM", "noreply@gotodo.com"),
		Env:               getEnv("ENV", "development"),
	}
}

func getEnv(key, fallback string) string{
	if v := os.Getenv(key); v != ""{
		return v
	}

	return fallback
}