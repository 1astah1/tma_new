package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Telegram  TelegramConfig
	App       AppConfig
	UploadDir string
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type TelegramConfig struct {
	BotToken    string
	WebhookURL  string
	EncryptKey  string
}

type AppConfig struct {
	Environment string
	LogLevel    string
	AdminURL    string
	TMAURL      string
}

func getExecDir() string {
	exe, _ := os.Executable()
	return filepath.Dir(exe)
}

func loadEnvFile() {
	candidates := []string{
		".env",
		filepath.Join(getExecDir(), ".env"),
	}
	for _, path := range candidates {
		if err := parseEnvFile(path); err == nil {
			return
		}
	}
}

func parseEnvFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key != "" {
			os.Setenv(key, val)
		}
	}
	return nil
}

func Load() *Config {
	loadEnvFile()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	encryptKey := os.Getenv("ACCOUNT_ENCRYPTION_KEY")
	if encryptKey == "" {
		panic("ACCOUNT_ENCRYPTION_KEY environment variable is required")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/tma_shop?sslmode=disable"),
		},
		JWT: JWTConfig{
			Secret:     jwtSecret,
			AccessTTL:  getDuration("JWT_ACCESS_TTL", 24*time.Hour),
			RefreshTTL: getDuration("JWT_REFRESH_TTL", 168*time.Hour),
		},
		Telegram: TelegramConfig{
			BotToken:   getEnv("BOT_TOKEN", ""),
			WebhookURL: getEnv("BOT_WEBHOOK_URL", ""),
			EncryptKey: encryptKey,
		},
		UploadDir: getEnv("UPLOAD_DIR", filepath.Join(getExecDir(), "uploads")),
		App: AppConfig{
			Environment: getEnv("ENVIRONMENT", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "debug"),
			AdminURL:    getEnv("ADMIN_PANEL_URL", "http://localhost:5173"),
			TMAURL:      getEnv("TMA_URL", "http://localhost:5173"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
