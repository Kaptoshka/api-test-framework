package endpoints

import (
	"fmt"
	"log/slog"
	"time"

	"apitests/internal/client"
	"apitests/pkg/api"
)

// UsersAPI provides methods for interacting with the Users API.
type UsersAPI struct {
	*api.BaseClient
}

// NewUsersAPI creates a new UsersAPI client with the specified base URL, timeout, and logger.
func NewUsersAPI(baseURL string, timeout time.Duration, log *slog.Logger) *UsersAPI {
	return &UsersAPI{
		BaseClient: api.New(baseURL, timeout, "UsersAPI", log),
	}
}

// User represents a user in the API.
type User struct {
	// ID is the unique identifier of the user.
	ID int `json:"id"`
	// Name is the full name of the user.
	Name string `json:"name"`
	// Username is the username of the user.
	Username string `json:"username"`
	// Email is the email address of the user.
	Email string `json:"email"`
}

// CreateUserRequest represents a request to create a new user.
type CreateUserRequest struct {
	// Name is the full name of the new user.
	Name string `json:"name"`
	// Username is the username of the new user.
	Username string `json:"username"`
	// Email is the email address of the new user.
	Email string `json:"email"`
}

// UsersListResponse represents the response for listing users.
type UsersListResponse struct {
	// Data is the list of users.
	Data []*User `json:"data"`
	// Pagination contains pagination information.
	Pagination Pagination `json:"pagination"`
}

// UserResponse represents the response for a single user.
type UserResponse struct {
	// Data is the user data.
	Data *User `json:"data"`
}

// Pagination contains pagination information for list responses.
type Pagination struct {
	// Total is the total number of items.
	Total int `json:"total"`
	// Page is the current page number.
	Page int `json:"page"`
	// Limit is the number of items per page.
	Limit int `json:"limit"`
	// HasNextPage indicates if there is a next page.
	HasNextPage bool `json:"hasNextPage"`
}

// GetAll retrieves all users from the API.
// Returns the list of users, the raw response, or an error if the request fails.
func (a *UsersAPI) GetAll() (*UsersListResponse, *client.Response, error) {
	resp, err := a.Client.Get("/users")
	if err != nil {
		return nil, nil, fmt.Errorf("GET /users failed: %w", err)
	}
	var result UsersListResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode users list: %w", err)
	}
	return &result, resp, nil
}

// GetByID retrieves a user by their ID from the API.
// Returns the user, the raw response, or an error if the request fails.
func (a *UsersAPI) GetByID(id int) (*UserResponse, *client.Response, error) {
	resp, err := a.Client.Get(fmt.Sprintf("/users/%d", id))
	if err != nil {
		return nil, nil, fmt.Errorf("GET /users/%d failed: %w", id, err)
	}
	var result UserResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode user: %w", err)
	}
	return &result, resp, nil
}

// Create creates a new user with the given request data.
// Returns the created user, the raw response, or an error if the request fails.
func (a *UsersAPI) Create(req *CreateUserRequest) (*UserResponse, *client.Response, error) {
	resp, err := a.Client.Post("/users", req)
	if err != nil {
		return nil, nil, fmt.Errorf("POST /users failed: %w", err)
	}
	var result UserResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode created user: %w", err)
	}
	return &result, resp, nil
}

// Update fully updates a user with the given ID using the request data.
// Returns the updated user, the raw response, or an error if the request fails.
func (a *UsersAPI) Update(id int, req *CreateUserRequest) (*UserResponse, *client.Response, error) {
	resp, err := a.Client.Put(fmt.Sprintf("/users/%d", id), req)
	if err != nil {
		return nil, nil, fmt.Errorf("PUT /users/%d failed: %w", id, err)
	}
	var result UserResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode updated user: %w", err)
	}
	return &result, resp, nil
}

// Delete deletes a user with the given ID.
// Returns the raw response or an error if the request fails.
func (a *UsersAPI) Delete(id int) (*client.Response, error) {
	resp, err := a.Client.Delete(fmt.Sprintf("/users/%d", id))
	if err != nil {
		return nil, fmt.Errorf("DELETE /users/%d failed: %w", id, err)
	}
	return resp, nil
}

// Patch partially updates a user with the given ID using the request data.
// Returns the raw response or an error if the request fails.
func (a *UsersAPI) Patch(id int, req *CreateUserRequest) (*client.Response, error) {
	resp, err := a.Client.Patch(fmt.Sprintf("/users/%d", id), req)
	if err != nil {
		return nil, fmt.Errorf("PATCH /users/%d failed: %w", id, err)
	}
	return resp, nil
}

// SearchUsers searches for users matching the given query.
// The search matches against name, username, and email fields.
// Returns the search results, the raw response, or an error if the request fails.
func (a *UsersAPI) SearchUsers(query string) (*UsersListResponse, *client.Response, error) {
	resp, err := a.Client.Get(fmt.Sprintf("/users/search?q=%s", query))
	if err != nil {
		return nil, nil, fmt.Errorf("GET /users/search failed: %w", err)
	}
	var result UsersListResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode user search results: %w", err)
	}
	return &result, resp, nil
}
