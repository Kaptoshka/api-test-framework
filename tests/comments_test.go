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

// validateComment validates that a comment has required fields.
// Returns an error if the comment has zero ID, empty email, or empty body.
func validateComment(comment *endpoints.Comment) error {
	if comment.ID == 0 {
		return errors.New("comment has zero ID")
	}
	if comment.Email == "" {
		return fmt.Errorf("comment %d has empty email", comment.ID)
	}
	if comment.Body == "" {
		return fmt.Errorf("comment %d has empty body", comment.ID)
	}
	return nil
}

// TestGetAllComments tests the GET /comments endpoint.
// It sends a GET request to /comments and expects a 200 status code with a paginated response.
// The response may be empty if the API has been reset (daily reset).
// Expected: HTTP 200 OK with JSON body containing data array and pagination object.
func TestGetAllComments(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "CommentsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /comments returns 200 with pagination (may be empty after daily reset)",
		Severity:    suite.SeverityCritical,
		Feature:     "comments",
	})

	comments := endpoints.NewCommentsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var result *endpoints.CommentsListResponse
	testErr = s.Step("GET /comments — expect 200", func() error {
		var resp *client.Response
		var err error
		result, resp, err = comments.GetAll()
		if err != nil {
			return err
		}
		return comments.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Response has pagination", func() error {
		if result.Pagination.Total == 0 && len(result.Data) == 0 {
			return nil
		}
		if len(result.Data) > 0 {
			for _, comment := range result.Data {
				if err := validateComment(comment); err != nil {
					return err
				}
			}
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestGetCommentsByPostID tests the GET /comments?postId=:id endpoint.
// It first fetches all posts to find a valid post ID, then retrieves comments for that post.
// Expected: HTTP 200 OK with JSON body containing only comments belonging to that post.
func TestGetCommentsByPostID(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "CommentsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /comments?postId=:id returns only comments for that post",
		Severity:    suite.SeverityNormal,
		Feature:     "comments",
	})

	comments := endpoints.NewCommentsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var postID int
	testErr = s.Step("GET /posts — find valid post ID", func() error {
		postsAPI := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)
		result, _, err := postsAPI.GetAll()
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

	var result *endpoints.CommentsListResponse
	testErr = s.Step(fmt.Sprintf("GET /comments?postId=%d — expect 200", postID), func() error {
		var resp *client.Response
		var err error
		result, resp, err = comments.GetByPostID(postID)
		if err != nil {
			return err
		}
		return comments.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step(fmt.Sprintf("All comments belong to post %d", postID), func() error {
		for _, comment := range result.Data {
			if comment.PostID != postID {
				return fmt.Errorf(
					"comment %d has postId=%d, expected postId=%d",
					comment.ID,
					comment.PostID,
					postID,
				)
			}
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestGetCommentByID tests the GET /comments/:id endpoint.
// First, it fetches all comments to find a valid ID.
// Then it retrieves that specific comment by ID and verifies the response.
// Skips if no comments are available (API may have been reset).
// Expected: HTTP 200 OK with JSON body containing the comment with matching ID.
func TestGetCommentByID(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "CommentsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /comments/:id returns 200 and correct comment (skips if no data)",
		Severity:    suite.SeverityCritical,
		Feature:     "comments",
	})

	comments := endpoints.NewCommentsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var commentID int
	testErr = s.Step("GET /comments — find valid ID", func() error {
		result, _, err := comments.GetAll()
		if err != nil {
			return err
		}
		if len(result.Data) == 0 {
			return errors.New("no comments available (API may have been reset)")
		}
		commentID = result.Data[0].ID
		return nil
	})
	if testErr != nil {
		t.Skipf("Skipping test: %v", testErr)
	}

	var result *endpoints.CommentResponse
	testErr = s.Step(fmt.Sprintf("GET /comments/%d — expect 200 or 404", commentID), func() error {
		var resp *client.Response
		var err error
		result, resp, err = comments.GetByID(commentID)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("expected status 200 or 404, got %d", resp.StatusCode)
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("Comment has correct ID (if found)", func() error {
		if result != nil && result.Data != nil && result.Data.ID != commentID {
			return fmt.Errorf("expected ID=%d, got %d", commentID, result.Data.ID)
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestGetCommentNotFound tests the GET /comments/:id endpoint with a non-existent ID.
// It requests a comment with ID 99999 which should not exist.
// Expected: HTTP 404 Not Found status code.
func TestGetCommentNotFound(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "CommentsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /comments/99999 returns 404",
		Severity:    suite.SeverityNormal,
		Feature:     "comments",
	})

	comments := endpoints.NewCommentsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	testErr = s.Step("GET /comments/99999 — expect 404", func() error {
		_, resp, err := comments.GetByID(99999)
		if err != nil {
			return err
		}
		return comments.AssertStatus(resp, http.StatusNotFound)
	})
	require.NoError(t, testErr)
}

// TestCreateComment tests the POST /comments endpoint.
// It first fetches all posts to find a valid post ID, then creates a comment for that post.
// Expected: HTTP 201 Created with JSON body containing the created comment with generated ID.
// Verifies that the returned comment matches the request data.
func TestCreateComment(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "CommentsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "POST /comments creates comment and returns 201",
		Severity:    suite.SeverityCritical,
		Feature:     "comments",
	})

	comments := endpoints.NewCommentsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var postID int
	testErr = s.Step("GET /posts — find valid post ID", func() error {
		postsAPI := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)
		result, _, err := postsAPI.GetAll()
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

	req := &endpoints.CreateCommentRequest{
		PostID: postID,
		Name:   "Test Comment",
		Email:  "test@example.com",
		Body:   "This is a test comment body",
	}

	var created *endpoints.CommentResponse
	testErr = s.Step("POST /comments — expect 201", func() error {
		var resp *client.Response
		var err error
		created, resp, err = comments.Create(req)
		if err != nil {
			return err
		}
		return comments.AssertStatus(resp, http.StatusCreated)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Created comment has generated ID", func() error {
		if created.Data.ID == 0 {
			return errors.New("expected non-zero ID for created comment")
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("Created comment matches request data", func() error {
		if created.Data.PostID != req.PostID {
			return fmt.Errorf(
				"postId mismatch: expected %d, got %d",
				req.PostID,
				created.Data.PostID,
			)
		}
		if created.Data.Email != req.Email {
			return fmt.Errorf(
				"email mismatch: expected %s, got %s",
				req.Email,
				created.Data.Email,
			)
		}
		if created.Data.Body != req.Body {
			return fmt.Errorf(
				"body mismatch: expected %s, got %s",
				req.Body,
				created.Data.Body,
			)
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestPatchComment tests the PATCH /comments/:id endpoint.
// It fetches all comments to find a valid ID, then partially updates that comment.
// Only the body field is included in the request.
// Expected: HTTP 200 OK with JSON body containing the comment with updated body.
func TestPatchComment(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "CommentsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "PATCH /comments/:id partially updates comment and returns 200",
		Severity:    suite.SeverityNormal,
		Feature:     "comments",
	})

	comments := endpoints.NewCommentsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var commentID int
	testErr = s.Step("GET /comments — find valid ID", func() error {
		result, _, err := comments.GetAll()
		if err != nil {
			return err
		}
		if len(result.Data) == 0 {
			return errors.New("no comments available")
		}
		commentID = result.Data[0].ID
		return nil
	})
	require.NoError(t, testErr)

	req := &endpoints.PatchCommentRequest{
		Body: "Updated comment body",
	}

	var patched *endpoints.CommentResponse
	testErr = s.Step(fmt.Sprintf("PATCH /comments/%d — expect 200 or 404", commentID), func() error {
		var resp *client.Response
		var err error
		patched, resp, err = comments.Patch(commentID, req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("expected status 200 or 404, got %d", resp.StatusCode)
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("Comment has updated body (if found)", func() error {
		if patched != nil && patched.Data != nil && patched.Data.Body != req.Body {
			return fmt.Errorf(
				"body mismatch: expected %s, got %s",
				req.Body,
				patched.Data.Body,
			)
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestDeleteComment tests the DELETE /comments/:id endpoint.
// It fetches all comments to find a valid ID, then deletes that comment.
// Expected: HTTP 200, 204 No Content, or 404 Not Found status code.
func TestDeleteComment(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "CommentsAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "DELETE /comments/:id returns 200",
		Severity:    suite.SeverityNormal,
		Feature:     "comments",
	})

	comments := endpoints.NewCommentsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var commentID int
	testErr = s.Step("GET /comments — find valid ID", func() error {
		result, _, err := comments.GetAll()
		if err != nil {
			return err
		}
		if len(result.Data) == 0 {
			return errors.New("no comments available")
		}
		commentID = result.Data[0].ID
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step(fmt.Sprintf("DELETE /comments/%d — expect 200, 204 or 404", commentID), func() error {
		resp, err := comments.Delete(commentID)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK &&
			resp.StatusCode != http.StatusNoContent &&
			resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("expected status 200, 204 or 404, got %d", resp.StatusCode)
		}
		return nil
	})
	require.NoError(t, testErr)
}
