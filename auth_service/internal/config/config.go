package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Database struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

type RefreshToken struct {
	Secret string `json:"secret"`
}

type Cookie struct {
	Secret string `json:"secret"`
}

type Resend struct {
	APIKey    string `json:"apiKey"`
	FromEmail string `json:"fromEmail"`
}

// OIDCProviderConfig describes a single third-party auth provider.
// Despite the "OIDC" name, this is also used for plain OAuth2 providers
// (e.g. GitHub) — IssuerURL is just left empty for those.
type OIDCProviderConfig struct {
	ProviderType string   `json:"providerType"`
	IssuerURL    string   `json:"issuerURL"`
	ClientID     string   `json:"clientID"`
	ClientSecret string   `json:"clientSecret"`
	RedirectURL  string   `json:"redirectURL"`
	Scopes       []string `json:"scopes"`
	Enabled      bool     `json:"enabled"`
}

type Config struct {
	Database      Database                      `json:"database"`
	RefreshToken  RefreshToken                  `json:"refreshToken"`
	Cookie        Cookie                        `json:"cookie"`
	Resend        Resend                        `json:"resend"`
	OIDCProviders map[string]OIDCProviderConfig `json:"oidcProviders"`
	FrontendURL   string                        `json:"frontendURL"`
}

const oidcEnvPrefix = "OIDC_"

// oidcFieldSuffixes lists the env-var suffixes that identify a provider field.
// To add a new field: add it to OIDCProviderConfig, this list, and the loader below.
var oidcFieldSuffixes = []string{
	"_PROVIDER_TYPE",
	"_ISSUER_URL",
	"_CLIENT_ID",
	"_CLIENT_SECRET",
	"_REDIRECT_URL",
	"_SCOPES",
	"_ENABLED",
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
		OIDCProviders: map[string]OIDCProviderConfig{},
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
	if v := getEnv("REFRESH_TOKEN_SECRET"); v != "" {
		cfg.RefreshToken.Secret = v
	}
	if v := getEnv("COOKIE_SECRET"); v != "" {
		cfg.Cookie.Secret = v
	}
	if v := getEnv("RESEND_API_KEY"); v != "" {
		cfg.Resend.APIKey = v
	}
	if v := getEnv("RESEND_FROM_EMAIL"); v != "" {
		cfg.Resend.FromEmail = v
	}
	if v := getEnv("FRONTEND_URL"); v != "" {
		cfg.FrontendURL = v
	}

	if err := loadOIDCProvidersFromEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadOIDCProvidersFromEnv discovers OIDC_<NAME>_* env vars and merges them
// into cfg.OIDCProviders. Values from env override values from CONFIG_FILE.
// Keys are normalized to lowercase (e.g. "google", "github").
func loadOIDCProvidersFromEnv(cfg *Config) error {
	if cfg.OIDCProviders == nil {
		cfg.OIDCProviders = map[string]OIDCProviderConfig{}
	}

	names := discoverOIDCProviderNames()

	for name := range names {
		prefix := oidcEnvPrefix + strings.ToUpper(name) + "_"
		p := cfg.OIDCProviders[name] // start from file-loaded value (if any)

		if v := getEnv(prefix + "PROVIDER_TYPE"); v != "" {
			p.ProviderType = v
		}
		if v := getEnv(prefix + "ISSUER_URL"); v != "" {
			p.IssuerURL = v
		}
		if v := getEnv(prefix + "CLIENT_ID"); v != "" {
			p.ClientID = v
		}
		if v := getEnv(prefix + "CLIENT_SECRET"); v != "" {
			p.ClientSecret = v
		}
		if v := getEnv(prefix + "REDIRECT_URL"); v != "" {
			p.RedirectURL = v
		}
		if v := getEnv(prefix + "SCOPES"); v != "" {
			p.Scopes = strings.Split(v, ",")
		}
		if v := getEnv(prefix + "ENABLED"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("invalid %sENABLED %q: %w", prefix, v, err)
			}
			p.Enabled = b
		} else if _, existed := cfg.OIDCProviders[name]; !existed {
			// New provider discovered purely from env: enable by default
			// if it has a client id. Existing file-loaded providers keep
			// their JSON-defined Enabled value.
			p.Enabled = p.ClientID != ""
		}

		if p.ProviderType == "" {
			p.ProviderType = name
		}

		cfg.OIDCProviders[name] = p
	}

	return nil
}

// discoverOIDCProviderNames walks the environment and returns the set of
// provider names that have at least one OIDC_<NAME>_<known suffix> var defined.
// Also handles the *_FILE convention used by getEnv.
func discoverOIDCProviderNames() map[string]struct{} {
	names := map[string]struct{}{}
	for _, kv := range os.Environ() {
		eq := strings.IndexByte(kv, '=')
		if eq < 0 {
			continue
		}
		key := kv[:eq]
		if !strings.HasPrefix(key, oidcEnvPrefix) {
			continue
		}
		bareKey := strings.TrimSuffix(key, "_FILE")
		for _, sfx := range oidcFieldSuffixes {
			if strings.HasSuffix(bareKey, sfx) {
				name := strings.TrimSuffix(strings.TrimPrefix(bareKey, oidcEnvPrefix), sfx)
				if name != "" {
					names[strings.ToLower(name)] = struct{}{}
				}
				break
			}
		}
	}
	return names
}

func getEnv(key string) string {
	if path := os.Getenv(key + "_FILE"); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			return strings.TrimRight(string(data), "\r\n")
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
