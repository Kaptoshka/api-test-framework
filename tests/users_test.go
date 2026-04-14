package api_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"apitests/internal/client"
	"apitests/pkg/suite"
	"apitests/tests/endpoints"

	"github.com/stretchr/testify/require"
)

// TestGetAllUsers tests the GET /users endpoint.
// It sends a GET request to /users and expects a 200 status code.
// Expected: HTTP 200 OK with JSON body containing non-empty data array and pagination.
func TestGetAllUsers(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "UsersAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /users returns 200 and non-empty list with pagination",
		Severity:    suite.SeverityCritical,
		Feature:     "users",
	})

	users := endpoints.NewUsersAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var result *endpoints.UsersListResponse
	testErr = s.Step("GET /users — expect 200", func() error {
		var resp *client.Response
		var err error
		result, resp, err = users.GetAll()
		if err != nil {
			return err
		}
		return users.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Response contains data array", func() error {
		if len(result.Data) == 0 {
			return errors.New("expected non-empty users list")
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("Response contains pagination", func() error {
		if result.Pagination.Total == 0 {
			return errors.New("expected pagination.total > 0")
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestGetUserByID tests the GET /users/:id endpoint.
// First, it fetches all users to find a valid ID.
// Then it retrieves that specific user by ID and verifies the response.
// Skips if no users are available (API may have been reset).
// Expected: HTTP 200 OK with JSON body containing the user with matching ID and non-empty name/email.
func TestGetUserByID(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "UsersAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /users/:id returns 200 and correct user data (skips if no data)",
		Severity:    suite.SeverityCritical,
		Feature:     "users",
	})

	users := endpoints.NewUsersAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var userID int
	testErr = s.Step("GET /users — find valid ID", func() error {
		result, _, err := users.GetAll()
		if err != nil {
			return err
		}
		if len(result.Data) == 0 {
			return errors.New("no users available (API may have been reset)")
		}
		userID = result.Data[0].ID
		return nil
	})
	if testErr != nil {
		t.Skipf("Skipping test: %v", testErr)
	}

	var result *endpoints.UserResponse
	testErr = s.Step(fmt.Sprintf("GET /users/%d — expect 200", userID), func() error {
		var resp *client.Response
		var err error
		result, resp, err = users.GetByID(userID)
		if err != nil {
			return err
		}
		return users.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("User has correct ID", func() error {
		if result.Data.ID != userID {
			return fmt.Errorf(
				"expected ID=%d, got %d",
				userID,
				result.Data.ID,
			)
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("User has non-empty name and email", func() error {
		if result.Data.Name == "" {
			return errors.New("user name is empty")
		}
		if result.Data.Email == "" {
			return errors.New("user email is empty")
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestGetUserNotFound tests the GET /users/:id endpoint with a non-existent ID.
// It requests a user with ID 99999 which should not exist.
// Expected: HTTP 404 Not Found status code.
func TestGetUserNotFound(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "UsersAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /users/99999 returns 404",
		Severity:    suite.SeverityNormal,
		Feature:     "users",
	})

	users := endpoints.NewUsersAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	testErr = s.Step("GET /users/99999 — expect 404", func() error {
		_, resp, err := users.GetByID(99999)
		if err != nil {
			return err
		}
		return users.AssertStatus(resp, http.StatusNotFound)
	})
	require.NoError(t, testErr)
}

// TestDeleteUser tests the DELETE /users/:id endpoint.
// It fetches all users to find a valid ID, then deletes that user.
// Expected: HTTP 204 No Content status code on successful deletion.
func TestDeleteUser(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "UsersAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "DELETE /users/:id returns 204",
		Severity:    suite.SeverityNormal,
		Feature:     "users",
	})

	users := endpoints.NewUsersAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var userID int
	testErr = s.Step("GET /users — find valid ID", func() error {
		result, _, err := users.GetAll()
		if err != nil {
			return err
		}
		if len(result.Data) == 0 {
			return errors.New("no users available")
		}
		userID = result.Data[0].ID
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step(fmt.Sprintf("DELETE /users/%d — expect 204", userID), func() error {
		resp, err := users.Delete(userID)
		if err != nil {
			return err
		}
		return users.AssertStatus(resp, http.StatusNoContent)
	})
	require.NoError(t, testErr)
}
