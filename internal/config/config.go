package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// Database
	DatabaseURL string
	DBMaxConns  int

	// Auth
	JWTSecret string

	// CORS
	AllowedOrigin string

	// Upload
	UploadDir    string
	UploadMaxMB  int64

	// Encryption
	DataEncryptionKey string
}

func Load() *Config {
	return &Config{
		// Server
		Port:         getEnvInt("PORT", 9090),
		ReadTimeout:  getEnvDuration("READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getEnvDuration("WRITE_TIMEOUT", 60*time.Second),
		IdleTimeout:  getEnvDuration("IDLE_TIMEOUT", 60*time.Second),

		// Database
		DatabaseURL: getEnv("DATABASE_URL",
			"postgres://postgres:postgres@localhost:5432/comic_harmony?sslmode=disable"),
		DBMaxConns: getEnvInt("DB_MAX_CONNS", 10),

		// Auth (no default — force explicit)
		JWTSecret: getEnv("JWT_SECRET",
			"comic-harmony-dev-secret-change-in-production"),

		// CORS
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "*"),

		// Upload
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		UploadMaxMB: getEnvInt64("UPLOAD_MAX_MB", 500),

		// Encryption
		DataEncryptionKey: getEnv("DATA_ENCRYPTION_KEY",
			"comic-harmony-encryption-key-change-in-prod"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
