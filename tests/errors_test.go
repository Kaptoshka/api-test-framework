package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"apitests/pkg/suite"

	"github.com/stretchr/testify/require"
)

// TestError404Endpoint tests the GET /error/404 endpoint.
// It requests the /error/404 endpoint which should return a 404 error.
// Expected: HTTP 404 Not Found status code.
func TestError404Endpoint(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "ErrorsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /error/404 returns 404 status code",
		Severity:    suite.SeverityNormal,
		Feature:     "errors",
	})

	testErr = s.Step("GET /error/404 — expect 404", func() error {
		resp, err := s.Client.Get("/error/404")
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("expected 404, got %d", resp.StatusCode)
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestError500Endpoint tests the GET /error/500 endpoint.
// It requests the /error/500 endpoint which should return a 500 server error.
// Expected: HTTP 500 Internal Server Error status code.
func TestError500Endpoint(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "ErrorsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /error/500 returns 500 status code",
		Severity:    suite.SeverityNormal,
		Feature:     "errors",
	})

	testErr = s.Step("GET /error/500 — expect 500", func() error {
		resp, err := s.Client.Get("/error/500")
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusInternalServerError {
			return fmt.Errorf("expected 500, got %d", resp.StatusCode)
		}
		return nil
	})
	require.NoError(t, testErr)
}
