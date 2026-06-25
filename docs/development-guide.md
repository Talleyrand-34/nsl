# Development Guide

This guide covers setting up a development environment and understanding the codebase architecture for NSL-Graph.

## Prerequisites

### Required Software
- **Go 1.24.2+**: Backend API and CLI tool
- **PHP 7.4+**: Web frontend
- **Git**: Version control

The data store is **CloverDB** (a pure-Go embedded document store, pulled in as a
Go module) — no external database engine to install.

### Optional Tools
- **D2**: For diagram generation (installed automatically via Go modules)
- **Web browser**: For frontend development

## Project Setup

### 1. Clone and Build
```bash
# Clone the repository
git clone <repository-url>
cd nsl-graph

# Install Go dependencies
go mod download

# Build the application
go build -o nsl-graph main.go

# Verify installation
./nsl-graph --help
```

### 2. Development Database
```bash
# Create test database (auto-generated on first use)
go run main.go add brand --name "TestBrand"

# Verify database creation
ls -la test.db
```

### 3. Start Development Services
```bash
# Option 1: Use the startup script
./run-server.sh

# Option 2: Start services manually
go run main.go server --port 8081 &
php -S localhost:8091 -t frontend/php &
```

## Project Architecture

### Directory Structure

```
nsl-graph/
├── cmd/                    # CLI command definitions (Cobra)
│   ├── add/ update/ delete/   # Data modification commands
│   ├── print/             # Data display commands
│   ├── scan/              # Device + connection scanning
│   ├── export/            # Data export commands
│   └── root/              # Core commands (server, diagram, etc.)
├── internal/              # Internal packages
│   ├── api/               # HTTP API handlers
│   ├── repository/        # Data layer
│   │   ├── application/   # Business logic service layer
│   │   ├── domain/        # Business interfaces
│   │   ├── entities/      # Data structures
│   │   └── infra/cloverdb/    # CloverDB document-store implementation
│   ├── scanner/           # SNMP/SSH scanning primitives
│   ├── configparser/      # Device-config parsers (openwrt, opnsense, …)
│   ├── secret/            # Credential vault
│   ├── format/            # Diagram generation (D2/SVG)
│   └── topology/          # Connection-scan correlation types
├── frontend/php/          # PHP web interface
├── docs/                  # Documentation
├── main.go               # Application entry point
└── go.mod                # Go module definition
```

### Architecture Patterns

#### Clean Architecture
The project follows clean architecture principles:

1. **Entities** (`internal/repository/entities/`): Core business entities
2. **Use Cases** (`internal/repository/application/`): Business logic
3. **Interface Adapters** (`api/`, `cmd/`): Controllers and presenters
4. **Frameworks & Drivers** (`internal/repository/infra/`): Database, web framework

#### Repository Pattern
Data access is abstracted through repository interfaces:
- `domain/interfaces.go`: Defines the `NetRepository` contract
- `infra/cloverdb/base/`: CloverDB implementation of that contract
- `application/app.go`: Service layer (`NetService`) consuming the repository

#### Command Pattern
CLI commands are organized using Cobra:
- Each command is a separate file
- Commands are automatically registered
- Consistent flag handling across commands

## Development Workflow

### 1. Database Schema Changes

CloverDB is schemaless (documents are JSON), so there is **no SQL schema or code
generation**. A "schema change" is just a code + docs change:

1. Update the entity struct in `internal/repository/entities/datastruct.go`.
2. Read/write the new field in `internal/repository/infra/cloverdb/base/` (the
   relevant `Add*` / `Update*` / `Get*` methods — set it with `doc.Set(...)`, read
   it with a type-guarded `doc.Get(...)`). Add a focused setter (e.g.
   `UpdateDeviceProfile`) rather than changing a widely-called signature when a
   field is optional.
3. Extend the `NetRepository` (`domain/interfaces.go`) and `NetService`
   (`application/app.go`) interfaces if you added a method.
4. Surface it in the API handlers (`internal/api/`), CLI (`cmd/`), and web UI
   (`frontend/php/`) as needed.
5. Document it in `docs/database-schema.dbml` and the [Glossary](glossary.md).

### 2. Adding New CLI Commands

#### Create Command File
```go
// cmd/add/newentity.go  (mirror in cmd/update/, cmd/delete/, cmd/print/)
package cmd_add

import (
    "github.com/spf13/cobra"
    // ... other imports
)

var newEntityCmd = &cobra.Command{
    Use:   "newentity",
    Short: "Add a new entity",
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation
    },
}

func init() {
    modifyCmd.AddCommand(newEntityCmd)
    newEntityCmd.Flags().StringVar(&entityName, "name", "", "Entity name")
}
```

#### Register Command
Commands are auto-registered through init() functions and import side effects in `main.go`.

### 3. Adding HTTP API Endpoints

#### Add Handler
```go
// api/add.go or api/get.go
func addNewEntityHandler(service q.NetServiceInt) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Implementation
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }
}
```

#### Register Route
```go
// api/server.go
func RegisterRoutes(r *mux.Router, service q.NetServiceInt) {
    // ... existing routes
    r.HandleFunc("/newentities", addNewEntityHandler(service)).Methods("POST")
    r.HandleFunc("/newentities", getNewEntitiesHandler(service)).Methods("GET")
}
```

### 4. Adding PHP Frontend Actions

#### Create Action Files
```php
<!-- frontend/php/action/getnewentity.php -->
<?php
$response = file_get_contents(NEWENTITIES_ENDPOINT);
$entities = json_decode($response, true);

if ($entities) {
    echo "<h3>New Entities:</h3>";
    foreach ($entities as $entity) {
        echo "<p>{$entity['name']}</p>";
    }
} else {
    echo "<p>No entities found.</p>";
}
?>
```

```php
<!-- frontend/php/action/addnewentity.php -->
<?php
if ($_POST['name']) {
    $data = json_encode(['name' => $_POST['name']]);
    $context = stream_context_create([
        'http' => [
            'method' => 'POST',
            'header' => 'Content-Type: application/json',
            'content' => $data
        ]
    ]);
    
    $result = file_get_contents(NEWENTITIES_ENDPOINT, false, $context);
    echo $result ? "Added successfully!" : "Error adding entity.";
}
?>

<form method="post">
    <label>Name: <input type="text" name="name" required></label>
    <button type="submit">Add New Entity</button>
</form>
```

#### Update Configuration
```php
// frontend/php/config.php
define('NEWENTITIES_ENDPOINT', API_BASE_URL . '/newentities');
```

#### Update Main Interface
```php
// frontend/php/main.php
$actionFiles = [
    // ... existing actions
    'getnewentity' => 'action/getnewentity.php',
    'addnewentity' => 'action/addnewentity.php',
];
```

## Testing

### Go Tests
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/repository/application

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestSpecificFunction ./internal/repository/application
```

### Test Database
Tests create their own throwaway CloverDB store (via `t.TempDir()`), e.g. the
`setupTestScanningService` helper in
`internal/repository/application/scanning_test.go`. Nothing to set up by hand.

### Manual Testing
```bash
# Test CLI commands
go run main.go add brand --name "TestBrand"
go run main.go print brand

# Test API endpoints
curl http://localhost:8081/brands

# Test PHP frontend
open http://localhost:8091
```

## Common Development Tasks

### Building for Production
```bash
# Build for current platform
go build -o nsl-graph main.go

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o nsl-graph-linux-amd64 main.go
GOOS=windows GOARCH=amd64 go build -o nsl-graph-windows-amd64.exe main.go
```

### Database Debugging
CloverDB stores its data as files in the database directory (default `test.db/`),
not a SQL database — inspect data through the API or CLI instead:
```bash
# Inspect via the running API
curl http://localhost:8081/devices | jq

# Or via the CLI
go run main.go print device
```

### Log Debugging
The application logs with the standard library's structured logger (`log/slog`):
```go
import "log/slog"

slog.Info("processing device", "entity", "device", "id", deviceId)
```

Enable verbose / debug logging:
```bash
go run main.go -v add device --label "test"   # verbose
go run main.go -d server --port 8081           # debug
```

### Code Generation
There is no code generation (no SQLC). Just format and lint:
```bash
go fmt ./...
go vet ./...
golangci-lint run   # if installed
```

## Best Practices

### Code Style
- Follow Go conventions (gofmt, golint)
- Use meaningful variable names
- Write descriptive commit messages
- Add comments for complex logic

### Database Operations
- Go through the `NetRepository` interface / `NetService`, not CloverDB directly
- Handle errors appropriately
- Enforce referential integrity at the application layer (CloverDB has no FKs)
- Validate input before storing

### API Design
- Return consistent JSON formats
- Use appropriate HTTP status codes
- Validate input data
- Handle errors gracefully

### CLI Commands
- Provide helpful flag descriptions
- Validate required parameters
- Use consistent flag naming
- Provide usage examples

## Troubleshooting

### Build Issues
```bash
# Clear module cache
go clean -modcache

# Rebuild dependencies
go mod download
go mod tidy
```

### Database Issues
```bash
# Reset database
rm test.db
go run main.go add brand --name "FirstBrand"
```

### Port Conflicts
```bash
# Check port usage
lsof -i :8081
lsof -i :8091

# Kill processes if needed
pkill -f "go run main.go server"
pkill -f "php -S"
```

## Contributing

1. Create feature branch
2. Make changes following patterns
3. Add tests for new functionality
4. Update documentation
5. Submit pull request

### Pull Request Checklist
- [ ] Tests pass
- [ ] Code formatted
- [ ] Documentation updated
- [ ] API endpoints tested
- [ ] CLI commands tested
- [ ] PHP frontend tested