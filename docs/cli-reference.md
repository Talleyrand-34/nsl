# CLI Reference

The `nsl-graph` command-line tool provides comprehensive network management capabilities. All commands operate on a SQLite database and support various output formats.

## Basic Usage

```bash
# Run with Go
go run main.go [command] [flags]

# Or use built binary
./nsl-graph [command] [flags]
```

## Global Flags

These flags are available for all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--source` | `-s` | string | `test.db` | SQLite database file path |
| `--config-file` | `-c` | string | | Configuration file path |
| `--outPath` | | string | `out/` | Output directory for generated files |
| `--outFile` | | string | `out.d2` | Output script filename |
| `--outImage` | | string | `out.svg` | Output image filename |
| `--verbose` | `-v` | bool | `false` | Enable verbose output |
| `--debug` | `-d` | bool | `false` | Enable debug output |

## Commands

### modify

Add or update network entities.

#### modify brand
```bash
nsl-graph modify brand --name <brand-name>
```
Add a new equipment brand.

**Example:**
```bash
nsl-graph modify brand --name "Cisco"
nsl-graph modify brand --name "Juniper"
```

#### modify deviceclass
```bash
nsl-graph modify deviceclass --name <class-name>
```
Add a new device class (router, switch, firewall, etc.).

**Example:**
```bash
nsl-graph modify deviceclass --name "router"
nsl-graph modify deviceclass --name "switch"
```

#### modify proprietary
```bash
nsl-graph modify proprietary --name <proprietary-name>
```
Add a proprietary/ownership entity.

**Example:**
```bash
nsl-graph modify proprietary --name "TDS"
nsl-graph modify proprietary --name "Navantia"
```

#### modify zonetype
```bash
nsl-graph modify zonetype --name <zonetype-name>
```
Add a zone type (physical, logical, etc.).

**Example:**
```bash
nsl-graph modify zonetype --name "physical"
nsl-graph modify zonetype --name "logical"
```

#### modify zone
```bash
nsl-graph modify zone --name <zone-name> --zonetype <type> --proprietary <owner> [--father <parent-zone>]
```
Add a network zone.

**Example:**
```bash
nsl-graph modify zone --name "datacenter" --zonetype "physical" --proprietary "TDS"
nsl-graph modify zone --name "rack01" --zonetype "physical" --proprietary "TDS" --father "datacenter"
```

#### modify model
```bash
nsl-graph modify model --name <model-name> --brand <brand> --class <device-class>
```
Add a device model.

**Example:**
```bash
nsl-graph modify model --name "ISR4431" --brand "Cisco" --class "router"
nsl-graph modify model --name "EX4300" --brand "Juniper" --class "switch"
```

#### modify device
```bash
nsl-graph modify device --name <device-name> --model <model-name> --zonename <zone> --proprietary <owner>
```
Add a network device.

**Example:**
```bash
nsl-graph modify device --name "router01" --model "ISR4431" --zonename "datacenter" --proprietary "TDS"
nsl-graph modify device --name "switch01" --model "EX4300" --zonename "datacenter" --proprietary "TDS"
```

#### modify modelport
```bash
nsl-graph modify modelport --name <port-name> --posx <x-position> --posy <y-position> --modelname <model>
```
Add a port to a device model.

**Example:**
```bash
nsl-graph modify modelport --name "GigE0/0/0" --posx 0 --posy 0 --modelname "ISR4431"
nsl-graph modify modelport --name "GigE0/0/1" --posx 1 --posy 0 --modelname "ISR4431"
```

#### modify deviceport
```bash
nsl-graph modify deviceport --deviceid <device-id> --modelportid <port-id>
```
Associate a model port with a device instance.

**Example:**
```bash
nsl-graph modify deviceport --deviceid 1 --modelportid 1
nsl-graph modify deviceport --deviceid 1 --modelportid 2
```

#### modify connection
```bash
nsl-graph modify connection --from-device <device-id> --from-model-port-id <port-id> --to-device <device-id> --to-model-port-id <port-id>
```
Create a connection between device ports.

**Example:**
```bash
nsl-graph modify connection --from-device 1 --from-model-port-id 1 --to-device 2 --to-model-port-id 1
```

### print

Display information about network entities in JSON format.

#### print brand
```bash
nsl-graph print brand
```
Display all brands.

#### print deviceclass
```bash
nsl-graph print deviceclass
```
Display all device classes.

#### print proprietary
```bash
nsl-graph print proprietary
```
Display all proprietary entities.

#### print zonetype
```bash
nsl-graph print zonetype
```
Display all zone types.

#### print zone
```bash
nsl-graph print zone
```
Display all zones.

#### print model
```bash
nsl-graph print model
```
Display all device models.

#### print device
```bash
nsl-graph print device
```
Display all devices.

#### print devices
```bash
nsl-graph print devices [flags]
```
Display devices with optional detailed information.

**Flags:**
- `-a, --all-model-ports`: Include all available ports
- `-i, --info-ports`: Include port usage information
- `-p, --possible-ports`: Show cartesian product of possible ports

#### print modelport
```bash
nsl-graph print modelport
```
Display all model ports.

#### print devport
```bash
nsl-graph print devport
```
Display all device ports.

#### print connection
```bash
nsl-graph print connection
```
Display all connections.

#### print possibleports
```bash
nsl-graph print possibleports
```
Display possible port combinations.

### export

Export network data in various formats.

#### export json
```bash
nsl-graph export json
```
Export all network data as JSON.

### diagram

Generate network diagrams.

```bash
nsl-graph diagram
```
Generate a D2 diagram script and SVG image of the network topology.

**Output files:**
- D2 script: `{outPath}/{outFile}` (default: `out/out.d2`)
- SVG image: `{outPath}/{outImage}` (default: `out/out.svg`)

### server

Start the HTTP API server.

```bash
nsl-graph server [--port <port>]
```

**Flags:**
- `--port`: Server port (default: 8080)

**Example:**
```bash
nsl-graph server --port 8081
```

The server provides REST API endpoints for all network operations. See [Frontend API documentation](frontend-api.md) for endpoint details.

## Complete Workflow Example

Here's a complete example of setting up a simple network:

```bash
# Set up basic entities
nsl-graph modify brand --name "Cisco"
nsl-graph modify deviceclass --name "router"
nsl-graph modify proprietary --name "MyCompany"
nsl-graph modify zonetype --name "physical"

# Create zones
nsl-graph modify zone --name "datacenter" --zonetype "physical" --proprietary "MyCompany"

# Create model and devices
nsl-graph modify model --name "ISR4431" --brand "Cisco" --class "router"
nsl-graph modify device --name "router01" --model "ISR4431" --zonename "datacenter" --proprietary "MyCompany"
nsl-graph modify device --name "router02" --model "ISR4431" --zonename "datacenter" --proprietary "MyCompany"

# Add ports to model
nsl-graph modify modelport --name "GigE0/0/0" --posx 0 --posy 0 --modelname "ISR4431"
nsl-graph modify modelport --name "GigE0/0/1" --posx 1 --posy 0 --modelname "ISR4431"

# Associate ports with devices
nsl-graph modify deviceport --deviceid 1 --modelportid 1
nsl-graph modify deviceport --deviceid 1 --modelportid 2
nsl-graph modify deviceport --deviceid 2 --modelportid 1
nsl-graph modify deviceport --deviceid 2 --modelportid 2

# Create connection
nsl-graph modify connection --from-device 1 --from-model-port-id 1 --to-device 2 --to-model-port-id 1

# Generate diagram
nsl-graph diagram

# View results
nsl-graph print device
nsl-graph print connection
```

## Database Files

The tool uses SQLite database files:
- **Default**: `test.db` (created automatically)
- **Custom**: Specify with `-s` flag
- **Location**: Current working directory or absolute path

## Output Files

Generated files are placed in the output directory:
- **D2 Script**: Network diagram source code
- **SVG Image**: Rendered network diagram
- **JSON Export**: Complete network data export