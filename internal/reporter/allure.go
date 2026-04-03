package reporter

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	uuidG "github.com/google/uuid"
)

// Status represents the test result status in Allure format.
type Status string

// Test result statuses.
const (
	// StatusPassed indicates the test passed successfully.
	StatusPassed Status = "passed"
	// StatusFailed indicates the test assertion failed.
	StatusFailed Status = "failed"
	// StatusBroken indicates a test infrastructure or framework error.
	StatusBroken Status = "broken"
	// StatusSkipped indicates the test was skipped.
	StatusSkipped Status = "skipped"
)

// Attachment represents a file attachment in the Allure report.
type Attachment struct {
	// Name is the display name of the attachment.
	Name string `json:"name"`
	// Source is the filename in the results directory.
	Source string `json:"source"`
	// Type is the MIME type (e.g., image/png).
	Type string `json:"type"`
}

// Step represents a test step in the Allure report hierarchy.
type Step struct {
	// Name is the step description.
	Name string `json:"name"`
	// Status is the step result status.
	Status Status `json:"status"`
	// Start is the step start time in milliseconds since epoch.
	Start int64 `json:"start"`
	// Stop is the step stop time in milliseconds since epoch.
	Stop int64 `json:"stop"`
	// Attachments are files attached to this step.
	Attachments []Attachment `json:"attachments"`
	// Steps are nested sub-steps.
	Steps []Step `json:"steps"`
}

// Label represents an Allure label for categorizing tests.
type Label struct {
	// Name is the label category (e.g., suite, feature).
	Name string `json:"name"`
	// Value is the label value.
	Value string `json:"value"`
}

// StatusDetails holds failure details for failed or broken tests.
type StatusDetails struct {
	// Message is the error message.
	Message string `json:"message,omitempty"`
	// Trace is the stack trace.
	Trace string `json:"trace,omitempty"`
}

// TestResult is the complete Allure test result structure.
type TestResult struct {
	// UUID is the unique test result identifier.
	UUID string `json:"uuid"`
	// HistoryID links to test history for trend tracking.
	HistoryID string `json:"historyId"`
	// FullName is the fully qualified test name (suite.test).
	FullName string `json:"fullName"`
	// Name is the test display name.
	Name string `json:"name"`
	// Status is the test result status.
	Status Status `json:"status"`
	// Start is the test start time in milliseconds since epoch.
	Start int64 `json:"start"`
	// Stop is the test stop time in milliseconds since epoch.
	Stop int64 `json:"stop"`
	// Description is the human-readable test description.
	Description string `json:"description,omitempty"`
	// Steps are the top-level test steps.
	Steps []Step `json:"steps"`
	// Attachments are files attached to the test.
	Attachments []Attachment `json:"attachments"`
	// Labels categorize the test.
	Labels []Label `json:"labels"`
	// StatusDetails contains failure information.
	StatusDetails *StatusDetails `json:"statusDetails.omitempty"`
}

// AllureReporter manages Allure report generation for a single test.
// It builds the test result structure incrementally and writes JSON on finalize.
type AllureReporter struct {
	// outputDir is the directory for result JSON files and attachments.
	outputDir string
	// result is the test result being built.
	result *TestResult
	// stepStack tracks nested steps for hierarchical reporting.
	stepStack []*Step
	// startTime is the test start time in milliseconds since epoch.
	startTime int64
	// log is the logger for reporter operations.
	log *slog.Logger
}

// New creates a new AllureReporter for a test with the specified output directory.
// It generates a UUID v7 for the test, sets initial status to passed,
// and adds default labels (suite, framework, language).
// Returns the configured reporter.
func New(outputDir, testName, suiteName string, log *slog.Logger) *AllureReporter {
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		log.Warn("Could not create allure results dir", "err", err)
	}

	uuid, err := uuidG.NewV7()
	if err != nil {
		log.Warn("Could not generate uuid", "err", err)
	}

	uuidStr := uuid.String()
	now := time.Now().UnixMilli()

	return &AllureReporter{
		outputDir: outputDir,
		startTime: now,
		log:       log,
		result: &TestResult{
			UUID:      uuidStr,
			HistoryID: testName,
			FullName:  fmt.Sprintf("%s.%s", suiteName, testName),
			Name:      testName,
			Status:    StatusPassed,
			Start:     now,
			Labels: []Label{
				{Name: "suite", Value: suiteName},
				{Name: "framework", Value: "playwright-go"},
				{Name: "language", Value: "golang"},
			},
		},
	}
}

// StartStep begins a new test step and pushes it to the step stack.
// If the stack is not empty, the step is nested under the current parent step.
// Otherwise, the step is added to the top-level test steps.
func (r *AllureReporter) StartStep(name string) {
	r.log.Info(name)
	step := &Step{
		Name:   name,
		Status: StatusPassed,
		Start:  time.Now().UnixMilli(),
	}

	if len(r.stepStack) > 0 {
		parent := r.stepStack[len(r.stepStack)-1]
		parent.Steps = append(parent.Steps, *step)
		r.stepStack = append(r.stepStack, &parent.Steps[len(parent.Steps)-1])
	} else {
		r.result.Steps = append(r.result.Steps, *step)
		r.stepStack = append(r.stepStack, &r.result.Steps[len(r.result.Steps)-1])
	}
}

// StopStep finalizes the current step with the given status and pops it from the stack.
// It sets the step stop time to the current time.
// Does nothing if the step stack is empty.
func (r *AllureReporter) StopStep(status Status) {
	if len(r.stepStack) == 0 {
		return
	}
	step := r.stepStack[len(r.stepStack)-1]
	step.Status = status
	step.Stop = time.Now().UnixMilli()
	r.stepStack = r.stepStack[:len(r.stepStack)-1]
}

// AddScreenshot saves the given PNG bytes to the output directory and attaches to the report.
// If the step stack is not empty, it attaches to the current step.
// Otherwise, it attaches to the test result directly.
// The filename is generated using UUID v7 with timestamp suffix.
// Returns an error if file writing fails.
func (r *AllureReporter) AddScreenshot(screenshotBytes []byte, name string) error {
	uuid, err := uuidG.NewV7()
	if err != nil {
		r.log.Warn("Could not generate uuid", "err", err)
	}
	filename := fmt.Sprintf("%s-%d", uuid.String(), time.Now().UnixMilli())
	destPath := filepath.Join(r.outputDir, filename)

	if err = os.WriteFile(destPath, screenshotBytes, 0o600); err != nil {
		return fmt.Errorf("failed to save screenshot attachment: %w", err)
	}

	attachment := Attachment{
		Name:   name,
		Source: filename,
		Type:   "image/png",
	}

	if len(r.stepStack) > 0 {
		step := r.stepStack[len(r.stepStack)-1]
		step.Attachments = append(step.Attachments, attachment)
	} else {
		r.result.Attachments = append(r.result.Attachments, attachment)
	}

	r.log.Info("Screenshot attached to allure report", "name", name)
	return nil
}

// SetFailed marks the test as failed with the given error message.
// It also marks all open steps as failed and sets their stop times.
// Use this when test assertions fail.
func (r *AllureReporter) SetFailed(err error) {
	r.result.Status = StatusFailed
	r.result.StatusDetails = &StatusDetails{
		Message: err.Error(),
	}

	for _, step := range r.stepStack {
		step.Status = StatusFailed
		step.Stop = time.Now().UnixMilli()
	}
}

// SetBroken marks the test as broken with the given error message.
// Use this when test setup or infrastructure fails (not assertion failures).
func (r *AllureReporter) SetBroken(err error) {
	r.result.Status = StatusBroken
	r.result.StatusDetails = &StatusDetails{
		Message: err.Error(),
	}
}

// AddLabel adds a custom label to the test result.
// Common labels include severity, feature, story, and epic.
func (r *AllureReporter) AddLabel(name, value string) {
	r.result.Labels = append(r.result.Labels, Label{
		Name:  name,
		Value: value,
	})
}

// SetDescription sets the human-readable test description in the report.
func (r *AllureReporter) SetDescription(desc string) {
	r.result.Description = desc
}

// Finalize sets the test stop time, marshals the result to JSON, and writes to disk.
// The output file is [outputDir]/[uuid]-result.json.
// Returns an error if marshaling or file write fails.
func (r *AllureReporter) Finalize() error {
	r.result.Stop = time.Now().UnixMilli()

	filename := fmt.Sprintf("%s-result.json", r.result.UUID)
	path := filepath.Join(r.outputDir, filename)

	data, err := json.MarshalIndent(r.result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal allure report: %w", err)
	}

	if err = os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write allure result: %w", err)
	}

	r.log.Info("Allure result written", "file", filename, "status", r.result.Status)
	return nil
}
