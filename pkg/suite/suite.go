package suite

import (
	"fmt"
	"log/slog"
	"testing"

	"autotests/internal/browser"
	"autotests/internal/config"
	"autotests/internal/logger"
	"autotests/internal/reporter"
	"autotests/internal/screenshot"
)

// TestSuite is the base struct for all test suites providing browser, reporting, and logging.
type TestSuite struct {
	// T is the testing context for the current test.
	T *testing.T
	// Config is the loaded framework configuration.
	Config *config.Config
	// Browser is the browser instance for the test.
	Browser *browser.Manager
	// Screenshot is the screenshot capture service.
	Screenshot *screenshot.Service
	// Reporter is the Allure report for test results.
	Reporter *reporter.AllureReporter
	// SuiteName is the name of the test suite for reporting.
	SuiteName string
	// Log is the test-scoped logger.
	Log *slog.Logger
}

// TestMeta contains metadata for a test to be recorded in Allure report.
type TestMeta struct {
	// Description is the human-readable test description.
	Description string
	// Severity is the impact level (blocker, critical, normal, minor, trivial).
	Severity string
	// Feature is the feature being tested (e.g., cart, search, wishlist).
	Feature string
}

// Severity levels for test prioritization.
const (
	SeverityBlocker  = "blocker"  // Critical bug blocking test execution
	SeverityCritical = "critical" // Major functionality broken
	SeverityNormal   = "normal"   // Standard test case
	SeverityMinor    = "minor"    // Minor issue or edge case
	SeverityTrivial  = "trivial"  // Cosmetic or low priority
)

// New creates a configured test suite without launching the browser.
// Call Setup() to initialize browser and reporters.
func New(t *testing.T, suiteName string) *TestSuite {
	cfg := config.Load()

	return &TestSuite{
		T:         t,
		Config:    cfg,
		SuiteName: suiteName,
	}
}

// Setup initializes the browser, logger, and reporters for a test.
// It creates a test-scoped logger, launches the browser, and initializes Allure reporter.
// Returns error if browser launch fails; marks test as broken in that case.
func (s *TestSuite) Setup(testName string) error {
	s.Log = logger.ForTest(s.T)
	s.Screenshot = screenshot.New(s.Log)
	s.Reporter = reporter.New(s.Config.AllureReportDir, testName, s.SuiteName, s.Log)

	s.Browser = browser.New(s.Config, s.Log)
	if err := s.Browser.Launch(); err != nil {
		s.Reporter.SetBroken(err)
		_ = s.Reporter.Finalize()
		return fmt.Errorf("browser setup failed: %w", err)
	}

	s.Log.Info("Test setup complete", "test", testName)
	return nil
}

// Teardown handles cleanup, captures screenshot on failure, and finalizes reports.
// It captures a screenshot if testErr is not nil, finalizes the Allure report,
// and closes the browser. Safe to call even if Setup failed.
func (s *TestSuite) Teardown(testName string, testErr *error) {
	if testErr != nil && *testErr != nil {
		s.Log.Warn("Test FAILED -- capturing screenshot", "test", testName)
		if bytes, err := s.Screenshot.CaptureAsBites(s.Browser.Page); err == nil {
			_ = s.Reporter.AddScreenshot(bytes, fmt.Sprintf("Failure: %s", testName))
		} else {
			s.Log.Warn("Failed to capture screenshot", "err", err)
		}
		s.Reporter.SetFailed(*testErr)
	}

	if err := s.Reporter.Finalize(); err != nil {
		s.Log.Warn("Could not finalize Allure report", "err", err)
	}

	err := s.Browser.Close()
	if err != nil {
		s.Log.Warn("Could not close browser", "err", err)
	}

	s.Log.Info("Test teardown complete", "test", testName)
}

// Step executes fn as a named test step with Allure reporting.
// It starts a step, executes the function, and stops with passed/failed status.
// Returns error with step name prefix if fn fails.
func (s *TestSuite) Step(name string, fn func() error) error {
	s.Reporter.StartStep(name)

	if err := fn(); err != nil {
		s.Reporter.StopStep(reporter.StatusFailed)
		return fmt.Errorf("step [%s] failed: %w", name, err)
	}

	s.Reporter.StopStep(reporter.StatusPassed)
	return nil
}

// NavigateTo opens the specified absolute URL using the managed browser page.
// Convenience method wrapping browser.NavigateTo.
func (s *TestSuite) NavigateTo(url string) error {
	return s.Browser.NavigateTo(url)
}

// label is an internal struct for tracking reporter labels.
type label struct {
	// key is the label name.
	key string
	// value is the label value.
	value string
}

// SetMeta sets test metadata in the Allure report.
// It sets the description and adds severity/feature labels.
// Skips empty values.
func (s *TestSuite) SetMeta(meta TestMeta) {
	if meta.Description != "" {
		s.Reporter.SetDescription(meta.Description)
	}

	labels := []label{
		{"severity", meta.Severity},
		{"feature", meta.Feature},
	}

	for _, l := range labels {
		if l.value != "" {
			s.Reporter.AddLabel(l.key, l.value)
		}
	}
}
