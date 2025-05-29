# Development Guide

This guide covers setting up a development environment and understanding the codebase architecture for NSL-Graph.

## Prerequisites

### Required Software
- **Go 1.24.2+**: Backend API and CLI tool
- **PHP 7.4+**: Web frontend
- **SQLite3**: Database (usually bundled with Go)
- **Git**: Version control

### Optional Tools
- **D2**: For diagram generation (installed automatically via Go modules)
- **Web browser**: For frontend development
- **SQLite browser**: For database inspection

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
go run main.go modify brand --name "TestBrand"

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
├── cmd/                    # CLI command definitions
│   ├── modify/            # Data modification commands
│   ├── print/             # Data display commands
│   ├── export/            # Data export commands
│   └── root/              # Core commands (server, diagram, etc.)
├── internal/              # Internal packages
│   ├── repository/        # Data layer
│   │   ├── application/   # Business logic service layer
│   │   ├── domain/        # Business interfaces
│   │   ├── entities/      # Data structures
│   │   └── infra/         # Infrastructure (SQLite implementation)
│   ├── format/            # Diagram generation
│   └── types/             # Shared data types
├── api/                   # HTTP API handlers
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
- `domain/interfaces.go`: Defines repository contracts
- `infra/sqlc_sqlite/`: SQLite implementation
- `application/app.go`: Service layer consuming repositories

#### Command Pattern
CLI commands are organized using Cobra:
- Each command is a separate file
- Commands are automatically registered
- Consistent flag handling across commands

## Development Workflow

### 1. Database Schema Changes

#### Using SQLC
The project uses SQLC for type-safe database operations:

```bash
# Install sqlc (if not already installed)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate code after schema changes
cd internal/repository/infra/sqlc_sqlite
sqlc generate
```

#### Schema Files
- `schema.sql`: Database schema definition
- `query.sql`: SQL queries
- `sqlc.yml`: SQLC configuration

#### Making Schema Changes
1. Modify `schema.sql`
2. Update `query.sql` if needed
3. Run `sqlc generate`
4. Update entity structs in `entities/`
5. Update service layer in `application/`

### 2. Adding New CLI Commands

#### Create Command File
```go
// cmd/modify/newentity.go
package modify

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
Tests use a separate test database:
```bash
# Test files create their own databases
ls internal/repository/infra/sqlc_sqlite/basicops/test.db
```

### Manual Testing
```bash
# Test CLI commands
go run main.go modify brand --name "TestBrand"
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
```bash
# Open SQLite CLI
sqlite3 test.db

# Common queries
.tables
.schema Device
SELECT * FROM Device;
```

### Log Debugging
The application uses logrus for logging:
```go
import "github.com/sirupsen/logrus"

logrus.WithFields(logrus.Fields{
    "entity": "device",
    "id": deviceId,
}).Info("Processing device")
```

Enable verbose logging:
```bash
go run main.go -v modify device --name "test"
```

### Code Generation
```bash
# Regenerate SQLC code
cd internal/repository/infra/sqlc_sqlite
sqlc generate

# Format Go code
go fmt ./...

# Run linter (if installed)
golangci-lint run
```

## Best Practices

### Code Style
- Follow Go conventions (gofmt, golint)
- Use meaningful variable names
- Write descriptive commit messages
- Add comments for complex logic

### Database Operations
- Always use SQLC-generated code for queries
- Handle errors appropriately
- Use transactions for multi-table operations
- Validate foreign key relationships

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
go run main.go modify brand --name "FirstBrand"
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

### SQLC Issues
```bash
# Verify SQLC installation
sqlc version

# Check configuration
cat internal/repository/infra/sqlc_sqlite/sqlc.yml
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