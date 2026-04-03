<h1 align="center"> API Mock Testing Framework</h1>

<div align="center">

![go](https://img.shields.io/static/v1?style=flat&label=Go&message=1.26&colorA=24273A&colorB=91d7e3&logo=go)
![license](https://img.shields.io/static/v1?style=flat&label=License&message=MIT&colorA=24273A&colorB=91d7e3&logo=gitbook&logoColor=91d7e3)

</div>

An automated API testing framework for REST APIs built with Go.

## Table of Contents

- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
    - [Layer Responsibilities](#layer-responsibilities)
- [Quick Start](#quick-start)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
- [Configuration](#configuration)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
    - [Basic Test](#basic-test)
    - [Test with Steps](#test-with-steps)
- [API Endpoints](#api-endpoints)
- [Advanced Testing](#advanced-testing)
    - [Filtering & Search](#filtering--search)
    - [Response Delay](#response-delay)
- [Viewing Reports](#viewing-reports)
- [Logging](#logging)
- [CI/CD](#cicd)

---

## Tech Stack

- **[Go 1.26](https://go.dev/)** — programming language
- **[Allure](https://allurereport.org/)** — test reporting
- **[testify](https://github.com/stretchr/testify)** — assertions
- **[slog](https://pkg.go.dev/log/slog)** — structured logging
- **[godotenv](https://github.com/joho/godotenv)** — configuration from `.env`

---

## Architecture

```
internal/                   ← Infrastructure Layer
  client/http.go            ← HTTP client for API requests
  config/config.go          ← configuration from environment variables
  logger/logger.go          ← slog logger initialization
  reporter/allure.go        ← Allure JSON report generation

pkg/                        ← Framework Core Layer
  api/base_client.go        ← base API client with assertions
  suite/suite.go            ← test lifecycle management

tests/                      ← Test Layer
  endpoints/                ← API endpoint clients (Posts, Users, Todos, Comments)
  *_test.go                 ← test cases
  advanced_test.go          ← advanced test cases (filtering, search, delay)
```

### Layer Responsibilities

| Layer          | Package     | Responsibility                                  |
| -------------- | ----------- | ----------------------------------------------- |
| Infrastructure | `internal/` | HTTP client, config, logger, Allure reporter    |
| Core           | `pkg/`      | BaseClient, TestSuite                           |
| Tests          | `tests/`    | API clients and test cases — all developer work |

> `internal/` and `pkg/` are the framework itself. All test work happens only in `tests/`.

---

## Quick Start

### Prerequisites

- [Go 1.26](https://go.dev/dl/)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourorg/api-mocker
cd api-mocker

# Install dependencies
go mod download

# Create configuration
cp .env.example .env
# Edit .env — set BASE_URL and other parameters

# Run tests
make test
```

---

## Configuration

All parameters are set via `.env` file or environment variables:

```bash
# Target API base URL
BASE_URL=https://jsonplaceholder.typicode.com

# Request timeout in milliseconds
TIMEOUT_MS=30000

# Allure results directory
ALLURE_RESULTS_DIR=./allure-results

# Logging
LOG_LEVEL=info          # debug | info | warn | error
LOG_DIR=./logs
```

> Environment variables take precedence over the `.env` file.

---

## Running Tests

```bash
# All tests
go test ./tests/... -v

# Run with Allure reporting
go test ./tests/... -v && allure serve ./allure-results

# Run specific test file
go test ./tests/posts_test.go -v

# Run tests matching pattern
go test ./tests/... -run "TestGet" -v

# Run with verbose output
go test ./tests/... -v -timeout 60s
```

---

## Writing Tests

### Basic Test

```go
func TestGetAllPosts(t *testing.T) {
    t.Parallel()
    s := suite.New(t, "PostsAPI")
    require.NoError(t, s.Setup(t.Name()))

    var testErr error
    defer s.Teardown(t.Name(), &testErr)

    s.SetMeta(suite.TestMeta{
        Description: "GET /posts returns 200 with pagination",
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
}
```

### Test with Steps

```go
func TestCreatePost(t *testing.T) {
    t.Parallel()
    s := suite.New(t, "PostsAPI")
    require.NoError(t, s.Setup(t.Name()))

    var testErr error
    defer s.Teardown(t.Name(), &testErr)

    s.SetMeta(suite.TestMeta{
        Description: "POST /posts creates post and returns 201",
        Severity:    suite.SeverityCritical,
        Feature:     "posts",
    })

    posts := endpoints.NewPostsAPI(s.Config.BaseURL, s.Config.Timeout, s.Log)

    req := &endpoints.CreatePostRequest{
        Title:  "Test Post Title",
        Body:   "Test post body content",
        UserID: 1,
    }

    var created *endpoints.PostResponse
    testErr = s.Step("POST /posts — expect 201", func() error {
        var resp *client.Response
        var err error
        created, resp, err = posts.Create(req)
        if err != nil {
            return err
        }
        return posts.AssertStatus(resp, http.StatusCreated)
    })
    require.NoError(t, testErr)

    testErr = s.Step("Created post matches request data", func() error {
        if created.Data.Title != req.Title {
            return fmt.Errorf("title mismatch: expected %s, got %s", req.Title, created.Data.Title)
        }
        return nil
    })
    require.NoError(t, testErr)
}
```

---

## API Endpoints

The framework provides ready-to-use API clients for:

| API      | Client        | File                              |
| -------- | ------------- | --------------------------------- |
| Posts    | `PostsAPI`    | `tests/endpoints/posts_api.go`    |
| Users    | `UsersAPI`    | `tests/endpoints/users_api.go`    |
| Todos    | `TodosAPI`    | `tests/endpoints/todos_api.go`    |
| Comments | `CommentsAPI` | `tests/endpoints/comments_api.go` |

### Available Methods

Each API client provides standard CRUD methods:

- `GetAll()` — GET all resources
- `GetByID(id)` — GET single resource by ID
- `Create(req)` — POST create new resource
- `Update(id, req)` — PUT full update
- `Patch(id, req)` — PATCH partial update
- `Delete(id)` — DELETE resource

---

## Advanced Testing

### Filtering & Search

```go
// Filter posts by title and sort
result, _, err := posts.GetWithFilter("web", "id", "desc")

// Search posts by content
result, _, err := posts.Search("development")

// Search users by name, username, or email
result, _, err := users.SearchUsers("john")
```

### Response Delay

Test frontend loading states with artificial delays:

```go
// Request with 2-second delay
result, _, err := posts.GetWithDelay(2000)
```

---

## Viewing Reports

```bash
# Open Allure report in browser
allure serve ./allure-results --host 0.0.0.0 --port 5050

# Or generate a static HTML report
allure generate ./allure-results -o ./allure-report --clean
```

---

## Logging

Logs are written to stdout and to `logs/session_YYYY-MM-DD_HH-MM-SS.log`.

Every log line includes test name and API client:

```
2026-04-03T16:09:00 INF Test setup complete test=TestPostSearch test=TestPostSearch
2026-04-03T16:09:00 INF GET /posts/search?q=development — expect 200 test=TestPostSearch
2026-04-03T16:09:00 INF HTTP request test=TestPostSearch method=GET url="https://apimocker.com/posts/search?q=development"
```

Log level is configured via `LOG_LEVEL` in `.env`.

---

## CI/CD

### GitHub Actions

GitHub Actions workflows are located in `.github/workflows/`:

- `lint.yml` — runs Go, Nix and YAML linters on every push and pull request
- `test.yml` — runs the full test suite and uploads Allure results as artifacts

---

## Linters

```bash
# Go
golangci-lint run ./...

# Nix
statix check .
deadnix --fail .
nixpkgs-fmt --check .

# YAML
yamllint .

# All linters at once
make lint
```

---

## License

[MIT](LICENSE)
