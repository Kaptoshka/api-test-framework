package pages

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"autotests/pkg/elements"

	"github.com/playwright-community/playwright-go"
)

// BasePage is the base struct for all Page Objects providing common navigation and element methods.
type BasePage struct {
	// Page is the Playwright page for interactions.
	Page playwright.Page
	// BaseURL is the application base URL for relative navigation.
	BaseURL string
	// Timeout is the default timeout for waits in milliseconds.
	Timeout time.Duration
	// Name is the page name for logging and error messages.
	Name string
	// Log is the page-scoped logger.
	Log *slog.Logger
}

// New creates a new BasePage with the specified dependencies.
func New(
	page playwright.Page,
	baseURL string,
	timeout time.Duration,
	name string,
	log *slog.Logger,
) *BasePage {
	return &BasePage{
		Page:    page,
		BaseURL: baseURL,
		Timeout: timeout,
		Name:    name,
		Log:     log,
	}
}

// Navigate opens the page at the given path relative to BaseURL.
// It waits for network idle before returning, ensuring page resources are loaded.
func (p *BasePage) Navigate(path string) error {
	url := p.BaseURL + path

	p.Log.Info("Navigating to", "url", url)

	if _, err := p.Page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   new(float64(p.Timeout)),
	}); err != nil {
		return fmt.Errorf("[%s] navigation FAILED: %w", p.Name, err)
	}

	return nil
}

// WaitForURL waits until the current URL matches the specified pattern.
// The pattern supports glob patterns (e.g., **/cart**).
// Returns error on timeout.
func (p *BasePage) WaitForURL(urlPattern string) error {
	p.Log.Info("Waiting for URL", "pattern", urlPattern)

	if err := p.Page.WaitForURL(urlPattern, playwright.PageWaitForURLOptions{
		Timeout: new(float64(p.Timeout)),
	}); err != nil {
		return fmt.Errorf("[%s] URL did not match [%s]: %w", p.Name, urlPattern, err)
	}

	return nil
}

// GetTitle returns the current page title.
// Returns empty string if title cannot be retrieved.
func (p *BasePage) GetTitle() (string, error) {
	title, err := p.Page.Title()
	if err != nil {
		return "", fmt.Errorf("[%s] could not get title: %w", p.Name, err)
	}

	return title, nil
}

// GetCurrentURL returns the absolute URL of the current page.
func (p *BasePage) GetCurrentURL() string {
	return p.Page.URL()
}

// CSS creates an Element with the specified CSS selector and description.
// The description is used for logging and error messages.
func (p *BasePage) CSS(selector, description string) *elements.Element {
	return elements.NewCSS(p.Page, selector, description, p.Timeout, p.Log)
}

// XPath creates an Element with the specified XPath expression and description.
// The description is used for logging and error messages.
func (p *BasePage) XPath(selector, description string) *elements.Element {
	return elements.NewXPath(p.Page, selector, description, p.Timeout, p.Log)
}

// WaitForNetworkIdle waits for all network requests to complete.
// Use this after navigation or actions that trigger network activity.
func (p *BasePage) WaitForNetworkIdle() error {
	p.Log.Debug("Waiting for network idle")

	if err := p.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: new(float64(p.Timeout)),
	}); err != nil {
		return fmt.Errorf("[%s] network did not become idle: %w", p.Name, err)
	}

	return nil
}

// ExecuteScript runs the specified JavaScript on the page and returns the result.
// Use this for complex DOM queries or page interactions not covered by Element methods.
func (p *BasePage) ExecuteScript(script string, args ...any) (any, error) {
	result, err := p.Page.Evaluate(script, args...)
	if err != nil {
		return nil, fmt.Errorf("[%s] script execution failed: %w", p.Name, err)
	}

	return result, nil
}

// ParseInt extracts all digits from the string and converts to integer.
// Useful for parsing prices with currency symbols (e.g., "1 500 ₽" -> 1500).
// Returns error if no digits found.
func (p *BasePage) ParseInt(s string) (int, error) {
	re := regexp.MustCompile(`[^0-9]`)
	cleanStr := re.ReplaceAllString(s, "")
	if cleanStr == "" {
		return 0, fmt.Errorf("string '%s' contains no digits", s)
	}
	return strconv.Atoi(cleanStr)
}
