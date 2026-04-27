package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIKey     string
	Port          string
	IsDeveloper   bool
	PromptID      string
	PromptVersion string
}

func Load() *Config {
	godotenv.Load("../../.env") // грузим .env файл
	prompt_ver := os.Getenv("PROMPT_VERSION")
	return &Config{
		OpenAIKey:     os.Getenv("OpenAIKey"),
		Port:          getEnv("PORT", "8080"),
		PromptID:      os.Getenv("PROMPT_ID"),
		PromptVersion: prompt_ver,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
