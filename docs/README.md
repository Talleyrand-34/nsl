# NSL-Graph Documentation

NSL-Graph is a network specification tool that manages network infrastructure data through a SQLite database, provides a web interface via PHP frontend, and offers comprehensive CLI commands for network management.

## Architecture Overview

The tool consists of three main components:

### 1. Internal Database Layer (`internal/`)
- **Purpose**: SQLite-based data persistence layer for network specifications
- **Location**: `internal/repository/`
- **Components**:
  - **Entities**: Data structures representing network components (brands, devices, connections, etc.)
  - **Domain**: Business logic interfaces and contracts
  - **Infrastructure**: SQLite implementation with SQLC-generated queries
  - **Application**: Service layer providing business operations

### 2. PHP Frontend (`frontend/php/`)
- **Purpose**: Web interface for managing network data
- **Location**: `frontend/php/`
- **Components**:
  - **Actions**: PHP scripts for CRUD operations on network entities
  - **Main Interface**: Web UI for interacting with the network database
  - **Configuration**: Database connection and settings

### 3. CLI Commands (`cmd/`)
- **Purpose**: Command-line interface for network management
- **Location**: `cmd/`
- **Components**:
  - **Modify**: Commands for adding/updating network entities
  - **Print**: Commands for displaying network data
  - **Export**: Data export functionality
  - **Server**: HTTP API server
  - **Diagram**: Network visualization generation

## Quick Start

### Prerequisites
- Go 1.24.2 or later
- PHP (for web frontend)

### Installation
```bash
# Clone the repository
git clone <repository-url>
cd nsl-graph

# Build the application
go build -o nsl-graph main.go
```

### Basic Usage
```bash
# Run CLI tool directly
go run main.go --help

# Or use the built binary
./nsl-graph --help

# Start the full stack (API + Web Frontend)
./run-server.sh
```

## Documentation Structure

- **[CLI Reference](cli-reference.md)**: Complete command-line interface documentation
- **[Database Schema](database-schema.md)**: SQLite database structure and relationships
- **[Frontend API](frontend-api.md)**: PHP web interface and HTTP API endpoints
- **[Development Guide](development-guide.md)**: Setup and development workflow

## Core Concepts

### Network Entities
The tool manages the following network entities:
- **Brands**: Equipment manufacturers (Cisco, Juniper, etc.)
- **Device Classes**: Types of network equipment (router, switch, firewall)
- **Models**: Specific equipment models from brands
- **Zones**: Logical or physical network segments
- **Devices**: Individual network equipment instances
- **Ports**: Connection points on devices
- **Connections**: Links between device ports
- **Proprietaries**: Ownership/management entities

### Data Flow
1. **CLI Commands** → SQLite Database
2. **PHP Frontend** → HTTP API → SQLite Database
3. **Diagram Generation** → Reads from SQLite → Generates D2/SVG diagrams

## Getting Started

1. **[Set up your development environment](development-guide.md)**
2. **[Understand the database schema](database-schema.md)**
3. **[Learn the CLI commands](cli-reference.md)**
4. **[Explore the web frontend](frontend-api.md)**

## License

This project is licensed under the GNU Affero General Public License v3.0.