package suite

import (
	"fmt"
	"log/slog"
	"testing"

	"apitests/internal/client"
	"apitests/internal/config"
	"apitests/internal/logger"
	"apitests/internal/reporter"
)

// TestSuite is the base struct for all test suites providing client, reporting, and logging.
type TestSuite struct {
	// T is the testing context for the current test.
	T *testing.T
	// Config is the loaded framework configuration.
	Config *config.Config
	// Client is the HTTP client for API requests.
	Client *client.HTTPClient
	// Reporter is the Allure reporter for test results.
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
	// Feature is the feature being tested (e.g., posts, users, todos).
	Feature string
}

// Severity levels for test prioritization.
const (
	// SeverityBlocker indicates a critical bug blocking test execution.
	SeverityBlocker = "blocker"
	// SeverityCritical indicates major functionality is broken.
	SeverityCritical = "critical"
	// SeverityNormal indicates a standard test case.
	SeverityNormal = "normal"
	// SeverityMinor indicates a minor issue or edge case.
	SeverityMinor = "minor"
	// SeverityTrivial indicates a cosmetic or low priority issue.
	SeverityTrivial = "trivial"
)

// New creates a new TestSuite with the specified testing context and suite name.
// It loads the framework configuration but does not initialize the client or reporter.
// Call Setup() to initialize them.
func New(t *testing.T, suiteName string) *TestSuite {
	cfg := config.Load()

	return &TestSuite{
		T:         t,
		Config:    cfg,
		SuiteName: suiteName,
	}
}

// Setup initializes the logger, HTTP client, and Allure reporter for a test.
// It creates a test-scoped logger and initializes the Allure reporter.
// Returns an error if setup fails; marks test as failed in that case.
func (s *TestSuite) Setup(testName string) error {
	s.Log = logger.ForTest(s.T, s.Config)

	s.Reporter = reporter.New(
		s.Config.AllureReportDir,
		testName,
		s.SuiteName,
		s.Log,
	)

	s.Client = client.New(s.Config.BaseURL, s.Config.Timeout, s.Log)

	s.Log.Info("Test setup complete", "test", testName)
	return nil
}

// Teardown handles cleanup and finalizes the Allure report.
// If testErr is not nil, it sets the test as failed in the reporter.
// It finalizes the Allure reporter and logs the teardown completion.
// Safe to call even if Setup failed.
func (s *TestSuite) Teardown(testName string, testErr *error) {
	if testErr != nil && *testErr != nil {
		s.Reporter.SetFailed(*testErr)
	}

	if err := s.Reporter.Finalize(); err != nil {
		s.Log.Warn("Could not finalize Allure report", "err", err)
	}

	s.Log.Info("Test teardown complete", "test", testName)
}

// Step executes fn as a named test step with Allure reporting.
// It starts a step, executes the function, and stops with passed or failed status.
// Returns an error with step name prefix if fn fails.
func (s *TestSuite) Step(name string, fn func() error) error {
	s.Reporter.StartStep(name)

	if err := fn(); err != nil {
		s.Reporter.StopStep(reporter.StatusFailed)
		return fmt.Errorf("step [%s] failed: %w", name, err)
	}

	s.Reporter.StopStep(reporter.StatusPassed)
	return nil
}

// label is an internal struct for tracking reporter labels.
type label struct {
	// key is the label name.
	key string
	// value is the label value.
	value string
}

// SetMeta sets test metadata in the Allure report.
// It sets the description and adds severity and feature labels.
// Empty values are skipped.
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
