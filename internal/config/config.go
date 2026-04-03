package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultTimeoutMS = 30000

	defaultLogDir = "./logs"
)

// Config holds all framework configuration loaded from environment variables.
type Config struct {
	// BaseURL is the application base URL for navigation.
	// Env: BASE_URL (default: https://example.com).
	BaseURL string
	// Timeout is the default operation timeout in milliseconds.
	// Env: TIMEOUT_MS (default: 30000).
	Timeout time.Duration
	// AllureReportDir is the directory for Allure results.
	// Env: ALLURE_RESULTS_DIR (default: ./allure-results).
	AllureReportDir string
	// LogDir is the directory for log files.
	// Env: LOG_DIR (default: ./logs).
	LogDir string
}

// Load reads configuration from environment variables with sensible defaults.
// It loads .env file from path specified in ENV_FILE if set.
// Returns a Config with all settings loaded.
func Load() *Config {
	_ = godotenv.Load(os.Getenv("ENV_FILE"))

	return &Config{
		BaseURL:         getEnv("BASE_URL", "https://example.com"),
		Timeout:         getDuration("TIMEOUT_MS", defaultTimeoutMS),
		AllureReportDir: getEnv("ALLURE_RESULTS_DIR", "./allure-results"),
		LogDir:          getEnv("LOG_DIR", defaultLogDir),
	}
}

// getEnv returns the environment variable value or defaultVal if unset.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// getDuration parses a duration in milliseconds from environment variable.
// Returns defaultMs converted to [time.Duration] on parse error or if unset.
func getDuration(key string, defaultMs int) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(time.Duration(defaultMs).Milliseconds())
	}
	ms, err := strconv.Atoi(v)
	if err != nil {
		return time.Duration(time.Duration(defaultMs).Milliseconds())
	}
	return time.Duration(time.Duration(ms).Milliseconds())
}
