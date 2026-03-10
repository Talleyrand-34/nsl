# Makefile for NSL-Graph network specification management tool

# Variables
BINARY_NAME=nsl-graph
MAIN_FILE=main.go
BUILD_DIR=bin
API_PORT=8081
PHP_PORT=8091
DB_FILE=test.db
DEFAULT_PLUGIN_CONFIG=plugins.yaml

# Go build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath

.PHONY: help build build-release clean test test-verbose test-coverage lint fmt vet deps-update
.PHONY: run run-server run-cli run-fullstack dev-server
.PHONY: generate db-generate install check
.PHONY: docker-build docker-run
.PHONY: docs serve-docs

# Default target
all: clean build test

## Help
help: ## Show this help message
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Build targets
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

build-release: ## Build optimized release binary
	@echo "Building release version of $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

## Development targets
dev: build ## Build and run in development mode (same as build)

run: build ## Build and run CLI (interactive)
	./$(BUILD_DIR)/$(BINARY_NAME)

run-cli: ## Run CLI directly with go run
	go run $(MAIN_FILE)

run-server: ## Run API server only
	@echo "Starting API server on port $(API_PORT)..."
	go run $(MAIN_FILE) server --port $(API_PORT)

run-php: ## Run PHP frontend only
	@echo "Starting PHP frontend on port $(PHP_PORT)..."
	php -S localhost:$(PHP_PORT) -t frontend/php

run-fullstack: ## Run both API server and PHP frontend
	@echo "Starting full stack (API + Frontend)..."
	@echo "API server: http://localhost:$(API_PORT)"
	@echo "PHP frontend: http://localhost:$(PHP_PORT)"
	@echo "Press Ctrl+C to stop both servers"
	bash scripts/run-server.sh

dev-server: ## Development server with auto-restart (requires entr or air)
	@if command -v air > /dev/null; then \
		echo "Using air for hot reload..."; \
		air; \
	elif command -v entr > /dev/null; then \
		echo "Using entr for file watching..."; \
		find . -name "*.go" | entr -r go run $(MAIN_FILE) server --port $(API_PORT); \
	else \
		echo "Install 'air' or 'entr' for auto-restart, falling back to regular server..."; \
		$(MAKE) run-server; \
	fi

## Testing targets
test: ## Run all tests
	@echo "Running tests..."
	go test ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	go test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report saved to $(BUILD_DIR)/coverage.html"

test-specific: ## Run specific test package (usage: make test-specific PKG=./internal/repository/application)
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-specific PKG=./path/to/package"; \
		exit 1; \
	fi
	go test -v $(PKG)

## Code quality targets
lint: ## Run linter (requires golangci-lint)
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

check: fmt vet lint ## Run all code quality checks

## Database and generation targets
generate: ## Generate code from schema (SQLC)
	@echo "Generating SQLC code..."
	@cd internal/repository/infra/sqlc_sqlite && sqlc generate

db-generate: generate ## Alias for generate

## Dependency management
deps-update: ## Update Go dependencies
	@echo "Updating dependencies..."
	go mod tidy
	go mod download

deps-vendor: ## Create vendor directory
	go mod vendor

## Installation targets
install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	go install $(MAIN_FILE)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@if command -v npm > /dev/null; then \
		echo "Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi

## Cleaning targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	rm -f main
	rm -rf .artifacts/*
	go clean -cache
	go clean -testcache

clean-all: clean ## Clean everything including dependencies
	rm -rf vendor/
	go clean -modcache

## Docker targets (optional)
docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):latest .

docker-run: ## Run Docker container
	docker run -p $(API_PORT):$(API_PORT) $(BINARY_NAME):latest

## Documentation targets
docs: ## Open documentation in browser
	@echo "Opening documentation..."
	@if command -v xdg-open > /dev/null; then \
		xdg-open docs/README.md; \
	elif command -v open > /dev/null; then \
		open docs/README.md; \
	else \
		echo "Please open docs/README.md manually"; \
	fi

serve-docs: ## Serve documentation with live reload (requires Python)
	@if command -v python3 > /dev/null; then \
		echo "Serving docs at http://localhost:8000/docs/"; \
		cd docs && python3 -m http.server 8000; \
	else \
		echo "Python3 not found. Install Python3 to serve docs."; \
	fi

## CLI command helpers
cli-help: build ## Show CLI help
	./$(BUILD_DIR)/$(BINARY_NAME) --help

cli-brands: build ## List all brands
	./$(BUILD_DIR)/$(BINARY_NAME) print brands

cli-devices: build ## List all devices
	./$(BUILD_DIR)/$(BINARY_NAME) print devices

cli-connections: build ## List all connections
	./$(BUILD_DIR)/$(BINARY_NAME) print connections

cli-diagram: build ## Generate connection diagram
	./$(BUILD_DIR)/$(BINARY_NAME) diagram connections

cli-vlan-diagram: build ## Generate VLAN diagram
	./$(BUILD_DIR)/$(BINARY_NAME) diagram connections --vlan true

## Database management
db-init: build ## Initialize database with sample data
	@if [ -f init.sh ]; then \
		echo "Running database initialization..."; \
		bash init.sh; \
	else \
		echo "No init.sh found. Creating empty database..."; \
		./$(BUILD_DIR)/$(BINARY_NAME) print brands > /dev/null 2>&1 || echo "Database created"; \
	fi

db-backup: ## Backup current database
	@if [ -f $(DB_FILE) ]; then \
		cp $(DB_FILE) $(DB_FILE).backup.$$(date +%Y%m%d_%H%M%S); \
		echo "Database backed up to $(DB_FILE).backup.$$(date +%Y%m%d_%H%M%S)"; \
	else \
		echo "No database file found at $(DB_FILE)"; \
	fi

db-reset: ## Reset database (WARNING: deletes all data)
	@echo "WARNING: This will delete all data in $(DB_FILE)"
	@echo "Press Ctrl+C to cancel, Enter to continue..."
	@read
	rm -f $(DB_FILE)
	@echo "Database reset. Run 'make db-init' to reinitialize."

## Plugin management
plugins-list: build ## List available plugins
	./$(BUILD_DIR)/$(BINARY_NAME) print plugins || echo "Plugin listing not implemented in CLI"

plugins-config: ## Show current plugin configuration
	@if [ -f $(DEFAULT_PLUGIN_CONFIG) ]; then \
		cat $(DEFAULT_PLUGIN_CONFIG); \
	else \
		echo "No plugin configuration found at $(DEFAULT_PLUGIN_CONFIG)"; \
	fi

## Release and deployment
version: ## Show version information
	@echo "NSL-Graph Network Management Tool"
	@echo "Go version: $$(go version)"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build date: $$(date)"

release: clean test build-release ## Create a release build
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

## Example workflows
example-setup: deps-update install-tools generate build db-init ## Complete setup for new developers
	@echo ""
	@echo "✅ Setup complete! Try these commands:"
	@echo "  make run-fullstack  # Start full application"
	@echo "  make test           # Run tests"
	@echo "  make cli-help       # See CLI options"

example-dev-workflow: fmt vet test build ## Typical development workflow
	@echo "✅ Development workflow complete!"

# Make sure build directory exists for any target that needs it
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)