package api_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"apitests/internal/client"
	"apitests/pkg/suite"
	"apitests/tests/endpoints"

	"github.com/stretchr/testify/require"
)

// TestAdvancedFiltering tests the GET /posts endpoint with advanced filtering.
// It filters posts by title containing "web" and sorts by ID in descending order.
// Expected: HTTP 200 OK with JSON body containing filtered and sorted posts.
func TestAdvancedFiltering(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "PostsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /posts?title_like=web&_sort=id&_order=desc filters and sorts posts",
		Severity:    suite.SeverityNormal,
		Feature:     "posts",
	})

	posts := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var result *endpoints.PostsListResponse
	testErr = s.Step("GET /posts?title_like=web&_sort=id&_order=desc — expect 200", func() error {
		var resp *client.Response
		var err error
		result, resp, err = posts.GetWithFilter("web", "id", "desc")
		if err != nil {
			return err
		}
		return posts.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Results match filter criteria", func() error {
		if len(result.Data) == 0 {
			return nil
		}
		for _, post := range result.Data {
			if post.ID == 0 {
				return errors.New("post has zero ID")
			}
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("Results are sorted by ID descending", func() error {
		if len(result.Data) < 2 {
			return nil
		}
		for i := range len(result.Data) - 1 {
			if result.Data[i].ID < result.Data[i+1].ID {
				return fmt.Errorf("results not sorted in descending order: %d > %d",
					result.Data[i].ID, result.Data[i+1].ID)
			}
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestPostSearch tests the GET /posts/search endpoint.
// It searches for posts with query "development" which should match title or body.
// Expected: HTTP 200 OK with JSON body containing matching posts.
func TestPostSearch(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "PostsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /posts/search?q=development searches posts by title and content",
		Severity:    suite.SeverityNormal,
		Feature:     "posts",
	})

	posts := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var result *endpoints.SearchPostsResponse
	testErr = s.Step("GET /posts/search?q=development — expect 200", func() error {
		var resp *client.Response
		var err error
		result, resp, err = posts.Search("development")
		if err != nil {
			return err
		}
		return posts.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Search returns results with query info", func() error {
		if result.Query == "" {
			return errors.New("expected query field in response")
		}
		if result.Total == 0 && len(result.Results) == 0 {
			return nil
		}
		if len(result.Results) > 0 {
			for _, post := range result.Results {
				if post.ID == 0 {
					return errors.New("post has zero ID")
				}
			}
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestUserSearch tests the GET /users/search endpoint.
// It searches for users with query "john" which should match name, username, or email.
// Expected: HTTP 200 OK with JSON body containing matching users.
func TestUserSearch(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "UsersAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /users/search?q=john searches users by name, username, or email",
		Severity:    suite.SeverityNormal,
		Feature:     "users",
	})

	users := endpoints.NewUsersAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var result *endpoints.UsersListResponse
	testErr = s.Step("GET /users/search?q=john — expect 200", func() error {
		var resp *client.Response
		var err error
		result, resp, err = users.SearchUsers("john")
		if err != nil {
			return err
		}
		return users.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Search returns users matching criteria", func() error {
		if len(result.Data) == 0 {
			return nil
		}
		for _, user := range result.Data {
			if user.ID == 0 {
				return errors.New("user has zero ID")
			}
			found := false
			if user.Name != "" {
				found = true
			}
			if user.Username != "" {
				found = true
			}
			if user.Email != "" {
				found = true
			}
			if !found {
				return fmt.Errorf("user %d has empty fields", user.ID)
			}
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestResponseDelay tests the GET /posts endpoint with delay parameter.
// It requests posts with a 2-second delay to simulate loading states.
// Expected: HTTP 200 OK with valid post data after approximately 2 seconds.
func TestResponseDelay(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "PostsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /posts?_delay=2000 simulates loading states for frontend testing",
		Severity:    suite.SeverityMinor,
		Feature:     "posts",
	})

	posts := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	delayMs := 2000
	startTime := time.Now()

	var result *endpoints.PostsListResponse
	testErr = s.Step(fmt.Sprintf("GET /posts?_delay=%d — expect 200", delayMs), func() error {
		var resp *client.Response
		var err error
		result, resp, err = posts.GetWithDelay(delayMs)
		if err != nil {
			return err
		}
		return posts.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	elapsed := time.Since(startTime)

	testErr = s.Step("Response took approximately the requested delay", func() error {
		minExpected := time.Duration(delayMs-200) * time.Millisecond
		maxExpected := time.Duration(delayMs+500) * time.Millisecond
		if elapsed < minExpected || elapsed > maxExpected {
			return fmt.Errorf("response time %v not within expected range [%v, %v]",
				elapsed, minExpected, maxExpected)
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("Response contains valid post data", func() error {
		if result.Pagination.Total == 0 && len(result.Data) == 0 {
			return nil
		}
		if len(result.Data) > 0 {
			for _, post := range result.Data {
				if post.ID == 0 {
					return errors.New("post has zero ID")
				}
			}
		}
		return nil
	})
	require.NoError(t, testErr)
}
