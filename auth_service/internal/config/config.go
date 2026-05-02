package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Database struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

type AccessToken struct {
	Secret   string        `json:"secret"`
	Issuer   string        `json:"issuer"`
	Audience string        `json:"audience"`
	TTL      time.Duration `json:"TTL"`
}

type RefreshToken struct {
	Secret string `json:"secret"`
}

type Config struct {
	Database     Database     `json:"database"`
	AccessToken  AccessToken  `json:"accessToken"`
	RefreshToken RefreshToken `json:"refreshToken"`
}

func Load() (*Config, error) {
	cfg := &Config{
		Database: Database{
			Host:    "localhost",
			Port:    5432,
			User:    "postgres",
			DBName:  "auth_service",
			SSLMode: "disable",
		},
	}

	if path := os.Getenv("CONFIG_FILE"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	if v := getEnv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := getEnvInt("DB_PORT"); v != 0 {
		cfg.Database.Port = v
	}
	if v := getEnv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := getEnv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := getEnv("DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}
	if v := getEnv("DB_SSLMODE"); v != "" {
		cfg.Database.SSLMode = v
	}
	if v := getEnv("JWT_SECRET"); v != "" {
		cfg.AccessToken.Secret = v
	}
	if v := getEnv("JWT_ISSUER"); v != "" {
		cfg.AccessToken.Issuer = v
	}
	if v := getEnv("JWT_AUDIENCE"); v != "" {
		cfg.AccessToken.Audience = v
	}
	if v := getEnv("JWT_TTL"); v != "" {
		ttl, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT_TTL %q: %w", v, err)
		}

		cfg.AccessToken.TTL = ttl
	}
	if v := getEnv("REFRESH_TOKEN_SECRET"); v != "" {
		cfg.RefreshToken.Secret = v
	}

	return cfg, nil
}

func getEnv(key string) string {
	if path := os.Getenv(key + "_FILE"); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			return string(data)
		}
	}
	return os.Getenv(key)
}

func getEnvInt(key string) int {
	if val := getEnv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return 0
}
