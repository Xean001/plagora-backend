package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DB          DBConfig
	JWT         JWTConfig
	Server      ServerConfig
	Admin       AdminConfig
	DatabaseURL string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret                 string
	ExpirationHours        int
	RefreshExpirationHours int
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type AdminConfig struct {
	Email    string
	Password string
}

// 🔥 Nuevo DSN inteligente
func (c *Config) DSN() string {
	// Si existe DATABASE_URL la usamos directamente
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}

	// Si no, construimos manual (modo local)
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host,
		c.DB.Port,
		c.DB.User,
		c.DB.Password,
		c.DB.Name,
		c.DB.SSLMode,
	)
}

func Load() (*Config, error) {
	// Solo carga .env si existe (local)
	_ = godotenv.Load()

	jwtExp, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "1"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION_HOURS: %w", err)
	}

	jwtRefExp, err := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRATION_HOURS", "168"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRATION_HOURS: %w", err)
	}

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),

		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "plagora"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		JWT: JWTConfig{
			Secret:                 mustEnv("JWT_SECRET"),
			ExpirationHours:        jwtExp,
			RefreshExpirationHours: jwtRefExp,
		},

		Server: ServerConfig{
			Port:    getEnv("PORT", getEnv("SERVER_PORT", "8080")), // 🔥 Render compatible
			GinMode: getEnv("GIN_MODE", "debug"),
		},

		Admin: AdminConfig{
			Email:    getEnv("ADMIN_EMAIL", "admin@plagora.com"),
			Password: getEnv("ADMIN_PASSWORD", "changeme"),
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}
