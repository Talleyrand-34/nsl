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

### add

Add new network entities. To change an existing entity, use the `update` command. (`modify` remains as an alias.)

#### add brand
```bash
nsl-graph add brand --name <brand-name>
```
Add a new equipment brand.

**Example:**
```bash
nsl-graph add brand --name "Cisco"
nsl-graph add brand --name "Juniper"
```

#### add deviceclass
```bash
nsl-graph add deviceclass --name <class-name>
```
Add a new device class (router, switch, firewall, etc.).

**Example:**
```bash
nsl-graph add deviceclass --name "router"
nsl-graph add deviceclass --name "switch"
```

#### add proprietary
```bash
nsl-graph add proprietary --name <proprietary-name>
```
Add a proprietary/ownership entity.

**Example:**
```bash
nsl-graph add proprietary --name "TDS"
nsl-graph add proprietary --name "Navantia"
```

#### add zonetype
```bash
nsl-graph add zonetype --name <zonetype-name>
```
Add a zone type (physical, logical, etc.).

**Example:**
```bash
nsl-graph add zonetype --name "physical"
nsl-graph add zonetype --name "logical"
```

#### add zone
```bash
nsl-graph add zone --name <zone-name> --zonetype <type> --proprietary <owner> [--father <parent-zone>]
```
Add a network zone.

**Example:**
```bash
nsl-graph add zone --name "datacenter" --zonetype "physical" --proprietary "TDS"
nsl-graph add zone --name "rack01" --zonetype "physical" --proprietary "TDS" --father "datacenter"
```

#### add model
```bash
nsl-graph add model --name <model-name> --brand <brand> --class <device-class>
```
Add a device model.

**Example:**
```bash
nsl-graph add model --name "ISR4431" --brand "Cisco" --class "router"
nsl-graph add model --name "EX4300" --brand "Juniper" --class "switch"
```

#### add device
```bash
nsl-graph add device --name <device-name> --model <model-name> --zonename <zone> --proprietary <owner>
```
Add a network device.

**Example:**
```bash
nsl-graph add device --name "router01" --model "ISR4431" --zonename "datacenter" --proprietary "TDS"
nsl-graph add device --name "switch01" --model "EX4300" --zonename "datacenter" --proprietary "TDS"
```

#### add modelport
```bash
nsl-graph add modelport --name <port-name> --posx <x-position> --posy <y-position> --modelname <model>
```
Add a port to a device model.

**Example:**
```bash
nsl-graph add modelport --name "GigE0/0/0" --posx 0 --posy 0 --modelname "ISR4431"
nsl-graph add modelport --name "GigE0/0/1" --posx 1 --posy 0 --modelname "ISR4431"
```

#### add deviceport
```bash
nsl-graph add deviceport --deviceid <device-id> --modelportid <port-id> [--macaddress <mac>] [--vlan-configs 100:tagged,200:untagged]
```
Associate a model port with a device instance. **Create-only**: errors if the port
already exists — use `update deviceport` to change it.

**Example:**
```bash
nsl-graph add deviceport --deviceid 1 --modelportid 1
nsl-graph add deviceport --deviceid 1 --modelportid 2 --vlan-configs 100:tagged
```

#### add connection
```bash
nsl-graph add connection --from-device <device-id> --from-model-port-id <port-id> --to-device <device-id> --to-model-port-id <port-id>
```
Create a connection between device ports.

**Example:**
```bash
nsl-graph add connection --from-device 1 --from-model-port-id 1 --to-device 2 --to-model-port-id 1
```

#### add connectiontype
```bash
nsl-graph add connectiontype --name <name>
```
Add a connection type (e.g. ethernet, fiber, wireless).

### update

Update existing entities (the counterpart of `add`). Most entities are updated by
`--id`; run `nsl-graph update <entity> --help` for the exact flags. Two examples
relevant to ports/interfaces:

```bash
# Change a device port's MAC and/or VLANs (identified by device + model-port IDs)
nsl-graph update deviceport --deviceid <id> --modelportid <id> --vlan-configs 100:tagged

# Change a device interface's VLAN configs (identified by interface ID)
nsl-graph update deviceinterface --id <id> --vlan-configs 10:untagged,20:tagged
```

> Note: `interfaceport` is a pure link (interface↔port) with no editable fields —
> there is no `update interfaceport`; delete and re-add instead.

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
nsl-graph print device      # alias: nsl-graph print devices
```
Display all devices with their interfaces.

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

#### print connectiontype
```bash
nsl-graph print connectiontype
```
Display all connection types.

#### print connection
```bash
nsl-graph print connection
```
Display all connections. Each connection includes computed VLAN information:
- `vlans`: Array of VLANs present on **both** port ends (intersection)
- `missing_vlans`: Array of VLANs present on **only one** port end (symmetric difference)

Example output:
```json
{
  "id": "abc123",
  "fromdevice": "switch01",
  "frommodel": "lan1",
  "todevice": "router01",
  "tomodel": "wan0",
  "vlans": [{"vlan_id": "1", "tagged": false}],
  "missing_vlans": [{"vlan_id": "2", "tagged": false}]
}
```

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

Generate network diagrams. Output files default to `out/out.d2` and `out/out.svg`.
There are two focuses — `connection` and `port` — each with an optional `--vlan`
mode for VLAN coloring.

#### diagram connection
```bash
nsl-graph diagram connection [--all-ports] [--vlan [--vlan-scope <s>] [--color-target <t>]]
```
Sorted connection diagram. With `--vlan` the connections are colored per VLAN and
a legend is added.

#### diagram port
```bash
nsl-graph diagram port [--all-ports] [--vlan [--vlan-scope <s>] [--color-target <t>]]
```
Port-focused diagram (lists all ports per device). Same flags as `connection`.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--all-ports` | `false` | Include ports with no connections. |
| `--vlan` | `false` | Color by VLAN and add a legend. |
| `--vlan-scope` | `untagged` | With `--vlan`: `untagged` (one line per link, colored by the source port's untagged VLAN) or `all` (one colored line per VLAN in the **intersection** of both port ends). |
| `--color-target` | `both` | With `--vlan`: `both` (connections + port nodes), `connections` (port nodes plain), or `ports` (connection lines plain). |

**Examples:**
```bash
# Plain connection diagram
nsl-graph diagram connection -s demo.db

# VLAN-colored, every shared VLAN as a separate line per trunk link
nsl-graph diagram connection --vlan --vlan-scope all -s demo.db

# Port diagram, color port nodes only
nsl-graph diagram port --vlan --color-target ports -s demo.db
```

### scan

Query devices via SNMP (and optionally SSH config). See the README for full
scanning usage; the subcommands below manage **scan profiles** — reusable
per-host parameters that are auto-applied when a host matches.

#### scan profile

```bash
nsl-graph scan profile add <name> --host <ip> [--snmp-community ...] [--ssh-user ...] \
                                  [--ssh-password ...] [--device-type ...] [--scan-source ssh]
nsl-graph scan profile list
nsl-graph scan profile show <name>     # SSH password is never printed
nsl-graph scan profile delete <name>
```

If `--ssh-password` is given, you are prompted for a **passphrase** that encrypts
it (AES-256-GCM, scrypt); it is never stored in clear. SNMP-only profiles need
no passphrase.

#### scan host / scan network profile flags

```bash
nsl-graph scan host <ip> --profile <name>        # use a named profile
nsl-graph scan host <ip>                         # auto-applies a profile whose host matches
nsl-graph scan host <ip> --save-profile <name>   # persist the effective parameters
```

Explicit flags always override profile values. When an SSH scan uses a stored
password you are prompted for its passphrase; a wrong passphrase fails.

#### scan connections

Discover layer-2/1 links between hosts and import them as connections. Every
selected host is scanned with **all available sources** (LLDP via SNMP or SSH,
CDP, and bridge MAC tables); the evidence is merged, correlated into edges, and —
after you review any discrepancies — committed to the DB. The full per-host
gather is also emitted as JSON.

```bash
nsl-graph scan connections [--from-db] [--subnet <cidr>] [--profiles] \
                           [--collector <name>] [--yes|--dry-run] [--output <file>]
```

**Targets** (combine freely; default is `--from-db`):

| Flag | Description |
|------|-------------|
| `--from-db` | Every device with a management IP and a resolvable scan profile. |
| `--subnet <cidr>` | SNMP-sweep the CIDR, then collect from each responder. |
| `--profiles` | The host of every saved scan profile. |
| `--local` | Also collect LLDP from the machine running the tool (default `true`); set `--local-device <label>` so it maps to that device's ports. |

**Sources:** by default all available sources per host are used and merged.
`--collector <name>` restricts to exactly one of `snmp-lldp`, `snmp-cdp`,
`snmp-fdb`, `ssh-lldp`, `ssh-fdb`, `local-lldp` (no merge/review). (Note:
`-s`/`--source` is the global database-path flag.)

**Intermediary ("middle") devices:** transparent switches that don't speak
LLDP/SNMP (e.g. a Netgear "Plus" switch) won't appear as endpoints, but the
`ssh-fdb`/`snmp-fdb` sources read the bridge forwarding tables, and the tool
flags any globally-administered, known-vendor MAC learned by **two or more**
hosts (and not itself an LLDP endpoint or a scanned device) as an
**intermediary device detected in the middle** — reported in the output (and
the `intermediaries` array of the JSON gather), not imported.

**Output format:** JSON to stdout by **default** (status messages go to stderr,
so `scan connections ... | jq` works); pass **`-H`/`--human`** for the readable
summary and interactive review. `--output <file>` always writes the full JSON
gather to a file as well.

**Review & commit:** edges are labelled by confidence — `confirmed` (seen from
both ends), `candidate` (one direct observation), `weak` (FDB-only), plus
`possible` (one end identified via a stored port MAC — shown, not imported).
With `-H` you confirm each edge interactively; in either mode `--yes` commits
confirmed/candidate edges and `--dry-run` writes nothing. Imported connections
record their **provenance** (`discovered_via`). An observed neighbour that is not
in the DB (e.g. an unmanaged switch) is reported as a discrepancy and never
imported.

> **Prerequisite:** the tool only *collects* — it never configures the targets.
> `lldpd` and/or SNMP must already be enabled on each host (set up out of band
> over SSH). Hosts without them simply yield no evidence (a non-fatal per-host
> error).

**Examples:**
```bash
nsl-graph scan connections --from-db
nsl-graph scan connections --subnet 10.0.0.0/24 --community public
nsl-graph scan connections --source ssh-lldp --dry-run
nsl-graph scan connections --from-db --yes --output gather.json
```

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
nsl-graph add brand --name "Cisco"
nsl-graph add deviceclass --name "router"
nsl-graph add proprietary --name "MyCompany"
nsl-graph add zonetype --name "physical"

# Create zones
nsl-graph add zone --name "datacenter" --zonetype "physical" --proprietary "MyCompany"

# Create model and devices
nsl-graph add model --name "ISR4431" --brand "Cisco" --class "router"
nsl-graph add device --name "router01" --model "ISR4431" --zonename "datacenter" --proprietary "MyCompany"
nsl-graph add device --name "router02" --model "ISR4431" --zonename "datacenter" --proprietary "MyCompany"

# Add ports to model
nsl-graph add modelport --name "GigE0/0/0" --posx 0 --posy 0 --modelname "ISR4431"
nsl-graph add modelport --name "GigE0/0/1" --posx 1 --posy 0 --modelname "ISR4431"

# Associate ports with devices
nsl-graph add deviceport --deviceid 1 --modelportid 1
nsl-graph add deviceport --deviceid 1 --modelportid 2
nsl-graph add deviceport --deviceid 2 --modelportid 1
nsl-graph add deviceport --deviceid 2 --modelportid 2

# Create connection
nsl-graph add connection --from-device 1 --from-model-port-id 1 --to-device 2 --to-model-port-id 1

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