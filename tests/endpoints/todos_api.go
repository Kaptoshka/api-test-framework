package endpoints

import (
	"fmt"
	"log/slog"
	"time"

	"apitests/internal/client"
	apiclient "apitests/pkg/api"
)

// TodosAPI provides methods for interacting with the Todos API.
type TodosAPI struct {
	*apiclient.BaseClient
}

// NewTodosAPI creates a new TodosAPI client with the specified base URL, timeout, and logger.
func NewTodosAPI(baseURL string, timeout time.Duration, log *slog.Logger) *TodosAPI {
	return &TodosAPI{
		BaseClient: apiclient.New(baseURL, timeout, "TodosAPI", log),
	}
}

// Todo represents a todo item in the API.
type Todo struct {
	// ID is the unique identifier of the todo.
	ID int `json:"id"`
	// Title is the title of the todo.
	Title string `json:"title"`
	// Completed indicates whether the todo is completed.
	Completed bool `json:"completed"`
	// UserID is the ID of the user who owns the todo.
	UserID int `json:"userId"`
}

// CreateTodoRequest represents a request to create a new todo.
type CreateTodoRequest struct {
	// Title is the title of the new todo.
	Title string `json:"title"`
	// Completed indicates whether the new todo is completed.
	Completed bool `json:"completed"`
	// UserID is the ID of the user who owns the todo.
	UserID int `json:"userId"`
}

// PatchTodoRequest represents a request to patch an existing todo.
type PatchTodoRequest struct {
	// Title is the new title for the todo (optional).
	Title string `json:"title,omitempty"`
	// Completed indicates whether the todo is completed (optional).
	Completed *bool `json:"completed,omitempty"`
}

// TodosListResponse represents the response for listing todos.
type TodosListResponse struct {
	// Data is the list of todos.
	Data []*Todo `json:"data"`
	// Pagination contains pagination information.
	Pagination Pagination `json:"pagination"`
}

// TodoResponse represents the response for a single todo.
type TodoResponse struct {
	// Data is the todo data.
	Data *Todo `json:"data"`
}

// GetAll retrieves all todos from the API.
// Returns the list of todos, the raw response, or an error if the request fails.
func (a *TodosAPI) GetAll() (*TodosListResponse, *client.Response, error) {
	resp, err := a.Client.Get("/todos")
	if err != nil {
		return nil, nil, fmt.Errorf("GET /todos failed: %w", err)
	}
	var result TodosListResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode todos list: %w", err)
	}
	return &result, resp, nil
}

// GetCompleted retrieves all completed todos from the API.
// Returns the list of completed todos, the raw response, or an error if the request fails.
func (a *TodosAPI) GetCompleted() (*TodosListResponse, *client.Response, error) {
	resp, err := a.Client.Get("/todos?completed=true")
	if err != nil {
		return nil, nil, fmt.Errorf("GET /todos?completed=true failed: %w", err)
	}
	var result TodosListResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode completed todos: %w", err)
	}
	return &result, resp, nil
}

// GetByID retrieves a todo by its ID from the API.
// Returns the todo, the raw response, or an error if the request fails.
func (a *TodosAPI) GetByID(id int) (*TodoResponse, *client.Response, error) {
	resp, err := a.Client.Get(fmt.Sprintf("/todos/%d", id))
	if err != nil {
		return nil, nil, fmt.Errorf("GET /todos/%d failed: %w", id, err)
	}
	var result TodoResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode todo: %w", err)
	}
	return &result, resp, nil
}

// Create creates a new todo with the given request data.
// Returns the created todo, the raw response, or an error if the request fails.
func (a *TodosAPI) Create(req *CreateTodoRequest) (*TodoResponse, *client.Response, error) {
	resp, err := a.Client.Post("/todos", req)
	if err != nil {
		return nil, nil, fmt.Errorf("POST /todos failed: %w", err)
	}
	var result TodoResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode created todo: %w", err)
	}
	return &result, resp, nil
}

// Update fully updates a todo with the given ID using the request data.
// Returns the updated todo, the raw response, or an error if the request fails.
func (a *TodosAPI) Update(id int, req *CreateTodoRequest) (*TodoResponse, *client.Response, error) {
	resp, err := a.Client.Put(fmt.Sprintf("/todos/%d", id), req)
	if err != nil {
		return nil, nil, fmt.Errorf("PUT /todos/%d failed: %w", id, err)
	}
	var result TodoResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode updated todo: %w", err)
	}
	return &result, resp, nil
}

// Patch partially updates a todo with the given ID using the request data.
// Returns the patched todo, the raw response, or an error if the request fails.
func (a *TodosAPI) Patch(id int, req *PatchTodoRequest) (*TodoResponse, *client.Response, error) {
	resp, err := a.Client.Patch(fmt.Sprintf("/todos/%d", id), req)
	if err != nil {
		return nil, nil, fmt.Errorf("PATCH /todos/%d failed: %w", id, err)
	}
	var result TodoResponse
	if err = resp.Decode(&result); err != nil {
		return nil, resp, fmt.Errorf("failed to decode patched todo: %w", err)
	}
	return &result, resp, nil
}

// Delete deletes a todo with the given ID.
// Returns the raw response or an error if the request fails.
func (a *TodosAPI) Delete(id int) (*client.Response, error) {
	resp, err := a.Client.Delete(fmt.Sprintf("/todos/%d", id))
	if err != nil {
		return nil, fmt.Errorf("DELETE /todos/%d failed: %w", id, err)
	}
	return resp, nil
}
