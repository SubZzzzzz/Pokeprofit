package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Discord  DiscordConfig
	Scraper  ScraperConfig
	Log      LogConfig
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	URL string
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	URL string
}

// DiscordConfig holds Discord bot settings.
type DiscordConfig struct {
	Token   string
	GuildID string
}

// ScraperConfig holds scraper settings.
type ScraperConfig struct {
	RateLimitDelay time.Duration
	MaxRetries     int
	UserAgents     []string
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{}

	// Database
	cfg.Database.URL = getEnv("DATABASE_URL", "")
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// Redis
	cfg.Redis.URL = getEnv("REDIS_URL", "redis://localhost:6379/0")

	// Discord
	cfg.Discord.Token = getEnv("DISCORD_TOKEN", "")
	if cfg.Discord.Token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is required")
	}
	cfg.Discord.GuildID = getEnv("DISCORD_GUILD_ID", "")

	// Scraper
	rateLimitMs := getEnvInt("SCRAPER_RATE_LIMIT_MS", 2000)
	cfg.Scraper.RateLimitDelay = time.Duration(rateLimitMs) * time.Millisecond
	cfg.Scraper.MaxRetries = getEnvInt("SCRAPER_MAX_RETRIES", 3)
	cfg.Scraper.UserAgents = getEnvSlice("SCRAPER_USER_AGENTS", []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	})

	// Logging
	cfg.Log.Level = getEnv("LOG_LEVEL", "info")

	return cfg, nil
}

// LoadWithDefaults loads configuration with defaults for testing.
func LoadWithDefaults() *Config {
	return &Config{
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://localhost:5432/pokeprofit_test?sslmode=disable"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		Discord: DiscordConfig{
			Token:   getEnv("DISCORD_TOKEN", "test_token"),
			GuildID: getEnv("DISCORD_GUILD_ID", ""),
		},
		Scraper: ScraperConfig{
			RateLimitDelay: time.Duration(getEnvInt("SCRAPER_RATE_LIMIT_MS", 2000)) * time.Millisecond,
			MaxRetries:     getEnvInt("SCRAPER_MAX_RETRIES", 3),
			UserAgents: []string{
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
