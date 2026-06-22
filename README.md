# NSL-Graph

Network Specification and Layout Graph - A comprehensive network infrastructure management tool with visualization, VLAN support, and multi-vendor device configuration integration.

## Overview

NSL-Graph is a network specification management tool designed to help you document, visualize, and manage complex network infrastructures. It provides three main interfaces:

- **CLI Tool**: Command-line interface for batch operations and automation
- **HTTP API**: RESTful backend for programmatic access
- **Web Frontend**: Browser-based interface for interactive management

### Key Features

- **Network Visualization**: Generate D2-based diagrams showing devices, connections, and VLANs
- **VLAN Support**: Tagged and untagged VLAN configurations with color-coded visualization
- **WiFi Support**: WiFi radio ports and SSID interface entities
- **Device Configuration Integration**: SSH, file, and manual config parsing for OPNsense, OpenWrt, Fortinet, and Cisco devices
- **SNMP + Config Coordination**: Merge operational SNMP data with declarative configuration
- **Clean Architecture**: Repository pattern with CloverDB backend
- **Multiple Interfaces**: CLI, HTTP API, and PHP web frontend

## Quick Start

### Prerequisites

- Go 1.21 or later ([Installation Guide](https://go.dev/doc/install))
- PHP 7.4+ with curl extension (for web frontend)
- D2 diagram tool (optional, for diagram generation)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd nsl
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
./nsl-graph add brand --name "Siemens"

# Add a device class
./nsl-graph add devclass --name "Switch"

# Add a zone
./nsl-graph add zone --name "DMZ" --zonetype "Security"

# Add a device model
./nsl-graph add model --name "XC206" --brand "Siemens" --devclass "Switch"

# Add a device
./nsl-graph add device --label "Switch-Main" --model "XC206" --zone "DMZ"

# List all devices
./nsl-graph print devices

# Generate a network diagram (add --vlan for VLAN colors)
./nsl-graph diagram connection --vlan
```

### Device Scanning with Configuration Integration

```bash
# Basic SNMP scanning
./nsl-graph scan host 192.168.1.1

# Enhanced scanning with SSH configuration retrieval
./nsl-graph scan host 10.0.0.1 --config-source ssh --ssh-user admin --merge-configs

# Manual device type specification (bypass auto-detection)
./nsl-graph scan host 10.0.0.245 --config-source ssh --ssh-user root --device-type openwrt

# Configuration file parsing
./nsl-graph scan host 10.0.0.132 --config-source file --config-file fortinet.conf --auto-import

# Advanced options with conflict resolution
./nsl-graph scan host 10.0.0.1 --config-source ssh --ssh-user admin --device-type opnsense --discrepancy-action prefer-config --auto-import
```

### Scan Profiles

Store the scan parameters for a host once, then reuse them. A profile is
**auto-applied** when you scan a host whose IP matches it; explicit flags always
override.

```bash
# Save a profile (SNMP only — no passphrase needed)
./nsl-graph scan profile add ow --host 10.0.2.245 --snmp-community public --device-type openwrt

# Save one with an SSH password — you are prompted for a passphrase that encrypts it
./nsl-graph scan profile add fw --host 10.0.0.1 --ssh-user admin --ssh-password '***' \
  --scan-source ssh --device-type opnsense

# List / show (the SSH password is never printed) / delete
./nsl-graph scan profile list
./nsl-graph scan profile show fw
./nsl-graph scan profile delete ow

# Scan — the matching profile is applied automatically
./nsl-graph scan host 10.0.2.245
# Or save the current flags as a profile while scanning
./nsl-graph scan host 10.0.2.245 --snmp-community public --save-profile ow
```

**Security:** the SSH password is the only secret; it is encrypted at rest with
**AES-256-GCM** using a key derived (scrypt) from a **passphrase you supply** —
there is **no environment variable** and it is never stored in clear. The
passphrase is asked when you save the password and again whenever an SSH scan
uses it; a wrong passphrase fails. SNMP-only profiles need no passphrase. The web
UI (Import page) and the API (`/scan/profiles`) manage SNMP profiles too.

### HTTP API

The API exposes RESTful endpoints for all operations:

```bash
# Get all devices
curl http://localhost:8081/devices

# Add a new device
curl -X POST http://localhost:8081/devices \
  -H "Content-Type: application/json" \
  -d '{"label":"Switch-1","model":"XC206","zoneid":"<zone-id>"}'

# Get diagram
curl "http://localhost:8081/diagram?format=connections&vlan=true&colorports=true" > network.svg
```

### Web Frontend

Access the web interface at `http://localhost:8091/main.php`

Features:
- Add, update, delete all network entities
- Visual diagram with resizable viewer
- VLAN configuration per port
- Dynamic API endpoint configuration
- **Device import** (`import.php`):
  - **SNMP** scan (host or subnet) and **SSH** scan (via a saved profile + passphrase).
  - A per-device **report with an editable VLAN table** — adjust the
    **vlan-id / ip-segment** for each interface IP before importing (no edits = accept
    the suggested mapping).
  - Upload a scan-result JSON file; save/load/delete scan profiles.

## Architecture

### Data Model

```
Brand → Device Class → Device Model → Model Ports
                          ↓
Zone Type → Zone → Device → Device Ports → Connections
                       ↓          ↓
               Device Interfaces  VLANs
                       ↓
               WiFi Radio Ports → SSID Interfaces
```

### Directory Structure

```
nsl/
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
│   │   └── infra/        # CloverDB implementation
│   ├── configparser/      # Configuration parsing framework
│   │   └── parsers/       # Device-specific parsers (OPNsense, OpenWrt, Fortinet, Cisco)
│   └── format/            # Diagram generation
├── frontend/php/          # Web interface
│   ├── action/            # Entity management forms
│   └── config.php         # API endpoint configuration
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

## Device Configuration Integration

### Supported Device Types

| Device Type | Format | Auto-detection |
|-------------|--------|---------------|
| **OPNsense** | XML (`/conf/config.xml`) | FreeBSD + "opnsense" in SysDescr/SysName |
| **OpenWrt** | UCI (Unified Configuration Interface) | Linux + "openwrt" in SysDescr/SysName |
| **Fortinet FortiGate** | CLI configuration | "fortinet", "fortigate", or "FG-" prefix |
| **Cisco** | CLI configuration | "cisco" or "IOS" in SysDescr |

### Configuration Sources

- `--config-source ssh`: Connect via SSH to retrieve configuration
- `--config-source file`: Read from a local file (`--config-file /path/to/config`)
- `--config-source manual`: SNMP only, no automatic config retrieval
- `--config-source none`: Default, pure SNMP-based scanning

### Conflict Resolution

When SNMP and configuration data conflict:

- `--discrepancy-action prefer-snmp` (default): Trust SNMP operational data
- `--discrepancy-action prefer-config`: Trust device configuration
- `--discrepancy-action fail`: Stop on any conflict

### Configuration Flags Reference

| Flag | Description |
|------|-------------|
| `--config-source` | Configuration source (none, ssh, file, manual) |
| `--config-file` | Path to configuration file (file source only) |
| `--device-type` | Manual type override (opnsense, openwrt, fortinet, cisco) |
| `--ssh-user` | SSH username |
| `--ssh-password` | SSH password |
| `--ssh-key` | Path to SSH private key |
| `--ssh-port` | SSH port (default: 22) |
| `--discrepancy-action` | Conflict resolution strategy |
| `--merge-configs` | Merge SNMP and configuration data |
| `--auto-import` | Automatically import parsed data |

## Development

### Building from Source

```bash
# Build binary
go build -o nsl-graph main.go

# Run tests
go test ./...

# Run specific package tests
go test ./internal/repository/application
go test ./internal/configparser
```

### Custom Database Location

```bash
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
- `GET/POST/PUT/DELETE /deviceinterfaces`
- `GET/POST/PUT/DELETE /connections`
- `GET/POST/PUT/DELETE /connectiontypes`
- `GET/POST/PUT/DELETE /vlans`

### Diagrams
- `GET /diagram?format={ports|connections}&vlan={true|false}&colorports={true|false}`

## License

Copyright © 2026 Talleyrand-34 (t34@t34.dev)

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
