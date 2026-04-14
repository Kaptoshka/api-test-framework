package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	neturl "net/url"
	"time"
)

// HTTPClient is an HTTP client for making REST API requests.
// It wraps the standard [http.Client] and provides methods for common HTTP operations.
type HTTPClient struct {
	// baseURL is the base URL for all API requests.
	baseURL string
	// httpClient is the underlying HTTP client.
	httpClient *http.Client
	// log is the logger for request/response logging.
	log *slog.Logger
}

// New creates a new HTTPClient with the specified base URL, timeout, and logger.
func New(baseURL string, timeout time.Duration, log *slog.Logger) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
		log:        log,
	}
}

// Response represents an HTTP response from the API.
type Response struct {
	// StatusCode is the HTTP status code returned by the server.
	StatusCode int
	// Body is the response body as raw bytes.
	Body []byte
	// Headers is the response headers.
	Headers http.Header
}

// Get performs an HTTP GET request to the specified path.
// Returns the response or an error if the request fails.
func (c *HTTPClient) Get(path string) (*Response, error) {
	return c.do(http.MethodGet, path, nil)
}

// Post performs an HTTP POST request to the specified path with the given body.
// Returns the response or an error if the request fails.
func (c *HTTPClient) Post(path string, body any) (*Response, error) {
	return c.do(http.MethodPost, path, body)
}

// Put performs an HTTP PUT request to the specified path with the given body.
// Returns the response or an error if the request fails.
func (c *HTTPClient) Put(path string, body any) (*Response, error) {
	return c.do(http.MethodPut, path, body)
}

// Delete performs an HTTP DELETE request to the specified path.
// Returns the response or an error if the request fails.
func (c *HTTPClient) Delete(path string) (*Response, error) {
	return c.do(http.MethodDelete, path, nil)
}

// Patch performs an HTTP PATCH request to the specified path with the given body.
// Returns the response or an error if the request fails.
func (c *HTTPClient) Patch(path string, body any) (*Response, error) {
	return c.do(http.MethodPatch, path, body)
}

// do performs the actual HTTP request with the specified method, path, and body.
// It handles request creation, execution, and response reading.
func (c *HTTPClient) do(method, path string, body any) (*Response, error) {
	url := c.baseURL + path

	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return nil, fmt.Errorf("invalid url scheme: %s", parsedURL.Scheme)
	}

	c.log.Info("HTTP request", "method", method, "url", url)

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.log.Info("HTTP response", "status", resp.StatusCode, "url", url)

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

// Decode unmarshals the response body into the given target.
// Returns an error if unmarshaling fails.
func (r *Response) Decode(target any) error {
	return json.Unmarshal(r.Body, target)
}
