package endpoints

import (
	"fmt"
	"log/slog"
	"time"

	"apitests/internal/client"
	apiclient "apitests/pkg/api"
)

// PostsAPI provides methods for interacting with the Posts API.
type PostsAPI struct {
	*apiclient.BaseClient
}

// NewPostsAPI creates a new PostsAPI client with the specified base URL, timeout, and logger.
func NewPostsAPI(baseURL string, timeout time.Duration, log *slog.Logger) *PostsAPI {
	return &PostsAPI{
		BaseClient: apiclient.New(baseURL, timeout, "PostsAPI", log),
	}
}

// Post represents a post in the API.
type Post struct {
	// ID is the unique identifier of the post.
	ID int `json:"id"`
	// Title is the title of the post.
	Title string `json:"title"`
	// Body is the content of the post.
	Body string `json:"body"`
	// UserID is the ID of the user who created the post.
	UserID int `json:"userId"`
}

// CreatePostRequest represents a request to create a new post.
type CreatePostRequest struct {
	// Title is the title of the new post.
	Title string `json:"title"`
	// Body is the content of the new post.
	Body string `json:"body"`
	// UserID is the ID of the user creating the post.
	UserID int `json:"userId"`
}

// PatchPostRequest represents a request to patch an existing post.
type PatchPostRequest struct {
	// Title is the new title for the post (optional).
	Title string `json:"title,omitempty"`
	// Body is the new content for the post (optional).
	Body string `json:"body,omitempty"`
}

// PostsListResponse represents the response for listing posts.
type PostsListResponse struct {
	// Data is the list of posts.
	Data []*Post `json:"data"`
	// Pagination contains pagination information.
	Pagination Pagination `json:"pagination"`
}

// PostResponse represents the response for a single post.
type PostResponse struct {
	// Data is the post data.
	Data *Post `json:"data"`
}

// LikesResponse represents the response for post likes.
type LikesResponse struct {
	// PostID is the ID of the post.
	PostID int `json:"postId"`
	// TotalLikes is the total number of likes on the post.
	TotalLikes int `json:"likes"`
}

// SearchPostsResponse represents the response for searching posts.
type SearchPostsResponse struct {
	// Query is the search query that was used.
	Query string `json:"query"`
	// Total is the total number of matching posts.
	Total int `json:"total"`
	// TotalPages is the total number of pages.
	TotalPages int `json:"totalPages"`
	// Page is the current page number.
	Page int `json:"page"`
	// Limit is the number of results per page.
	Limit int `json:"limit"`
	// HasNext indicates if there is a next page.
	HasNext bool `json:"hasNext"`
	// HasPrev indicates if there is a previous page.
	HasPrev bool `json:"hasPrev"`
	// Results is the list of matching posts.
	Results []*Post `json:"results"`
}

// AddLikeRequest represents a request to add a like to a post.
type AddLikeRequest struct {
	// UserID is the ID of the user adding the like (optional).
	UserID int `json:"userId,omitempty"`
}

// GetAll retrieves all posts from the API.
// Returns the list of posts, the raw response, or an error if the request fails.
func (a *PostsAPI) GetAll() (*PostsListResponse, *client.Response, error) {
	resp, err := a.Client.Get("/posts")
	if err != nil {
		return nil, nil, fmt.Errorf("GET /posts failed: %w", err)
	}
	var result PostsListResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode posts list: %w", err)
	}
	return &result, resp, nil
}

// GetByID retrieves a post by its ID from the API.
// Returns the post, the raw response, or an error if the request fails.
func (a *PostsAPI) GetByID(id int) (*PostResponse, *client.Response, error) {
	resp, err := a.Client.Get(fmt.Sprintf("/posts/%d", id))
	if err != nil {
		return nil, nil, fmt.Errorf("GET /posts/%d failed: %w", id, err)
	}
	var result PostResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode post: %w", err)
	}
	return &result, resp, nil
}

// Create creates a new post with the given request data.
// Returns the created post, the raw response, or an error if the request fails.
func (a *PostsAPI) Create(req *CreatePostRequest) (*PostResponse, *client.Response, error) {
	resp, err := a.Client.Post("/posts", req)
	if err != nil {
		return nil, nil, fmt.Errorf("POST /posts failed: %w", err)
	}
	var result PostResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode created post: %w", err)
	}
	return &result, resp, nil
}

// Update fully updates a post with the given ID using the request data.
// Returns the updated post, the raw response, or an error if the request fails.
func (a *PostsAPI) Update(id int, req *CreatePostRequest) (*PostResponse, *client.Response, error) {
	resp, err := a.Client.Put(fmt.Sprintf("/posts/%d", id), req)
	if err != nil {
		return nil, nil, fmt.Errorf("PUT /posts/%d failed: %w", id, err)
	}
	var result PostResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode updated post: %w", err)
	}
	return &result, resp, nil
}

// Patch partially updates a post with the given ID using the request data.
// Returns the patched post, the raw response, or an error if the request fails.
func (a *PostsAPI) Patch(id int, req *PatchPostRequest) (*PostResponse, *client.Response, error) {
	resp, err := a.Client.Patch(fmt.Sprintf("/posts/%d", id), req)
	if err != nil {
		return nil, nil, fmt.Errorf("PATCH /posts/%d failed: %w", id, err)
	}
	var result PostResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode patched post: %w", err)
	}
	return &result, resp, nil
}

// Delete deletes a post with the given ID.
// Returns the raw response or an error if the request fails.
func (a *PostsAPI) Delete(id int) (*client.Response, error) {
	resp, err := a.Client.Delete(fmt.Sprintf("/posts/%d", id))
	if err != nil {
		return nil, fmt.Errorf("DELETE /posts/%d failed: %w", id, err)
	}
	return resp, nil
}

// GetLikes retrieves the like count for a post with the given ID.
// Returns the like count, the raw response, or an error if the request fails.
func (a *PostsAPI) GetLikes(id int) (*LikesResponse, *client.Response, error) {
	resp, err := a.Client.Get(fmt.Sprintf("/posts/%d/likes", id))
	if err != nil {
		return nil, nil, fmt.Errorf("GET /posts/%d/likes failed: %w", id, err)
	}
	var result LikesResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode likes: %w", err)
	}
	return &result, resp, nil
}

// AddLike adds a like to the post with the given ID.
// Returns the updated like count, the raw response, or an error if the request fails.
func (a *PostsAPI) AddLike(id int, req *AddLikeRequest) (*LikesResponse, *client.Response, error) {
	resp, err := a.Client.Post(fmt.Sprintf("/posts/%d/likes", id), req)
	if err != nil {
		return nil, nil, fmt.Errorf("POST /posts/%d/likes failed: %w", id, err)
	}
	var result LikesResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode add like response: %w", err)
	}
	return &result, resp, nil
}

// Search searches for posts matching the given query.
// Returns the search results, the raw response, or an error if the request fails.
func (a *PostsAPI) Search(query string) (*SearchPostsResponse, *client.Response, error) {
	resp, err := a.Client.Get(fmt.Sprintf("/posts/search?q=%s", query))
	if err != nil {
		return nil, nil, fmt.Errorf("GET /posts/search failed: %w", err)
	}
	var result SearchPostsResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode search results: %w", err)
	}
	return &result, resp, nil
}

// GetWithFilter retrieves posts with advanced filtering and sorting.
// Parameters:
//   - titleLike: filter by title containing this string
//   - sort: field to sort by (e.g., "id", "title")
//   - order: sort order ("asc" or "desc")
//
// Returns the filtered and sorted posts, the raw response, or an error if the request fails.
func (a *PostsAPI) GetWithFilter(titleLike, sort, order string) (*PostsListResponse, *client.Response, error) {
	path := "/posts?"
	if titleLike != "" {
		path += fmt.Sprintf("title_like=%s&", titleLike)
	}
	if sort != "" {
		path += fmt.Sprintf("_sort=%s&", sort)
	}
	if order != "" {
		path += fmt.Sprintf("_order=%s&", order)
	}
	path = path[:len(path)-1]

	resp, err := a.Client.Get(path)
	if err != nil {
		return nil, nil, fmt.Errorf("GET /posts with filter failed: %w", err)
	}
	var result PostsListResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode filtered posts: %w", err)
	}
	return &result, resp, nil
}

// GetWithDelay retrieves posts with an artificial delay.
// The delay parameter specifies milliseconds to wait before responding.
// Returns the posts, the raw response, or an error if the request fails.
func (a *PostsAPI) GetWithDelay(delayMs int) (*PostsListResponse, *client.Response, error) {
	resp, err := a.Client.Get(fmt.Sprintf("/posts?_delay=%d", delayMs))
	if err != nil {
		return nil, nil, fmt.Errorf("GET /posts with delay failed: %w", err)
	}
	var result PostsListResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode delayed response: %w", err)
	}
	return &result, resp, nil
}
