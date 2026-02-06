package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	ServerPort string
	
	// LLM
	AnthropicAPIKey string
	OpenAIAPIKey    string
	LLMProvider     string // "anthropic" or "openai"
	LLMModel        string
	
	// Vector DB
	QdrantURL        string
	QdrantAPIKey     string
	QdrantCollection string
	
	// GitHub
	GitHubToken string
	
	// Working Directory
	WorkDir string
	
	// Agent
	MaxIterations int
	MaxTokens     int
	DryRun        bool
	
	// Logging
	LogLevel string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if present (ignored in production)
	_ = godotenv.Load()
	
	cfg := &Config{
		// Server defaults
		ServerPort: getEnv("SERVER_PORT", "8080"),
		
		// LLM defaults
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		LLMProvider:     getEnv("LLM_PROVIDER", "anthropic"),
		LLMModel:        getEnv("LLM_MODEL", "claude-3-5-sonnet-20241022"),
		
		// Qdrant defaults
		QdrantURL:        getEnv("QDRANT_URL", "localhost:6334"),
		QdrantAPIKey:     os.Getenv("QDRANT_API_KEY"),
		QdrantCollection: getEnv("QDRANT_COLLECTION", "sentinel_fixes"),
		
		// GitHub
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		
		// Working Directory
		WorkDir: getEnv("WORK_DIR", "/tmp/sentinel"),
		
		// Agent
		MaxIterations: getEnvInt("MAX_ITERATIONS", 10),
		MaxTokens:     getEnvInt("MAX_TOKENS", 4000),
		DryRun:        getEnvBool("DRY_RUN", "true"),
		
		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
	
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key, defaultValue string) bool {
	value := getEnv(key, defaultValue)
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return boolValue
}
