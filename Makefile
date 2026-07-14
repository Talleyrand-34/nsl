# Makefile for NSL-Graph network specification management tool

# Variables
BINARY_NAME=nsl-graph
MAIN_FILE=main.go
BUILD_DIR=bin
API_PORT=8081
PHP_PORT=8091
DB_FILE=test-dbs/test.db

# Go build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath

.PHONY: help build build-release clean test test-verbose test-coverage lint fmt vet deps-update
.PHONY: run run-server run-cli run-fullstack dev-server
.PHONY: install check
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

check: fmt vet lint yang-validate ## Run all code quality checks

## YANG schema
# Order matters: a module must be listed after the modules it augments, or
# libyang treats the augmented module as merely imported and silently SKIPS the
# "when" checks that gate our augments.
YANG_DIR   := internal/yang/modules
YANG_DATA  := internal/yang/testdata
YANG_MODS  := $(YANG_DIR)/ietf-network.yang $(YANG_DIR)/ietf-network-topology.yang \
              $(YANG_DIR)/ietf-l2-topology.yang $(YANG_DIR)/nsl-topology.yang \
              $(YANG_DIR)/nsl-inventory.yang
YANG_ALL   := $(YANG_MODS) $(YANG_DIR)/ietf-interfaces.yang $(YANG_DIR)/ietf-ip.yang \
              $(YANG_DIR)/ieee802-dot1q-bridge.yang

yang-validate: ## Validate the YANG modules and their instance fixtures (requires yanglint)
	@if ! command -v yanglint > /dev/null; then \
		echo "yanglint not found. Install with: sudo apt install libyang3-tools"; \
		exit 1; \
	fi
	@echo "==> Schema: all modules parse and augments resolve"
	@yanglint -p $(YANG_DIR) $(YANG_ALL)
	@echo "==> Instance: sample-topology.json must PASS"
	@yanglint -t config -p $(YANG_DIR) $(YANG_MODS) $(YANG_DATA)/sample-topology.json
	@echo "==> Instance: invalid-weak-unreviewed.json must FAIL (the 'must' constraint)"
	@if yanglint -t config -p $(YANG_DIR) $(YANG_MODS) \
	      $(YANG_DATA)/invalid-weak-unreviewed.json 2>/dev/null; then \
		echo "FAIL: a weak, unreviewed link was ACCEPTED — the 'must' constraint is not working."; \
		exit 1; \
	else \
		echo "    correctly rejected"; \
	fi
	@echo "YANG validation passed."

# Validate a real export, not just the fixtures. This is the claim that matters:
# not "the modules parse" but "what this tool actually emits conforms to RFC 8345".
# DB is not in the repo (test-dbs/ is gitignored), so this is a separate target and
# skips cleanly when the database is absent.
YANG_DB ?= test-dbs/real.db

yang-validate-export: build ## Export a DB as YANG JSON and validate it (usage: make yang-validate-export YANG_DB=test-dbs/real.db)
	@if [ ! -d "$(YANG_DB)" ]; then \
		echo "skip: database $(YANG_DB) not found (set YANG_DB=...)"; \
		exit 0; \
	fi
	@if ! command -v yanglint > /dev/null; then \
		echo "yanglint not found. Install with: sudo apt install libyang3-tools"; \
		exit 1; \
	fi
	@echo "==> Exporting $(YANG_DB) as RFC 7951 JSON"
	@./bin/nsl-graph export yang -s "$(YANG_DB)" > /tmp/nsl-yang-export.json
	@echo "==> Validating the export against the modules"
	@yanglint -t config -p $(YANG_DIR) $(YANG_MODS) /tmp/nsl-yang-export.json
	@echo "The export conforms to RFC 8345 / RFC 8944 / IEEE 802.1Q."

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
	@if command -v npm > /dev/null; then \
		echo "Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi

install-nmap: ## Install nmap for network scanning
	@echo "Installing nmap..."
	@if command -v apt-get > /dev/null; then \
		echo "Installing nmap via apt-get..."; \
		sudo apt-get update && sudo apt-get install -y nmap; \
	elif command -v yum > /dev/null; then \
		echo "Installing nmap via yum..."; \
		sudo yum install -y nmap; \
	elif command -v dnf > /dev/null; then \
		echo "Installing nmap via dnf..."; \
		sudo dnf install -y nmap; \
	elif command -v brew > /dev/null; then \
		echo "Installing nmap via brew..."; \
		brew install nmap; \
	elif command -v pacman > /dev/null; then \
		echo "Installing nmap via pacman..."; \
		sudo pacman -S nmap; \
	else \
		echo "Unable to detect package manager. Please install nmap manually."; \
		echo "Visit: https://nmap.org/download.html"; \
	fi

check-nmap: ## Check if nmap is installed and working
	@if command -v nmap > /dev/null; then \
		echo "✓ nmap is installed"; \
		echo "Version: $$(nmap --version | head -1)"; \
		echo "Testing scan capability..."; \
		nmap -sn 127.0.0.1 > /dev/null && echo "✓ nmap is working correctly" || echo "✗ nmap test failed"; \
	else \
		echo "✗ nmap is not installed"; \
		echo "Run 'make install-nmap' to install it"; \
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

## Network scanning targets
scan-help: build ## Show scanning help
	./$(BUILD_DIR)/$(BINARY_NAME) scan --help

scan-network: build ## Scan local network (192.168.1.0/24)
	@echo "Scanning local network (this requires nmap to be installed)..."
	./$(BUILD_DIR)/$(BINARY_NAME) scan network 192.168.1.0/24 --review

scan-host: build ## Scan specific host (requires HOST variable)
	@if [ -z "$(HOST)" ]; then \
		echo "Usage: make scan-host HOST=192.168.1.1"; \
		exit 1; \
	fi
	./$(BUILD_DIR)/$(BINARY_NAME) scan host $(HOST) --auto-import

scan-demo: build ## Demo scan with mock data
	@echo "Running network scan demo..."
	@echo "Note: This requires nmap to be installed and a test network"
	./$(BUILD_DIR)/$(BINARY_NAME) scan network 127.0.0.1 --auto-import --timeout 5

scan-import: build ## Import scan results from file (requires FILE variable)
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make scan-import FILE=scan_results.json"; \
		exit 1; \
	fi
	./$(BUILD_DIR)/$(BINARY_NAME) scan import $(FILE) --review

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
example-setup: deps-update install-tools build db-init check-nmap ## Complete setup for new developers
	@echo ""
	@echo "✅ Setup complete! Try these commands:"
	@echo "  make run-fullstack  # Start full application"
	@echo "  make test           # Run tests"
	@echo "  make cli-help       # See CLI options"
	@echo "  make scan-help      # See scanning options"

example-scan-setup: install-nmap check-nmap build ## Setup for network scanning
	@echo ""
	@echo "✅ Scanning setup complete! Try these commands:"
	@echo "  make scan-demo      # Run demo scan"
	@echo "  make scan-network   # Scan local network"
	@echo "  make scan-host HOST=<ip>  # Scan specific host"

example-dev-workflow: fmt vet test build ## Typical development workflow
	@echo "✅ Development workflow complete!"

# Make sure build directory exists for any target that needs it
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)