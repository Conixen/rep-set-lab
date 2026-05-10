package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	AnthropicKey string
	OpenAIKey    string
	GeminiKey    string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		AnthropicKey: os.Getenv("ANTHROPIC_API_KEY"),
		OpenAIKey:    os.Getenv("OPENAI_API_KEY"),
		GeminiKey:    os.Getenv("GEMINI_API_KEY"),
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
