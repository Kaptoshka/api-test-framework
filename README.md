# 🎭 Playwright Go Testing Framework

An automated GUI testing framework for web applications built with Go and Playwright.

---

## Tech Stack

- **[playwright-go](https://github.com/playwright-community/playwright-go)** — browser automation
- **[Go 1.26](https://go.dev/)** — programming language
- **[Allure](https://allurereport.org/)** — test reporting
- **[testify](https://github.com/stretchr/testify)** — assertions
- **[slog](https://pkg.go.dev/log/slog)** — structured logging
- **[godotenv](https://github.com/joho/godotenv)** — configuration from `.env`

---

## Architecture

```
internal/                    ← Infrastructure Layer
  browser/manager.go         ← browser lifecycle management
  config/config.go           ← configuration from environment variables
  logger/logger.go           ← slog logger initialization
  reporter/allure.go         ← Allure JSON report generation
  screenshot/service.go      ← screenshot capture

pkg/                         ← Framework Core Layer
  elements/element.go        ← playwright.Locator wrapper
  pages/base_page.go         ← base Page Object
  suite/suite.go             ← test lifecycle management

tests/                       ← Test Layer
  pages/                     ← Page Objects for site pages
  components/                ← reusable UI components (header, etc.)
  testdata.go                ← shared constants and test data
  *_test.go                  ← test cases
```

### Layer Responsibilities

| Layer          | Package     | Responsibility                                   |
| -------------- | ----------- | ------------------------------------------------ |
| Infrastructure | `internal/` | Browser, config, logger, Allure, screenshots     |
| Core           | `pkg/`      | BasePage, Element, TestSuite                     |
| Tests          | `tests/`    | Page Objects and test cases — all developer work |

> `internal/` and `pkg/` are the framework itself. All test work happens only in `tests/`.

---

## Quick Start

### Prerequisites

- [Nix](https://nixos.org/) with flakes enabled (for NixOS/WSL)
- or [Go 1.26](https://go.dev/dl/) and [Node.js](https://nodejs.org/) (for Windows/macOS/Linux)

### NixOS / WSL

```bash
# Clone the repository
git clone https://github.com/Kaptoshka/autotests-framework
cd autotests-framework

# Enter the dev environment — Go, Node.js and browsers load automatically
nix develop

# Create configuration
cp .env.example .env
# Edit .env — set BASE_URL and other parameters

# Run tests
make test
```

### Windows / macOS / Linux

```bash
git clone https://github.com/Kaptoshka/autotests-framework
cd autotests-framework

# Install dependencies
go mod download

# Install Playwright browsers
go run github.com/playwright-community/playwright-go/cmd/playwright install chromium

# Create configuration
cp .env.example .env

# Run tests
make test
```

---

## Configuration

All parameters are set via `.env` file or environment variables:

```bash
# Target application
BASE_URL=https://your-app.com

# Browser
BROWSER=chromium          # chromium | firefox | webkit
HEADLESS=true             # true | false
TRACE=false               # true | false
SLOW_MO_MS=0              # ms to slow down actions (0 == off)

# Timeouts
TIMEOUT_MS=30000        # element wait timeout in ms

# Logging
LOG_LEVEL=info          # debug | info | warn | error
LOG_DIR=./artifacts/logs

# Artifacts
ALLURE_RESULTS_DIR=./artifacts/allure-results
ALLURE_REPORTS_DIR=./artifacts/allure-reports
TRACE_DIR=./artifacts/traces
```

> Environment variables take precedence over the `.env` file.

---

## Running Tests

```bash
# All tests
make test

# Run linters and then tests
make ci

# Specific browser
make test-chrome
make test-firefox
make test-webkit

# Headed mode (for debugging)
make test-headed

# Single test by name
TEST_NAME=TestFilterByPrice make test-one

# With Playwright tracing
TRACE=true make test
```

---

## Viewing Reports

```bash
# Open Allure report in browser
allure serve ./artifacts/allure-results --host 0.0.0.0 --port 5050

# Or generate a static HTML report
allure generate ./artifacts/allure-results -o ./artifacts/allure-report --clean

# Or use Make
# Serve Allure report in browser (random port)
make allure-serve

# Serve Allure report in browser (localhost:5050)
make allure-serve-wsl

# Generate static HTML report
make allure-generate
```

On Windows — open `http://localhost:5050` in your browser after running `allure serve`.

---

## Writing a Test Case

```go
func TestExample(t *testing.T) {
    s := suite.New(t, "SuiteName")

    if err := s.Setup(t.Name()); err != nil {
        t.Fatalf("setup failed: %v", err)
    }

    var testErr error
    defer s.Teardown(t.Name(), &testErr)

    // Allure metadata
    s.SetMeta(suite.TestMeta{
        Description: "What this test verifies",
        Severity:    suite.SeverityCritical,
        Feature:     "feature-name",
    })

    // Page Objects
    page := testpages.NewExamplePage(
        s.Browser.Page,
        s.Config.BaseURL,
        s.Config.Timeout,
        s.Log.With("page", "ExamplePage"),
    )

    // Test steps
    testErr = s.Step("Open the page", func() error {
        return page.Navigate("/example")
    })
    require.NoError(t, testErr)
}
```

---

## Creating a Page Object

```go
// tests/pages/example_page.go
package testpages

import (
    "log/slog"
    "time"
    "github.com/playwright-community/playwright-go"
    "github.com/yourorg/playwright-framework/pkg/pages"
)

type ExamplePage struct {
    *pages.BasePage
}

func NewExamplePage(
    page playwright.Page,
    baseURL string,
    timeout time.Duration,
    log *slog.Logger,
) *ExamplePage {
    return &ExamplePage{
        BasePage: pages.NewBasePage(page, baseURL, timeout, "ExamplePage", log),
    }
}

func (p *ExamplePage) ClickSubmit() error {
    return p.Page.Locator("button[type='submit']").Click()
}

func (p *ExamplePage) GetHeading() (string, error) {
    return p.Page.Locator("h1").TextContent()
}
```

---

## Logging

Logs are written to stdout and to `artifacts/logs/session_YYYY-MM-DD_HH-MM-SS.log`.

Every log line includes the test name:

```
2026-03-23T05:20:39 INF  Step started      test=TestLogin page=LoginPage
2026-03-23T05:20:39 DBG  Clicking element  test=TestLogin element="Submit Button"
2026-03-23T05:20:39 WRN  Test FAILED       test=TestLogin
```

Log level is configured via `LOG_LEVEL` in `.env`.

---

## Artifacts

After running tests all artifacts are saved to `artifacts/`:

```
artifacts/
├── allure-results/     ← JSON test results (used by allure serve)
├── allure-report/      ← generated static HTML report
├── logs/               ← session log files
└── traces/             ← Playwright traces (when TRACE=true)
```

> All directories under `artifacts/` are in `.gitignore`.

---

## Browser Support

| Browser | NixOS/WSL    | Windows | macOS | Docker |
| ------- | ------------ | ------- | ----- | ------ |
| Chrome  | ✅ Nix store | ✅      | ✅    | ✅     |
| Firefox | ✅ Nix store | ✅      | ✅    | ✅     |
| WebKit  | ✅ Nix store | ✅      | ✅    | ✅     |

On NixOS browsers are served from the Nix store — no downloading required.

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

## NixOS Environment Variables

When entering `nix develop` the following variables are exported automatically via `shellHook`:

```bash
PLAYWRIGHT_BROWSERS_PATH                # path to browsers in Nix store
PLAYWRIGHT_NODEJS_PATH                  # path to Node.js binary
PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH     # path to Chrome executable
PLAYWRIGHT_FIREFOX_EXECUTABLE_PATH      # path to Firefox executable
PLAYWRIGHT_WEBKIT_EXECUTABLE_PATH       # path to WebKit executable
PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=true
PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
```

---

## License

MIT
