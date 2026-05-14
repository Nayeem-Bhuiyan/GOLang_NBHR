package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	RateLimit RateLimitConfig
}

type AppConfig struct {
	Name           string
	Env            string
	Port           string
	Version        string
	Debug          bool
	RequestTimeout time.Duration
}

type JWTConfig struct {
	AccessSecret   string
	RefreshSecret  string
	AccessExpiry   time.Duration
	RefreshExpiry  time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

type RateLimitConfig struct {
	Requests int64
	Period   time.Duration
}

// Load reads environment variables and returns a validated Config.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Name:           getEnv("APP_NAME", "nbhr"),
			Env:            getEnv("APP_ENV", "development"),
			Port:           getEnv("APP_PORT", "8080"),
			Version:        getEnv("APP_VERSION", "v1"),
			Debug:          getEnvBool("APP_DEBUG", false),
			RequestTimeout: time.Duration(getEnvInt("REQUEST_TIMEOUT", 30)) * time.Second,
		},
		Database: DatabaseConfig{
			Host:            getEnvRequired("DB_HOST"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnvRequired("DB_USER"),
			Password:        getEnvRequired("DB_PASSWORD"),
			Name:            getEnvRequired("DB_NAME"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME", 300)) * time.Second,
		},
		JWT: JWTConfig{
			AccessSecret:  getEnvRequired("JWT_ACCESS_SECRET"),
			RefreshSecret: getEnvRequired("JWT_REFRESH_SECRET"),
			AccessExpiry:  time.Duration(getEnvInt("JWT_ACCESS_EXPIRY", 15)) * time.Minute,
			RefreshExpiry: time.Duration(getEnvInt("JWT_REFRESH_EXPIRY", 10080)) * time.Minute,
		},
		CORS: CORSConfig{
			AllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		},
		RateLimit: RateLimitConfig{
			Requests: int64(getEnvInt("RATE_LIMIT_REQUESTS", 100)),
			Period:   time.Duration(getEnvInt("RATE_LIMIT_PERIOD", 60)) * time.Second,
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if len(c.JWT.AccessSecret) < 32 {
		return fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 characters")
	}
	if len(c.JWT.RefreshSecret) < 32 {
		return fmt.Errorf("JWT_REFRESH_SECRET must be at least 32 characters")
	}
	return nil
}

func (c *AppConfig) IsProduction() bool {
	return c.Env == "production"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvRequired(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}