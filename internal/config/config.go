package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// BrowserType defines the supported browser types for the test framework.
type BrowserType string

// Supported browser types.
const (
	BrowserChromium BrowserType = "chromium" // Google Chrome/Chromium browser
	BrowserFirefox  BrowserType = "firefox"  // Mozilla Firefox browser
	BrowserWebKit   BrowserType = "webkit"   // WebKit-based browsers (Safari)
)

// Default configuration values and output directories.
const (
	// Default timeout for operations in milliseconds.
	defaultTimeoutMS = 30000
	// Default slow motion delay between operations in milliseconds.
	defaultSlowMoMS = 0
	// Default viewport width in pixels.
	defaultViewportWidth = 1920
	// Default viewport height in pixels.
	defaultViewportHeight = 1080
	// Directory for log files.
	DefaultLogDir = "./artifacts/logs"
	// Directory for Playwright trace files.
	DefaultTracesDir = "./artifacts/traces"
)

// Config holds all framework configuration loaded from environment variables.
type Config struct {
	// Browser is the target browser type. Env: BROWSER (default: chromium).
	Browser BrowserType
	// Headless runs the browser without visible UI. Env: HEADLESS (default: true).
	Headless bool
	// Trace enables Playwright tracing for debugging. Env: TRACE (default: false).
	Trace bool
	// BaseURL is the application base URL for navigation. Env: BASE_URL (default: https://example.com).
	BaseURL string
	// Timeout is the default operation timeout. Env: TIMEOUT_MS (default: 30000).
	Timeout time.Duration
	// SlowMo adds delay between operations in milliseconds. Env: SLOW_MO_MS (default: 0).
	SlowMo time.Duration
	// AllureReportDir is the directory for Allure results. Env: ALLURE_RESULTS_DIR (default: ./allure-results).
	AllureReportDir string
	// LogLevel is the logging level. Env: LOG_LEVEL (default: info).
	LogLevel string
	// ViewportWidth is the browser viewport width. Env: VIEWPORT_WIDTH (default: 1920).
	ViewportWidth int
	// ViewportHeight is the browser viewport height. Env: VIEWPORT_HEIGHT (default: 1080).
	ViewportHeight int
}

// Load reads configuration from environment variables with sensible defaults.
// It loads .env file from path specified in ENV_FILE if set.
func Load() *Config {
	_ = godotenv.Load(os.Getenv("ENV_FILE"))

	return &Config{
		Browser:         getBrowserType(),
		Headless:        getBool("HEADLESS", true),
		Trace:           getBool("TRACE", false),
		BaseURL:         getEnv("BASE_URL", "https://example.com"),
		Timeout:         getDuration("TIMEOUT_MS", defaultTimeoutMS),
		SlowMo:          getDuration("SLOW_MO_MS", defaultSlowMoMS),
		AllureReportDir: getEnv("ALLURE_RESULTS_DIR", "./allure-results"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ViewportWidth:   getInt("VIEWPORT_WIDTH", defaultViewportWidth),
		ViewportHeight:  getInt("VIEWPORT_HEIGHT", defaultViewportHeight),
	}
}

// getBrowserType parses BROWSER env var and returns BrowserType.
// Returns chromium for unrecognized values.
func getBrowserType() BrowserType {
	b := getEnv("BROWSER", "chrome")
	switch BrowserType(b) {
	case BrowserChromium:
		return BrowserChromium
	case BrowserFirefox:
		return BrowserFirefox
	case BrowserWebKit:
		return BrowserWebKit
	default:
		return BrowserChromium
	}
}

// getEnv returns the environment variable value or defaultVal if unset.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// getBool parses a boolean environment variable.
// Accepts true/false, 1/0. Returns defaultVal on parse error or if unset.
func getBool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultVal
	}
	return b
}

// getDuration parses a duration in milliseconds from environment variable.
// Returns defaultMs on parse error or if unset.
func getDuration(key string, defaultMs int) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(defaultMs)
	}
	ms, err := strconv.Atoi(v)
	if err != nil {
		return time.Duration(defaultMs)
	}
	return time.Duration(ms)
}

// getInt parses an integer from environment variable.
// Returns defaultVal on parse error or if unset.
func getInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return i
}
