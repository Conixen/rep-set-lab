package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port                 string
	DatabaseURL          string
	JWTSecret            string
	AnthropicKey         string
	OpenAIKey            string
	GeminiKey            string
	ExerciseDBKey        string
	AllowedOrigins       []string
	BootstrapAdminEmail  string
	BootstrapAdminPass   string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		AnthropicKey:        os.Getenv("ANTHROPIC_API_KEY"),
		OpenAIKey:           os.Getenv("OPENAI_API_KEY"),
		GeminiKey:           os.Getenv("GEMINI_API_KEY"),
		ExerciseDBKey:       os.Getenv("EXERCISEDB_API_KEY"),
		AllowedOrigins:      parseList(getEnv("ALLOWED_ORIGINS", "http://localhost:5173")),
		BootstrapAdminEmail: os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
		BootstrapAdminPass:  os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
