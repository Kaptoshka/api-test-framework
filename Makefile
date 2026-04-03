.DEFAULT_GOAL := help

# ─── Help ────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── Dependencies ────────────────────────────────────────────────────────────

.PHONY: deps
deps:
	go mod tidy
	go mod download

# ─── Test ───────────────────────────────────────────────────────────────────

TEST_DIR  = ./tests/...
GO_TEST   = ALLURE_RESULTS_DIR=$(ALLURE_RESULTS_DIR) go test $(TEST_DIR) -v -count=1

.PHONY: test
test: ## Run all tests
	$(GO_TEST)

.PHONY: test-one
test-one: ## Run a specific test: make test-one TEST=TestName
	@test -n "$(TEST)" || (echo "Usage: make test-one TEST=TestName" && exit 1)
	$(GO_TEST) -run $(TEST)

# ─── Allure ──────────────────────────────────────────────────────────────────

ALLURE_RESULTS_DIR=$(PWD)/artifacts/allure-results
ALLURE_REPORTS_DIR=$(PWD)/artifacts/allure-reports
LOGS_DIR=$(PWD)/artifacts/logs

.PHONY: allure-serve
allure-serve: ## Serve Allure report
	allure serve $(ALLURE_RESULTS_DIR)

.PHONY: allure-serve-wsl
allure-serve-wsl: ## Serve Allure report on WSL, use localhost:5050
	allure serve $(ALLURE_RESULTS_DIR) --host 0.0.0.0 --port 5050

.PHONY: allure-generate
allure-generate: ## Generate Allure report
	allure generate $(ALLURE_RESULTS_DIR) -o $(ALLURE_REPORTS_DIR) --clean

# ─── Clean ───────────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Remove generated artifacts
	rm -rf $(ALLURE_RESULTS_DIR) $(ALLURE_REPORTS_DIR) $(LOGS_DIR)
	mkdir -p $(ALLURE_RESULTS_DIR)

# ─── Format ──────────────────────────────────────────────────────────────────

.PHONY: fmt
fmt: fmt-go fmt-nix fmt-yaml ## Run all formatters

.PHONY: fmt-go
fmt-go: ## Format Go files
	goimports -w .

.PHONY: fmt-nix
fmt-nix: ## Format Nix files
	nixpkgs-fmt flake.nix

.PHONY: fmt-yaml
fmt-yaml: ## Format YAML files
	yamlfmt .lint: lint-go lint-yaml lint-nix

# ─── Lint ────────────────────────────────────────────────────────────────────

.PHONY: lint
lint: lint-go lint-nix lint-yaml ## Run all linters

.PHONY: lint-go
lint-go: ## Lint Go files
	golangci-lint run --config .golangci.yml ./...

.PHONY: lint-nix
lint-nix: ## Lint Nix files
	nixpkgs-fmt --check flake.nix
	statix check .
	deadnix --fail .

.PHONY: lint-yaml
lint-yaml: ## Lint YAML files
	yamllint -c .yamllint.yml .

# ─── CI ──────────────────────────────────────────────────────────────────────

.PHONY: ci
ci: lint test ## Run lint + test (for CI/CD)
