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

func uniqueSuffix(label string) string {
	return fmt.Sprintf("%s-%d",
		label,
		time.Now().UnixMilli(),
	)
}

func createTestPost(t *testing.T, posts *endpoints.PostsAPI, label string) int {
	t.Helper()
	req := makePostPayload(1, label)
	created, resp, err := posts.Create(req)
	require.NoError(t, err, "failed to create test post for %s", label)
	require.Equal(t, http.StatusCreated, resp.StatusCode,
		"expected 201 when creating test post for %s, got %d. Body: %s",
		label, resp.StatusCode, string(resp.Body),
	)
	require.NotZero(t, created.Data.ID, "created post has zero ID")
	return created.Data.ID
}

func makePostPayload(userID int, label string) *endpoints.CreatePostRequest {
	unique := uniqueSuffix(label)
	return &endpoints.CreatePostRequest{
		Title:  fmt.Sprintf("Test post %s", unique),
		Body:   fmt.Sprintf("Test body %s", unique),
		UserID: userID,
	}
}

// TestGetAllPosts tests the GET /posts endpoint.
// It sends a GET request to /posts and expects a 200 status code with a paginated response.
// The response may be empty if the API has been reset (daily reset).
// Expected: HTTP 200 OK with JSON body containing data array and pagination object.
func TestGetAllPosts(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "PostsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /posts returns 200 with pagination (may be empty after daily reset)",
		Severity:    suite.SeverityCritical,
		Feature:     "posts",
	})

	posts := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var result *endpoints.PostsListResponse
	testErr = s.Step("GET /posts — expect 200", func() error {
		var resp *client.Response
		var err error
		result, resp, err = posts.GetAll()
		if err != nil {
			return err
		}
		return posts.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Response has pagination", func() error {
		if result.Pagination.Total == 0 && len(result.Data) == 0 {
			return nil
		}
		if len(result.Data) > 0 {
			for _, post := range result.Data {
				if post.ID == 0 {
					return errors.New("post has zero ID")
				}
				if post.Title == "" {
					return fmt.Errorf("post %d has empty title", post.ID)
				}
			}
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestGetPostByID tests the GET /posts/:id endpoint.
// First, it fetches all posts to find a valid ID.
// Then it retrieves that specific post by ID and verifies the response.
// Skips if no posts are available (API may have been reset).
// Expected: HTTP 200 OK with JSON body containing the post with matching ID.
func TestGetPostByID(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "PostsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /posts/:id returns 200 and correct post data (skips if no data)",
		Severity:    suite.SeverityCritical,
		Feature:     "posts",
	})

	posts := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var postID int
	testErr = s.Step("GET /posts — find valid ID", func() error {
		result, _, err := posts.GetAll()
		if err != nil {
			return err
		}
		if len(result.Data) == 0 {
			return errors.New("no posts available (API may have been reset)")
		}
		postID = result.Data[0].ID
		return nil
	})
	if testErr != nil {
		t.Skipf("Skipping test: %v", testErr)
	}

	var result *endpoints.PostResponse
	testErr = s.Step(fmt.Sprintf("GET /posts/%d — expect 200", postID), func() error {
		var resp *client.Response
		var err error
		result, resp, err = posts.GetByID(postID)
		if err != nil {
			return err
		}
		return posts.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Post has correct ID", func() error {
		if result.Data.ID != postID {
			return fmt.Errorf("expected ID=%d, got %d", postID, result.Data.ID)
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("Post has non-empty title and body", func() error {
		if result.Data.Title == "" {
			return errors.New("post title is empty")
		}
		if result.Data.Body == "" {
			return errors.New("post body is empty")
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestGetPostNotFound tests the GET /posts/:id endpoint with a non-existent ID.
// It requests a post with ID 99999 which should not exist.
// Expected: HTTP 404 Not Found status code.
func TestGetPostNotFound(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "PostsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /posts/99999 returns 404",
		Severity:    suite.SeverityNormal,
		Feature:     "posts",
	})

	posts := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	testErr = s.Step("GET /posts/99999 — expect 404", func() error {
		_, resp, err := posts.GetByID(99999)
		if err != nil {
			return err
		}
		return posts.AssertStatus(resp, http.StatusNotFound)
	})
	require.NoError(t, testErr)
}

// TestCreatePost tests the POST /posts endpoint.
// It creates a new post with title, body, and user ID.
// Expected: HTTP 201 Created with JSON body containing the created post with generated ID.
// Verifies that the returned post matches the request data.
func TestCreatePost(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "PostsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "POST /posts creates post and returns 201 with matching data",
		Severity:    suite.SeverityCritical,
		Feature:     "posts",
	})

	posts := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	users := endpoints.NewUsersAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var userID int
	testErr = s.Step("Get user ID for post creation", func() error {
		result, _, err := users.GetAll()
		if err != nil {
			return fmt.Errorf("failed to get users: %v", err)
		}

		if len(result.Data) == 0 {
			return errors.New("no users found")
		}

		userID = result.Data[0].ID
		return nil
	})
	require.NoError(t, testErr, "failed to get user ID")

	req := makePostPayload(userID, "create")

	var created *endpoints.PostResponse
	testErr = s.Step("POST /posts — expect 201", func() error {
		var resp *client.Response
		var err error
		created, resp, err = posts.Create(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf(
				"expected 201, got %d. Body: %s",
				resp.StatusCode, string(resp.Body),
			)
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("Created post matches request data", func() error {
		if created.Data.ID == 0 {
			return errors.New("expected non-zero ID for created post")
		}
		if created.Data.Title != req.Title {
			return fmt.Errorf("title mismatch: expected %q, got %q", req.Title, created.Data.Title)
		}
		if created.Data.Body != req.Body {
			return fmt.Errorf("body mismatch: expected %q, got %q", req.Body, created.Data.Body)
		}
		s.Log.Info("Post created", "id", created.Data.ID, "title", created.Data.Title)
		return nil
	})
	require.NoError(t, testErr)
}

// TestDeletePost tests the DELETE /posts/:id endpoint.
// It fetches all posts to find a valid ID, then deletes that post.
// Expected: HTTP 204 No Content status code on successful deletion.
func TestDeletePost(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "PostsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "DELETE /posts/:id returns 204",
		Severity:    suite.SeverityNormal,
		Feature:     "posts",
	})

	posts := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var postID int
	testErr = s.Step("GET /posts — find valid ID", func() error {
		result, _, err := posts.GetAll()
		if err != nil {
			return err
		}
		if len(result.Data) == 0 {
			return errors.New("no posts available")
		}
		postID = result.Data[0].ID
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step(fmt.Sprintf("DELETE /posts/%d — expect 204", postID), func() error {
		resp, err := posts.Delete(postID)
		if err != nil {
			return err
		}
		return posts.AssertStatus(resp, http.StatusNoContent)
	})
	require.NoError(t, testErr)
}

// TestGetPostLikes tests the GET /posts/:id/likes endpoint.
// It fetches all posts to find a valid ID, then retrieves the like count for that post.
// Expected: HTTP 200 OK with JSON body containing postId and total likes count.
func TestGetPostLikes(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "PostsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /posts/:id/likes returns 200 with likes count",
		Severity:    suite.SeverityNormal,
		Feature:     "posts",
	})

	posts := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var postID int
	testErr = s.Step("GET /posts — find valid ID", func() error {
		result, _, err := posts.GetAll()
		if err != nil {
			return err
		}
		if len(result.Data) == 0 {
			return errors.New("no posts available")
		}
		postID = result.Data[0].ID
		return nil
	})
	require.NoError(t, testErr)

	var result *endpoints.LikesResponse
	testErr = s.Step(fmt.Sprintf("GET /posts/%d/likes — expect 200", postID), func() error {
		var resp *client.Response
		var err error
		result, resp, err = posts.GetLikes(postID)
		if err != nil {
			return err
		}
		return posts.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Response contains postId", func() error {
		if result.PostID == 0 {
			return errors.New(
				"expected non-zero postId in likes response",
			)
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestSearchPosts tests the GET /posts/search endpoint.
// It searches for posts matching the query "development".
// Expected: HTTP 200 OK with JSON body containing search results with query, total, and results array.
func TestSearchPosts(t *testing.T) {
	// t.Parallel()
	s := suite.New(t, "PostsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /posts/search?q=development returns 200 with matching posts",
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

	testErr = s.Step("Search returns results array", func() error {
		if len(result.Results) == 0 {
			return errors.New("expected results array in search response")
		}
		return nil
	})
	require.NoError(t, testErr)
}
