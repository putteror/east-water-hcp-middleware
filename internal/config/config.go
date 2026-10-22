package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration grouped by domain
type Config struct {
	App        AppConfig
	ThirdParty ThirdPartyConfig
	Auth       AuthConfig
}

// AppConfig holds HTTP server configuration
type AppConfig struct {
	Name string
	Env  string
	Port string
}

// ThirdPartyConfig holds credentials and endpoints for external 3rd party APIs
type ThirdPartyConfig struct {
	BaseURL string
	APIKey  string
	Secret  string
	Timeout time.Duration
	Debug   bool
	UseMock bool
}

// AuthConfig holds API key or authentication configuration for clients calling this backend
type AuthConfig struct {
	APIKey string
}

// LoadConfig loads environment variables and returns Config struct.
// ALL os.Getenv calls are strictly isolated in this file.
func LoadConfig() *Config {
	// Load .env if present (ignore error in production or docker containers)
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] Notice: .env file not found, using system environment variables")
	}

	appPort := getEnvWithDefault("PORT", "8080")
	appEnv := getEnvWithDefault("APP_ENV", "development")
	appName := getEnvWithDefault("APP_NAME", "eastwater-backend")

	// Third Party API configuration
	thirdPartyBaseURL := getEnvWithDefault("THIRD_PARTY_BASE_URL", "https://api.thirdparty.example.com")
	thirdPartyAPIKey := os.Getenv("THIRD_PARTY_API_KEY")
	thirdPartySecret := os.Getenv("THIRD_PARTY_SECRET")

	timeoutSecStr := getEnvWithDefault("THIRD_PARTY_TIMEOUT_SECONDS", "30")
	timeoutSec, err := strconv.Atoi(timeoutSecStr)
	if err != nil {
		timeoutSec = 30
	}

	debugStr := getEnvWithDefault("THIRD_PARTY_DEBUG", "false")
	debug := debugStr == "true" || debugStr == "1"

	useMockStr := getEnvWithDefault("USE_MOCK_3RDPARTY", "false")
	useMock := useMockStr == "true" || useMockStr == "1" || thirdPartyBaseURL == "mock"

	// Auth configuration
	authAPIKey := os.Getenv("API_KEY")

	return &Config{
		App: AppConfig{
			Name: appName,
			Env:  appEnv,
			Port: appPort,
		},
		ThirdParty: ThirdPartyConfig{
			BaseURL: thirdPartyBaseURL,
			APIKey:  thirdPartyAPIKey,
			Secret:  thirdPartySecret,
			Timeout: time.Duration(timeoutSec) * time.Second,
			Debug:   debug,
			UseMock: useMock,
		},
		Auth: AuthConfig{
			APIKey: authAPIKey,
		},
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
