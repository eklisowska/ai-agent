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
	QdrantURL      string
	OllamaURL      string
	LLMModel       string
	EmbedModel     string
	CollectionName string
	TopK           int
	HTTPTimeout    time.Duration
	ListenAddr     string
}

func Load() (Config, error) {
	if err := loadDotenv(); err != nil {
		return Config{}, err
	}

	cfg := Config{
		QdrantURL:      envOrDefault("QDRANT_URL", "http://localhost:6333"),
		OllamaURL:      envOrDefault("OLLAMA_URL", "http://localhost:11434"),
		LLMModel:       envOrDefault("LLM_MODEL", "llama3"),
		EmbedModel:     envOrDefault("EMBED_MODEL", "nomic-embed-text"),
		CollectionName: envOrDefault("QDRANT_COLLECTION", "stock_facts"),
		TopK:           envOrDefaultInt("TOP_K", 8),
		HTTPTimeout:    envOrDefaultDuration("HTTP_TIMEOUT", 30*time.Second),
		ListenAddr:     envOrDefault("LISTEN_ADDR", ":8080"),
	}

	if cfg.TopK <= 0 {
		return Config{}, fmt.Errorf("TOP_K must be > 0")
	}

	return cfg, nil
}

// loadDotenv loads environment variables from a file before os.Getenv is used.
// If ENV_FILE is set, that path is required and must exist.
// Otherwise, if a file named ".env" exists in the working directory, it is loaded.
func loadDotenv() error {
	if p := strings.TrimSpace(os.Getenv("ENV_FILE")); p != "" {
		return godotenv.Load(p)
	}
	if _, err := os.Stat(".env"); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return godotenv.Load(".env")
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envOrDefaultInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envOrDefaultDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
