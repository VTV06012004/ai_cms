package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultPort = "8080"

var defaultCORSOrigins = []string{"http://localhost:3000", "http://localhost:3001"}

var (
	ErrMissingJWTSecret = errors.New("JWT_SECRET is required")
	ErrMissingDSN       = errors.New("DB_DSN is required")
)

type InitialAdmin struct {
	Name     string
	Email    string
	Password string
}

// Config chứa toàn bộ cấu hình đọc từ biến môi trường.
type Config struct {
	Port               string
	DatabaseDSN        string
	JWTSecret          string
	CORSAllowedOrigins []string
	InitialAdmin       InitialAdmin
}

// Load đọc và kiểm tra cấu hình từ môi trường tiến trình.
func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Port:               strings.TrimSpace(getenv("PORT")),
		DatabaseDSN:        strings.TrimSpace(getenv("DB_DSN")),
		JWTSecret:          getenv("JWT_SECRET"),
		CORSAllowedOrigins: splitList(getenv("CORS_ALLOWED_ORIGINS")),
		InitialAdmin: InitialAdmin{
			Name:     getenv("INITIAL_ADMIN_NAME"),
			Email:    getenv("INITIAL_ADMIN_EMAIL"),
			Password: getenv("INITIAL_ADMIN_PASSWORD"),
		},
	}

	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return Config{}, ErrMissingJWTSecret
	}
	if cfg.DatabaseDSN == "" {
		return Config{}, ErrMissingDSN
	}
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	if port, err := strconv.Atoi(cfg.Port); err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535, got %q", cfg.Port)
	}
	if len(cfg.CORSAllowedOrigins) == 0 {
		cfg.CORSAllowedOrigins = append([]string(nil), defaultCORSOrigins...)
	}
	return cfg, nil
}

// splitList tách chuỗi phân cách bằng dấu phẩy, bỏ khoảng trắng và phần tử rỗng.
func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
