package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	JWTSecret      string
	JWTExpiresIn   time.Duration
	Port           string
	AllowedOrigins []string
	AdminUsername  string
	AdminPassword  string
	AdminFullName  string
}

func Load() (Config, error) {
	_ = godotenv.Load()
	expires, err := parseDuration(getenv("JWT_EXPIRES_IN", "24h"))
	if err != nil {
		return Config{}, fmt.Errorf("JWT_EXPIRES_IN: %w", err)
	}
	cfg := Config{
		DBHost: getenv("DB_HOST", "localhost"), DBPort: getenv("DB_PORT", "5432"),
		DBUser: getenv("DB_USER", "postgres"), DBPassword: os.Getenv("DB_PASSWORD"),
		DBName: getenv("DB_NAME", "intan_florist"), DBSSLMode: getenv("DB_SSLMODE", "disable"),
		JWTSecret: os.Getenv("JWT_SECRET"), JWTExpiresIn: expires, Port: getenv("PORT", "8080"),
		AllowedOrigins: splitOrigins(getenv("ALLOWED_ORIGINS", "http://localhost:3000")),
		AdminUsername:  getenv("ADMIN_USERNAME", "admin"), AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		AdminFullName: getenv("ADMIN_FULL_NAME", "Administrator"),
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	if cfg.AdminPassword == "" {
		return Config{}, errors.New("ADMIN_PASSWORD is required")
	}
	return cfg, nil
}

func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func parseDuration(value string) (time.Duration, error) {
	if duration, err := time.ParseDuration(value); err == nil {
		return duration, nil
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("must be a Go duration such as 24h or seconds")
	}
	return time.Duration(seconds) * time.Second, nil
}
