package endpoints

import (
	"fmt"
	"log/slog"
	"time"

	"apitests/internal/client"
	"apitests/pkg/api"
)

// CommentsAPI provides methods for interacting with the Comments API.
type CommentsAPI struct {
	*api.BaseClient
}

// NewCommentsAPI creates a new CommentsAPI client with the specified base URL, timeout, and logger.
func NewCommentsAPI(baseURL string, timeout time.Duration, log *slog.Logger) *CommentsAPI {
	return &CommentsAPI{
		BaseClient: api.New(baseURL, timeout, "CommentsAPI", log),
	}
}

// Comment represents a comment in the API.
type Comment struct {
	// ID is the unique identifier of the comment.
	ID int `json:"id"`
	// PostID is the ID of the post the comment belongs to.
	PostID int `json:"postId"`
	// Name is the name of the comment author.
	Name string `json:"name"`
	// Email is the email address of the comment author.
	Email string `json:"email"`
	// Body is the content of the comment.
	Body string `json:"body"`
}

// CreateCommentRequest represents a request to create a new comment.
type CreateCommentRequest struct {
	// PostID is the ID of the post the comment belongs to.
	PostID int `json:"postId"`
	// Name is the name of the comment author.
	Name string `json:"name"`
	// Email is the email address of the comment author.
	Email string `json:"email"`
	// Body is the content of the new comment.
	Body string `json:"body"`
}

// PatchCommentRequest represents a request to patch an existing comment.
type PatchCommentRequest struct {
	// Name is the new name for the comment author (optional).
	Name string `json:"name,omitempty"`
	// Email is the new email for the comment author (optional).
	Email string `json:"email,omitempty"`
	// Body is the new content for the comment (optional).
	Body string `json:"body,omitempty"`
}

// CommentsListResponse represents the response for listing comments.
type CommentsListResponse struct {
	// Data is the list of comments.
	Data []*Comment `json:"data"`
	// Pagination contains pagination information.
	Pagination Pagination `json:"pagination"`
}

// CommentResponse represents the response for a single comment.
type CommentResponse struct {
	// Data is the comment data.
	Data *Comment `json:"data"`
}

// GetAll retrieves all comments from the API.
// Returns the list of comments, the raw response, or an error if the request fails.
func (a *CommentsAPI) GetAll() (*CommentsListResponse, *client.Response, error) {
	resp, err := a.Client.Get("/comments")
	if err != nil {
		return nil, nil, fmt.Errorf("GET /comments failed: %w", err)
	}
	var result CommentsListResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode comments list: %w", err)
	}
	return &result, resp, nil
}

// GetByPostID retrieves all comments for a specific post from the API.
// Returns the list of comments, the raw response, or an error if the request fails.
func (a *CommentsAPI) GetByPostID(
	postID int,
) (*CommentsListResponse, *client.Response, error) {
	resp, err := a.Client.Get(fmt.Sprintf("/comments?postId=%d", postID))
	if err != nil {
		return nil, nil, fmt.Errorf("GET /comments?postId=%d failed: %w", postID, err)
	}
	var result CommentsListResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode comments by postId: %w", err)
	}
	return &result, resp, nil
}

// GetByID retrieves a comment by its ID from the API.
// Returns the comment, the raw response, or an error if the request fails.
func (a *CommentsAPI) GetByID(id int) (*CommentResponse, *client.Response, error) {
	resp, err := a.Client.Get(fmt.Sprintf("/comments/%d", id))
	if err != nil {
		return nil, nil, fmt.Errorf("GET /comments/%d failed: %w", id, err)
	}
	var result CommentResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode comment: %w", err)
	}
	return &result, resp, nil
}

// Create creates a new comment with the given request data.
// Returns the created comment, the raw response, or an error if the request fails.
func (a *CommentsAPI) Create(
	req *CreateCommentRequest,
) (*CommentResponse, *client.Response, error) {
	resp, err := a.Client.Post("/comments", req)
	if err != nil {
		return nil, nil, fmt.Errorf("POST /comments failed: %w", err)
	}
	var result CommentResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode created comment: %w", err)
	}
	return &result, resp, nil
}

// Update fully updates a comment with the given ID using the request data.
// Returns the updated comment, the raw response, or an error if the request fails.
func (a *CommentsAPI) Update(
	id int,
	req *CreateCommentRequest,
) (*CommentResponse, *client.Response, error) {
	resp, err := a.Client.Put(fmt.Sprintf("/comments/%d", id), req)
	if err != nil {
		return nil, nil, fmt.Errorf("PUT /comments/%d failed: %w", id, err)
	}
	var result CommentResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode updated comment: %w", err)
	}
	return &result, resp, nil
}

// Patch partially updates a comment with the given ID using the request data.
// Returns the patched comment, the raw response, or an error if the request fails.
func (a *CommentsAPI) Patch(
	id int,
	req *PatchCommentRequest,
) (*CommentResponse, *client.Response, error) {
	resp, err := a.Client.Patch(fmt.Sprintf("/comments/%d", id), req)
	if err != nil {
		return nil, nil, fmt.Errorf("PATCH /comments/%d failed: %w", id, err)
	}
	var result CommentResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode patched comment: %w", err)
	}
	return &result, resp, nil
}

// Delete deletes a comment with the given ID.
// Returns the raw response or an error if the request fails.
func (a *CommentsAPI) Delete(id int) (*client.Response, error) {
	resp, err := a.Client.Delete(fmt.Sprintf("/comments/%d", id))
	if err != nil {
		return nil, fmt.Errorf("DELETE /comments/%d failed: %w", id, err)
	}
	return resp, nil
}
