package config

import (
	"encoding/json"
	"os"
	"strconv"
)

type Database struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

type Config struct {
	Database Database `json:"database"`
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
