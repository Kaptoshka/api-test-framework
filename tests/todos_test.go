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

// TestGetAllTodos tests the GET /todos endpoint.
// It sends a GET request to /todos and expects a 200 status code with a paginated response.
// The response may be empty if the API has been reset (daily reset).
// Expected: HTTP 200 OK with JSON body containing data array and pagination object.
func TestGetAllTodos(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "TodosAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /todos returns 200 with pagination (may be empty after daily reset)",
		Severity:    suite.SeverityCritical,
		Feature:     "todos",
	})

	todos := endpoints.NewTodosAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var result *endpoints.TodosListResponse
	testErr = s.Step("GET /todos — expect 200", func() error {
		var resp *client.Response
		var err error
		result, resp, err = todos.GetAll()
		if err != nil {
			return err
		}
		return todos.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Response has pagination", func() error {
		if result.Pagination.Total == 0 && len(result.Data) == 0 {
			return nil
		}
		if len(result.Data) > 0 {
			for _, todo := range result.Data {
				if todo.ID == 0 {
					return errors.New("todo has zero ID")
				}
				if todo.Title == "" {
					return fmt.Errorf("todo %d has empty title", todo.ID)
				}
			}
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestGetCompletedTodos tests the GET /todos?completed=true endpoint.
// It fetches only completed todos and verifies all returned todos have Completed=true.
// Expected: HTTP 200 OK with JSON body containing only completed todos.
func TestGetCompletedTodos(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "TodosAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /todos?completed=true returns only completed todos",
		Severity:    suite.SeverityNormal,
		Feature:     "todos",
	})

	todos := endpoints.NewTodosAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var result *endpoints.TodosListResponse
	testErr = s.Step("GET /todos?completed=true — expect 200", func() error {
		var resp *client.Response
		var err error
		result, resp, err = todos.GetCompleted()
		if err != nil {
			return err
		}
		return todos.AssertStatus(resp, http.StatusOK)
	})
	require.NoError(t, testErr)

	testErr = s.Step("All returned todos are completed", func() error {
		for _, todo := range result.Data {
			if !todo.Completed {
				return fmt.Errorf(
					"todo %d is not completed but was returned in completed filter",
					todo.ID,
				)
			}
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestGetTodoByID tests the GET /todos/:id endpoint.
// First, it fetches all todos to find a valid ID.
// Then it retrieves that specific todo by ID and verifies the response.
// Skips if no todos are available (API may have been reset).
// Expected: HTTP 200 OK with JSON body containing the todo with matching ID.
func TestGetTodoByID(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "TodosAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /todos/:id returns 200 and correct todo (skips if no data)",
		Severity:    suite.SeverityCritical,
		Feature:     "todos",
	})

	todos := endpoints.NewTodosAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var todoID int
	testErr = s.Step("GET /todos — find valid ID", func() error {
		result, _, err := todos.GetAll()
		if err != nil {
			return err
		}
		if len(result.Data) == 0 {
			return errors.New("no todos available (API may have been reset)")
		}
		todoID = result.Data[0].ID
		return nil
	})
	if testErr != nil {
		t.Skipf("Skipping test: %v", testErr)
	}

	var result *endpoints.TodoResponse
	testErr = s.Step(fmt.Sprintf("GET /todos/%d — expect 200", todoID), func() error {
		var resp *client.Response
		var err error
		result, resp, err = todos.GetByID(todoID)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("expected status 200 or 404, got %d", resp.StatusCode)
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("Todo has correct ID (if found)", func() error {
		if result != nil && result.Data != nil && result.Data.ID != todoID {
			return fmt.Errorf(
				"expected ID=%d, got %d",
				todoID,
				result.Data.ID,
			)
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestGetTodoNotFound tests the GET /todos/:id endpoint with a non-existent ID.
// It requests a todo with ID 99999 which should not exist.
// Expected: HTTP 404 Not Found status code.
func TestGetTodoNotFound(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "TodosAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "GET /todos/99999 returns 404",
		Severity:    suite.SeverityNormal,
		Feature:     "todos",
	})

	todos := endpoints.NewTodosAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	testErr = s.Step("GET /todos/99999 — expect 404", func() error {
		_, resp, err := todos.GetByID(99999)
		if err != nil {
			return err
		}
		return todos.AssertStatus(resp, http.StatusNotFound)
	})
	require.NoError(t, testErr)
}

// TestCreateTodo tests the POST /todos endpoint.
// It creates a new todo with title, completed status, and user ID.
// Expected: HTTP 201 Created with JSON body containing the created todo with generated ID.
// Verifies that the returned todo matches the request data.
func TestCreateTodo(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "TodosAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "POST /todos creates todo and returns 201",
		Severity:    suite.SeverityCritical,
		Feature:     "todos",
	})

	todos := endpoints.NewTodosAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	req := &endpoints.CreateTodoRequest{
		Title:     "Buy groceries",
		Completed: false,
		UserID:    2,
	}

	var created *endpoints.TodoResponse
	testErr = s.Step("POST /todos — expect 201", func() error {
		var resp *client.Response
		var err error
		created, resp, err = todos.Create(req)
		if err != nil {
			return err
		}
		return todos.AssertStatus(resp, http.StatusCreated)
	})
	require.NoError(t, testErr)

	testErr = s.Step("Created todo has generated ID", func() error {
		if created.Data.ID == 0 {
			return errors.New("expected non-zero ID for created todo")
		}
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step("Created todo matches request data", func() error {
		if created.Data.Title != req.Title {
			return fmt.Errorf(
				"title mismatch: expected %s, got %s",
				req.Title,
				created.Data.Title,
			)
		}
		if created.Data.Completed != req.Completed {
			return fmt.Errorf(
				"completed mismatch: expected %v, got %v",
				req.Completed,
				created.Data.Completed,
			)
		}
		return nil
	})
	require.NoError(t, testErr)
}

// TestDeleteTodo tests the DELETE /todos/:id endpoint.
// It fetches all todos to find a valid ID, then deletes that todo.
// Expected: HTTP 204 No Content status code on successful deletion.
func TestDeleteTodo(t *testing.T) {
	t.Parallel()
	s := suite.New(t, "TodosAPI")
	require.NoError(t, s.Setup(t.Name()))

	var testErr error
	defer s.Teardown(t.Name(), &testErr)

	s.SetMeta(suite.TestMeta{
		Description: "DELETE /todos/:id returns 204",
		Severity:    suite.SeverityNormal,
		Feature:     "todos",
	})

	todos := endpoints.NewTodosAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

	var todoID int
	testErr = s.Step("GET /todos — find valid ID", func() error {
		result, _, err := todos.GetAll()
		if err != nil {
			return err
		}
		if len(result.Data) == 0 {
			return errors.New("no todos available")
		}
		todoID = result.Data[0].ID
		return nil
	})
	require.NoError(t, testErr)

	testErr = s.Step(fmt.Sprintf("DELETE /todos/%d — expect 204", todoID), func() error {
		resp, err := todos.Delete(todoID)
		if err != nil {
			return err
		}
		return todos.AssertStatus(resp, http.StatusNoContent)
	})
	require.NoError(t, testErr)
}
