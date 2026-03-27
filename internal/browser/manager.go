package browser

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"autotests/internal/config"

	"github.com/playwright-community/playwright-go"
)

// Manager manages the browser and page lifecycle.
type Manager struct {
	// pw is the Playwright instance that controls browser execution.
	pw *playwright.Playwright
	// browser is the active browser instance (Chromium, Firefox, or WebKit).
	browser playwright.Browser
	// context is the browser context that holds page state and viewport settings.
	context playwright.BrowserContext
	// Page is the current page for interactions.
	Page playwright.Page
	// cfg is the framework configuration.
	cfg *config.Config
	// log is the logger for browser operations.
	log *slog.Logger
}

// New creates a new instance of Manager without launching the browser.
// Call Launch() separately to initialize the browser.
func New(cfg *config.Config, log *slog.Logger) *Manager {
	return &Manager{
		cfg: cfg,
		log: log,
	}
}

// Launch starts the browser and creates a new page with configured timeouts.
// It initializes Playwright, launches the configured browser type, creates
// a browser context with the specified viewport, and sets up default timeouts.
// Tracing is enabled automatically if TRACE=true environment variable is set.
func (m *Manager) Launch() error {
	m.log.Info("Launching Playwright...")

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start Playwright: %w", err)
	}
	m.pw = pw

	browserType, err := m.getBrowserType()
	if err != nil {
		return err
	}

	m.log.Info("Starting browser", "browser", m.cfg.Browser, "headless", m.cfg.Headless)

	browser, err := browserType.Launch(m.getBrowserLaunchOptions())
	if err != nil {
		return fmt.Errorf("could not launch browser: %w", err)
	}
	m.browser = browser

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  m.cfg.ViewportWidth,
			Height: m.cfg.ViewportHeight,
		},
	})
	if err != nil {
		return fmt.Errorf("could not create browser context: %w", err)
	}
	m.context = context

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}
	m.Page = page

	page.SetDefaultNavigationTimeout(float64(m.cfg.Timeout.Milliseconds()))
	page.SetDefaultTimeout(float64(m.cfg.Timeout.Milliseconds()))

	m.log.Info("Browser launched successfully")

	if os.Getenv("TRACE") == "true" {
		if err = m.context.Tracing().Start(playwright.TracingStartOptions{
			Screenshots: new(true),
			Snapshots:   new(true),
			Sources:     new(true),
		}); err != nil {
			m.log.Warn("Failed to start tracing", "error", err)
		}
	}

	return nil
}

// Close closes the browser, saves trace if TRACE=true, and releases all resources.
// Tracing is saved to artifacts/traces/ directory with timestamp-based naming.
func (m *Manager) Close() error {
	if os.Getenv("TRACE") == "true" {
		traceDir := config.DefaultTracesDir
		tracePath := fmt.Sprintf("%s/trace-%d.zip", traceDir, time.Now().UnixMilli())
		_ = os.Mkdir(traceDir, 0o750)
		if err := m.context.Tracing().Stop(tracePath); err != nil {
			m.log.Warn("Failed to save trace", "error", err)
		} else {
			m.log.Info("Trace saved", "path", tracePath)
		}
	}

	if m.context != nil {
		m.log.Debug("Closing browser context")
		err := m.context.Close()
		if err != nil {
			m.log.Error("Failed to close browser context", "error", err)
		}
	}

	if m.browser != nil {
		m.log.Debug("Closing browser")
		err := m.browser.Close()
		if err != nil {
			m.log.Error("Failed to close browser", "error", err)
		}
	}

	if m.pw != nil {
		m.log.Debug("Stopping playwright")
		err := m.pw.Stop()
		if err != nil {
			m.log.Error("Failed to stop Playwright", "error", err)
			return fmt.Errorf("failed to stop Playwright: %w", err)
		}
	}

	return nil
}

// NavigateTo navigates to the specified absolute URL.
// It waits for network idle before returning, ensuring page resources are loaded.
func (m *Manager) NavigateTo(url string) error {
	m.log.Info("Navigating to", "url", url)
	if _, err := m.Page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("navigation to %s failed: %w", url, err)
	}
	return nil
}

// getBrowserType returns the appropriate Playwright browser type based on config.
// Returns chromium for unrecognized browser values.
func (m *Manager) getBrowserType() (playwright.BrowserType, error) {
	switch m.cfg.Browser {
	case config.BrowserFirefox:
		return m.pw.Firefox, nil
	case config.BrowserChromium:
		return m.pw.Chromium, nil
	case config.BrowserWebKit:
		return m.pw.WebKit, nil
	default:
		return nil, fmt.Errorf("unsupported browser: %s", m.cfg.Browser)
	}
}

// getBrowserLaunchOptions returns the launch options for the configured browser.
// It reads custom executable paths from PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH,
// PLAYWRIGHT_FIREFOX_EXECUTABLE_PATH, and PLAYWRIGHT_WEBKIT_EXECUTABLE_PATH env vars.
func (m *Manager) getBrowserLaunchOptions() playwright.BrowserTypeLaunchOptions {
	headless := m.cfg.Headless
	slowMo := float64(m.cfg.SlowMo.Milliseconds())

	opts := playwright.BrowserTypeLaunchOptions{
		Headless: &headless,
		SlowMo:   &slowMo,
	}

	switch m.cfg.Browser {
	case config.BrowserChromium:
		if execPath := os.Getenv("PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH"); execPath != "" {
			m.log.Info("Using custom Chromium path", "path", execPath)
			opts.ExecutablePath = &execPath
		}
	case config.BrowserFirefox:
		if execPath := os.Getenv("PLAYWRIGHT_FIREFOX_EXECUTABLE_PATH"); execPath != "" {
			m.log.Info("Using custom Firefox path", "path", execPath)
			opts.ExecutablePath = &execPath
		}
	case config.BrowserWebKit:
		if execPath := os.Getenv("PLAYWRIGHT_WEBKIT_EXECUTABLE_PATH"); execPath != "" {
			m.log.Info("Using custom WebKit path", "path", execPath)
			opts.ExecutablePath = &execPath
		}
	default:
		m.log.Warn("Unsupported browser", "browser", m.cfg.Browser)
	}

	return opts
}
