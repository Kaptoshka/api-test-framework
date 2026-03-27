package screenshot

import (
	"fmt"
	"log/slog"

	"github.com/playwright-community/playwright-go"
)

// Service provides screenshot capture functionality for test reporting.
type Service struct {
	// log is the logger for capture operations.
	log *slog.Logger
}

// New creates a new screenshot service with the specified logger.
func New(log *slog.Logger) *Service {
	return &Service{log: log}
}

// CaptureAsBites captures a full-page screenshot of the page as PNG bytes.
// Returns raw bytes suitable for Allure attachments or further processing.
func (s *Service) CaptureAsBites(page playwright.Page) ([]byte, error) {
	bytes, err := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: new(true),
	})
	if err != nil {
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}
	return bytes, nil
}
