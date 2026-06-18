package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	MinIO    MinIOConfig
	JWT      JWTConfig
	Asynq    AsynqConfig
	Seed     SeedConfig
}

type AppConfig struct {
	Env      string
	Port     string
	Name     string
	LogLevel string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	URL      string
	MaxConns int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	URL      string
}

type MinIOConfig struct {
	RootUser     string
	RootPassword string
	Endpoint     string
	UseSSL       bool
	Bucket       string
	AccessKey    string
	SecretKey    string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type AsynqConfig struct {
	Concurrency int
}

type SeedConfig struct {
	OwnerEmail    string
	OwnerPassword string
}

func Load() (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			return nil, fmt.Errorf("load .env: %w", err)
		}
	}

	cfg := &Config{
		App: AppConfig{
			Env:      env("APP_ENV", "development"),
			Port:     env("APP_PORT", "8080"),
			Name:     env("APP_NAME", "e-fiber-admin"),
			LogLevel: env("LOG_LEVEL", "debug"),
		},
		Postgres: PostgresConfig{
			Host:     env("POSTGRES_HOST", "localhost"),
			Port:     env("POSTGRES_PORT", "5432"),
			User:     env("POSTGRES_USER", "efa"),
			Password: env("POSTGRES_PASSWORD", "efa_secret"),
			Database: env("POSTGRES_DB", "efa"),
			URL:      env("DATABASE_URL", "postgres://efa:efa_secret@localhost:5432/efa?sslmode=disable"),
			MaxConns: envInt("DB_MAX_CONNS", 25),
		},
		Redis: RedisConfig{
			Host:     env("REDIS_HOST", "localhost"),
			Port:     env("REDIS_PORT", "6379"),
			Password: env("REDIS_PASSWORD", ""),
			DB:       envInt("REDIS_DB", 0),
			URL:      env("REDIS_URL", "redis://localhost:6379/0"),
		},
		MinIO: MinIOConfig{
			RootUser:     env("MINIO_ROOT_USER", "efa"),
			RootPassword: env("MINIO_ROOT_PASSWORD", "efa_secret"),
			Endpoint:     env("MINIO_ENDPOINT", "localhost:9000"),
			UseSSL:       envBool("MINIO_USE_SSL", false),
			Bucket:       env("MINIO_BUCKET", "efa-media"),
			AccessKey:    env("S3_ACCESS_KEY", "efa"),
			SecretKey:    env("S3_SECRET_KEY", "efa_secret"),
		},
		JWT: JWTConfig{
			AccessSecret:  env("JWT_ACCESS_SECRET", ""),
			RefreshSecret: env("JWT_REFRESH_SECRET", ""),
			AccessTTL:     envDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    envDuration("JWT_REFRESH_TTL", 720*time.Hour),
		},
		Asynq: AsynqConfig{
			Concurrency: envInt("ASYNQ_CONCURRENCY", 10),
		},
		Seed: SeedConfig{
			OwnerEmail:    env("INIT_OWNER_EMAIL", "admin@e-fiber.local"),
			OwnerPassword: env("INIT_OWNER_PASSWORD", "change-me-now"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	var missing []string
	if c.JWT.AccessSecret == "" {
		missing = append(missing, "JWT_ACCESS_SECRET")
	}
	if c.JWT.RefreshSecret == "" {
		missing = append(missing, "JWT_REFRESH_SECRET")
	}
	if c.Postgres.URL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.Redis.URL == "" {
		missing = append(missing, "REDIS_URL")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c *Config) IsDev() bool { return c.App.Env == "development" }

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
