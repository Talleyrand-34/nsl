# NSL-Graph

Network Specification and Layout Graph - A comprehensive network infrastructure management tool with visualization, VLAN support, and a flexible plugin system.

## Overview

NSL-Graph is a network specification management tool designed to help you document, visualize, and manage complex network infrastructures. It provides three main interfaces:

- **CLI Tool**: Command-line interface for batch operations and automation
- **HTTP API**: RESTful backend for programmatic access
- **Web Frontend**: Browser-based interface for interactive management

### Key Features

- 📊 **Network Visualization**: Generate D2-based diagrams showing devices, connections, and VLANs
- 🔌 **Plugin System**: Extensible architecture for custom connection sorting strategies
- 🌐 **VLAN Support**: Tagged and untagged VLAN configurations with color-coded visualization
- 🗄️ **Clean Architecture**: Separation of concerns with repository pattern and SQLite backend
- 🔄 **Multiple Interfaces**: CLI, HTTP API, and PHP web frontend
- 📝 **Comprehensive Data Model**: Brands, device classes, zones, models, devices, ports, connections, and VLANs

## Quick Start

### Prerequisites

- Go 1.21 or later ([Installation Guide](https://go.dev/doc/install))
- PHP 7.4+ with curl extension (for web frontend)
- D2 diagram tool (optional, for diagram generation)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd nsl5
```

2. Build the binary:
```bash
go build -o nsl-graph main.go
```

3. Run the application:
```bash
# CLI mode
./nsl-graph --help

# Start HTTP API server
./nsl-graph server --port 8081

# Start web frontend
php -S localhost:8091 -t frontend/php
```

4. Or use the convenience script:
```bash
./run-server.sh  # Starts both API (8081) and web UI (8091)
```

## Usage

### CLI Tool

The CLI provides commands for managing all network entities:

```bash
# Add a brand
./nsl-graph modify brand --name "Siemens"

# Add a device class
./nsl-graph modify devclass --name "Switch"

# Add a zone
./nsl-graph modify zone --name "DMZ" --zonetype "Security"

# Add a device model
./nsl-graph modify model --name "XC206" --brand "Siemens" --devclass "Switch"

# Add a device
./nsl-graph modify device --label "Switch-Main" --model "XC206" --zone "DMZ"

# List all devices
./nsl-graph print devices

# Generate a network diagram
./nsl-graph diagram --format connections --vlan true --colorports true
```

### HTTP API

The API exposes RESTful endpoints for all operations:

```bash
# Get all devices
curl http://localhost:8081/devices

# Add a new device
curl -X POST http://localhost:8081/devices \
  -H "Content-Type: application/json" \
  -d '{"label":"Switch-1","model":"XC206","zoneid":"<zone-id>"}'

# Get available plugins
curl http://localhost:8081/plugins

# Change active plugin (no restart required)
curl -X POST http://localhost:8081/plugins/active \
  -H "Content-Type: application/json" \
  -d '{"plugin_id":"zone_name"}'

# Get diagram
curl "http://localhost:8081/diagram?format=connections&vlan=true&colorports=true" > network.svg
```

### Web Frontend

Access the web interface at `http://localhost:8091/main.php`

Features:
- Add, update, delete all network entities
- Visual diagram with resizable viewer
- Real-time plugin switching
- VLAN configuration per port
- Dynamic API endpoint configuration

## Architecture

### Data Model

```
Brand → Device Class → Device Model → Model Ports
                          ↓
Zone Type → Zone → Device → Device Ports → Connections
                                ↓
                              VLANs
```

### Plugin System

NSL-Graph includes a flexible plugin system for customizing connection sorting behavior:

#### Available Built-in Plugins

1. **insertion_order** (default): Maintains database insertion order
2. **zone_name**: Sorts by zone name hierarchy
3. **device_name**: Sorts by device name
4. **reverse_id**: Reverses connection order

#### Configuration

Edit `plugins.yaml` to set the default plugin:

```yaml
plugins:
  connection_sorters:
    active: "zone_name"
```

#### Runtime Plugin Switching

Plugins can be changed at runtime without restarting the server:

**Via API:**
```bash
curl -X POST http://localhost:8081/plugins/active \
  -H "Content-Type: application/json" \
  -d '{"plugin_id":"device_name"}'
```

**Via Web UI:**
Use the plugin selector dropdown in the main interface.

### Directory Structure

```
nsl5/
├── cmd/                    # CLI commands
│   ├── modify/            # Add/update entities
│   ├── print/             # Display data
│   ├── export/            # Data export
│   └── root/              # Root command and diagram
├── api/                   # HTTP API handlers
├── internal/
│   ├── repository/
│   │   ├── application/   # Business logic service layer
│   │   ├── domain/        # Repository interfaces
│   │   ├── entities/      # Core data structures
│   │   ├── infra/         # SQLite implementation
│   │   └── plugins/       # Plugin system
│   └── format/            # Diagram generation
├── frontend/php/          # Web interface
│   ├── action/            # Entity management forms
│   └── config.php         # API endpoint configuration
├── plugins.yaml           # Plugin configuration
└── main.go               # Application entry point
```

## VLAN Support

NSL-Graph provides comprehensive VLAN management:

### Tagged vs Untagged VLANs

- **Untagged VLAN**: Native VLAN for the port (only one allowed per port)
- **Tagged VLAN**: Trunk VLANs (multiple allowed per port)

### Visualization

When VLAN visualization is enabled (`--vlan true` or `vlan=true`):
- Ports are colored by their untagged VLAN
- Connections are colored by the source port's untagged VLAN
- A legend shows VLAN number to color mapping
- Use `--colorports true` to enable port coloring

## Development

### Building from Source

```bash
# Build binary
go build -o nsl-graph main.go

# Run tests
go test ./...

# Run specific package tests
go test ./internal/repository/application
```

### Database Schema Changes

After modifying `internal/repository/infra/sqlc_sqlite/schema.sql`:

```bash
cd internal/repository/infra/sqlc_sqlite
sqlc generate
cd ../../../..
go build -o nsl-graph main.go
```

### Custom Database Location

```bash
# Specify custom database file
./nsl-graph -s /path/to/custom.db server --port 8081
```

## API Endpoints

### Entities
- `GET/POST/PUT/DELETE /brands`
- `GET/POST/PUT/DELETE /deviceclasses`
- `GET/POST/PUT/DELETE /proprietaries`
- `GET/POST/PUT/DELETE /zonetypes`
- `GET/POST/PUT/DELETE /zones`
- `GET/POST/PUT/DELETE /models`
- `GET/POST/PUT/DELETE /devices`
- `GET/POST/PUT/DELETE /modelports`
- `POST /modelports/bulk` - Bulk create model ports
- `GET/POST/PUT/DELETE /deviceports`
- `GET/POST/PUT/DELETE /connections`
- `GET/POST/PUT/DELETE /connectiontypes`
- `GET/POST/PUT/DELETE /vlans`

### Plugins
- `GET /plugins` - List available plugins
- `POST /plugins/active` - Set active plugin

### Diagrams
- `GET /diagram?format={ports|connections}&vlan={true|false}&colorports={true|false}`

## License

Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

## Support

For issues, questions, or contributions, please open an issue on the project repository.

