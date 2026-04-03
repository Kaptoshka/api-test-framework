package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"apitests/internal/client"
)

// BaseClient is the base client for API tests providing HTTP client and assertion methods.
type BaseClient struct {
	// Client is the HTTP client for making requests.
	Client *client.HTTPClient
	// Log is the logger for the API client.
	Log *slog.Logger
	// Name is the name of the API for logging and reporting.
	Name string
}

// New creates a new BaseClient with the specified base URL, timeout, name, and logger.
func New(
	baseURL string,
	timeout time.Duration,
	name string,
	log *slog.Logger,
) *BaseClient {
	return &BaseClient{
		Client: client.New(baseURL, timeout, log),
		Log:    log.With("api", name),
		Name:   name,
	}
}

// AssertStatus asserts that the response status code matches the expected value.
// Returns an error if the status code does not match.
func (c *BaseClient) AssertStatus(resp *client.Response, expected int) error {
	if resp.StatusCode != expected {
		return fmt.Errorf(
			"[%s] expected status %d, got %d. Body: %s",
			c.Name, expected, resp.StatusCode, string(resp.Body),
		)
	}
	return nil
}

// AssertStatusOK asserts that the response status code is 200 OK.
// Returns an error if the status code is not 200.
func (c *BaseClient) AssertStatusOK(resp *client.Response) error {
	return c.AssertStatus(resp, http.StatusOK)
}

// AssertStatusCreated asserts that the response status code is 201 Created.
// Returns an error if the status code is not 201.
func (c *BaseClient) AssertStatusCreated(resp *client.Response) error {
	return c.AssertStatus(resp, http.StatusCreated)
}
