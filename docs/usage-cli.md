# CLI Usage Guide

A task-oriented walkthrough of the `nsl-graph` command line. For the exhaustive
flag list see the **[CLI Reference](cli-reference.md)**; for term definitions see
the **[Glossary](glossary.md)**.

- [1. Invocation & global flags](#1-invocation--global-flags)
- [2. The entity model (CRUD)](#2-the-entity-model-crud)
- [3. Building a topology from scratch](#3-building-a-topology-from-scratch)
- [4. Scanning devices (scan → edit → import)](#4-scanning-devices-scan--edit--import)
- [5. Scan profiles & the credential vault](#5-scan-profiles--the-credential-vault)
- [6. Discovering connections](#6-discovering-connections)
- [7. Diagrams (SVG & ASCII)](#7-diagrams-svg--ascii)
- [8. Export, summary & compare](#8-export-summary--compare)
- [9. Running the API server](#9-running-the-api-server)

---

## 1. Invocation & global flags

Run the built binary or `go run`:

```bash
go build -o nsl-graph .
./nsl-graph --help
# or, during development:
go run . --help
```

Global flags apply to every command:

| Flag | Default | Meaning |
|------|---------|---------|
| `-s, --source <path>` | `test-dbs/test.db` | CloverDB store (a directory). Created on first use. |
| `-b, --backend <type>` | `cloverdb` | Storage backend. |
| `-d, --debug` | `false` | Verbose debug output. |
| `-v, --verbose` | `false` | More console output. |
| `--outPath` / `--outFile` / `--outImage` | `out/` / `out.d2` / `out.svg` | Diagram output locations (see [§7](#7-diagrams-svg--ascii)). |

Every command operates on the store given by `-s`. Point it at whichever store
you want — `./nsl-graph print device -s test-dbs/demo.db`.

The command tree mirrors the data model:

```
nsl-graph
├── add | update | delete | print   <entity>     # CRUD over the data model
├── scan        run|host|network|import|connections|profile
├── diagram     connection|port
├── export      json
├── compare                                       # SNMP vs SSH scan diff
└── server                                        # HTTP API for the web UI
```

---

## 2. The entity model (CRUD)

The same four verbs work across every entity:

```bash
nsl-graph add    <entity> [flags]
nsl-graph print  <entity>
nsl-graph update <entity> [flags]
nsl-graph delete <entity> <name|id>
```

Entities (see the [Glossary](glossary.md) for precise definitions):

| Entity | What it is |
|--------|------------|
| `brand` | Equipment manufacturer (Cisco, Netgear…). |
| `modeltype` | Category of equipment: router, switch, firewall, AP… |
| `ostype` | OS/firmware family that selects the config parser (openwrt, opnsense…). Seeded automatically. |
| `model` | A product = brand + model type + (optional) os type. |
| `modelport` | A port on a model (the template). |
| `owner` | Ownership/administrative entity. |
| `zonetype` | Category of zone (building, rack, VLAN segment…). |
| `zone` | A network segment, with a zone type and an owner. |
| `device` | An instance of a model, in a zone, optionally tied to a scan profile. |
| `deviceport` | A physical port on a device (instance of a model port). |
| `deviceinterface` / `interfaceport` | Logical interfaces and their mapping to physical ports. |
| `vlan` | A VLAN definition. |
| `connection` / `connectiontype` | A link between two device ports, and its type. |

Most `add`/`update` commands take their fields as flags (use `--help` to see
them). Names are generally accepted where IDs are; print first if unsure:

```bash
nsl-graph add brand --name Cisco -s test-dbs/demo.db
nsl-graph print brand -s test-dbs/demo.db          # JSON list with ids
nsl-graph delete brand Cisco -s test-dbs/demo.db
```

---

## 3. Building a topology from scratch

Create the catalogue entities first, then the model, then device instances,
then wire them up. A minimal end-to-end example (all against `test-dbs/demo.db`):

```bash
S="-s test-dbs/demo.db"

# 1. Catalogue
nsl-graph add brand     --name OpenWrt-Brand $S
nsl-graph add modeltype --name router        $S
# ostypes are seeded (openwrt, opnsense, …) — list them:
nsl-graph print ostype  $S

# 2. A model = brand + model type + os type
nsl-graph add model --name WRT-1 --brand OpenWrt-Brand --model-type router --os-type openwrt $S

# 3. A zone (needs a zone type and an owner)
nsl-graph add zonetype --name building $S
nsl-graph add owner    --name Lab       $S
nsl-graph add zone     --name CoreRoom --location-type building --owner Lab $S

# 4. A device instance of the model, in the zone
nsl-graph add device --label rtr1 --model WRT-1 --zone CoreRoom $S

# 5. Inspect
nsl-graph print device  $S
nsl-graph print summary $S
```

`print summary` gives a high-level count of everything in the store. Adding
ports and connections follows the same pattern (`add modelport`, `add
deviceport`, `add connection`) — see the [CLI Reference](cli-reference.md#add).

> In practice you rarely build a topology by hand: you **scan** it (next
> section) and let the importer create the brands/models/zones/devices for you.

---

## 4. Scanning devices (scan → edit → import)

The scan commands follow one convention, identical to the web "Import devices"
workflow:

- **stdout** carries the machine artifact; **stderr** carries progress/status,
  so a scan pipes cleanly (`scan ... > plan.json`).
- The default artifact is an **import plan** — a JSON array of per-device plans,
  the same data the web UI lets you edit (per-IP `subnet` and `vlan_number`, plus
  the suggested name/zone, each annotated with a `confidence`/`reason`).
- `-H/--human` switches to a readable summary + interactive review/import.
- `--raw` (host/network) emits the raw `ScanResult` instead of the plan.

### The round-trip

```bash
# 1. Scan → emit an editable plan
nsl-graph scan run --method ssh --target 10.0.2.0/24 \
                   --profile owrt --os-type openwrt -s test-dbs/demo.db > plan.json

# 2. Edit the plan: fix VLAN ids, subnets, suggested names/zones
$EDITOR plan.json

# 3. Import the edited plan
nsl-graph scan import plan.json -s test-dbs/demo.db
```

A plan entry looks like:

```json
[
  {
    "device": {
      "device": { "ip": "10.0.2.1", "sys_name": "rtr1", "interfaces": [ ... ] },
      "brand": "", "model": "", "model_type": "router",
      "suggested_name": "rtr1", "suggested_zone": "Discovered", "os_type": "openwrt"
    },
    "interface_plans": [
      { "interface": { "name": "eth0.10" },
        "ip_mappings": [
          { "ip": "10.0.2.1",
            "subnet": "10.0.2.0/24",   // ← editable
            "vlan_number": "10",       // ← editable
            "confidence": "heuristic", "reason": "RFC1918 /24 guess" }
        ] }
    ],
    "summary": "1 interface, 1 IP, 1 VLAN"
  }
]
```

### `scan run` — the unified entry point

Mirrors the web/API `RunScan`. The method (`snmp`|`ssh`) is explicit; single
host vs subnet sweep is inferred from the target (a bare IP is one host; a CIDR
or comma/space-separated list is a batch).

```bash
# SNMP single host → plan on stdout
nsl-graph scan run 10.0.2.245 -s test-dbs/demo.db > plan.json

# SNMP subnet sweep
nsl-graph scan run --method snmp --target 10.0.2.0/24 --community public -s test-dbs/demo.db > plan.json

# SSH (needs a profile for credentials + an os-type)
nsl-graph scan run --method ssh --target 10.0.2.0/24 --profile owrt --os-type openwrt -s test-dbs/demo.db > plan.json

# Human summary + interactive import (no JSON)
nsl-graph scan run 10.0.2.245 -H -s test-dbs/demo.db
```

### `scan host` / `scan network`

`host` queries one host (IP or `~/.ssh/config` alias), `network` sweeps one or
more subnets. Both obey the same convention: **plan JSON by default**, `-H` for
the interactive summary, `--raw` for the legacy `ScanResult`, `--auto-import` to
import directly.

```bash
nsl-graph scan host 10.0.2.1 -s test-dbs/demo.db > plan.json     # plan (default)
nsl-graph scan host 10.0.2.1 -H -s test-dbs/demo.db              # human + review
nsl-graph scan host 10.0.2.1 --raw -o host.json -s test-dbs/demo.db   # raw ScanResult to a file
nsl-graph scan host opnsense --scan-source ssh --auto-import -s test-dbs/demo.db  # SSH via alias, import now

nsl-graph scan network 10.0.0.0/24 10.0.1.0/24 -s test-dbs/demo.db > plan.json
```

> **Behaviour change:** `scan host`/`scan network` used to print a human summary
> by default; they now emit the plan JSON on stdout. Use `-H` for the old view,
> `--raw` for the old `ScanResult` output.

### `scan import`

Auto-detects the input shape:

- A JSON **array** = an import plan (from `scan run`/`host`/`network`, possibly
  edited). Executed as-is; `-H` reviews each device (Approve/Edit/Skip/Quit).
- A JSON **object** = a raw `ScanResult` (from `--raw` or older scans). It is
  discovered, analyzed, then imported (`--auto-import`/`--review`).

```bash
nsl-graph scan import plan.json -s test-dbs/demo.db          # execute edited plan
nsl-graph scan import plan.json -H -s test-dbs/demo.db       # review each device first
nsl-graph scan import host.json --auto-import -s test-dbs/demo.db   # raw ScanResult
```

---

## 5. Scan profiles & the credential vault

A **scan profile** stores reusable per-host parameters (SNMP community/version,
SSH user/key/password, os-type…) so you enter them once. There are two kinds:

- **device** — bound to a host (`--host`); auto-applied when that host is scanned.
- **generic** (`--generic`) — credentials only, no host; used as an explicit SSH
  fallback (e.g. `--generic-profile <name>` / `--profile` for `scan run`).

```bash
nsl-graph scan profile add owrt --host 10.0.2.1 --ssh-user root --os-type openwrt --scan-source ssh -s test-dbs/demo.db
nsl-graph scan profile add lab-creds --generic --ssh-user admin --ssh-key ~/.ssh/lab -s test-dbs/demo.db
nsl-graph scan profile list   -s test-dbs/demo.db    # KIND column = device / generic
nsl-graph scan profile show   owrt -s test-dbs/demo.db   # SSH password is never printed
nsl-graph scan profile delete owrt -s test-dbs/demo.db
```

When a profile carries an SSH **password**, it is encrypted by the **credential
vault** (AES-256-GCM under the vault's data key, never stored in clear). You set
a master passphrase the first time; later SSH scans prompt to unlock it once per
run. SNMP-only profiles never touch the vault.

Saved profiles are applied automatically: `scan host 10.0.2.1` picks up a device
profile whose host matches, and explicit flags always override profile values.
Persist the effective parameters of a one-off scan with `--save-profile <name>`.

---

## 6. Discovering connections

`scan connections` gathers layer-2/1 adjacency from a set of hosts (LLDP over
SNMP or SSH, CDP, and bridge MAC tables), correlates it into edges, lets you
review discrepancies, and commits confirmed links. It is **observe-only** — it
never configures the targets (LLDP/SNMP must already be enabled out of band).

Output: **JSON gather to stdout by default** (status on stderr); `-H/--human`
for the readable summary + interactive review.

```bash
# Every DB device with a mgmt IP and a resolvable profile (the default target)
nsl-graph scan connections --from-db -s test-dbs/demo.db

# Sweep a CIDR, collect from responders, commit without prompting
nsl-graph scan connections --subnet 10.0.0.0/24 --community public --yes -s test-dbs/demo.db

# One source only, don't write to the DB, save the gather
nsl-graph scan connections --collector ssh-lldp --dry-run --output gather.json -s test-dbs/demo.db

# Human review
nsl-graph scan connections --from-db -H -s test-dbs/demo.db
```

Targets (`--from-db`, `--subnet`, `--profiles`) combine freely; sources default
to all-merged, restrict with `--collector`. Runtime SSH credentials for hosts
without a device profile come from `--ssh-config`, inline `--ssh-user`/`--ssh-key`,
or `--generic-profile`. See the [CLI Reference](cli-reference.md#scan-connections)
for the full flag set.

---

## 7. Diagrams (SVG & ASCII)

Render the stored topology as a [d2](https://d2lang.com) diagram. Two focuses —
`connection` and `port` — each with an optional `--vlan` coloring mode.

```bash
# SVG (default) → out/out.d2 + out/out.svg
nsl-graph diagram connection -s test-dbs/demo.db
nsl-graph diagram port --vlan --color-target ports -s test-dbs/demo.db
nsl-graph diagram connection --vlan --vlan-scope all -s test-dbs/demo.db
```

### ASCII output

Pass `--ascii` to render the same diagram as ASCII art using d2's in-process
renderer (no external binary). The art is **printed to stdout** and also written
to a `.txt` file next to the `.d2` source. `--charset` chooses `unicode`
(box-drawing, default) or `ascii` (plain `+ - |`). Both flags are shared by every
diagram subcommand.

```bash
nsl-graph diagram connection --ascii -s test-dbs/demo.db                 # to terminal + out/out.txt
nsl-graph diagram connection --ascii --charset ascii -s test-dbs/demo.db # plain ASCII
```

```
┌────────┐      ┌────────┐
│ router │──────│ switch │
└────────┘      └────────┘
```

Diagram output paths are controlled by the global `--outPath`/`--outFile`/`--outImage`
flags (the ASCII `.txt` filename is derived from `--outImage`, e.g. `out.svg → out.txt`).

---

## 8. Export, summary & compare

```bash
# Dump the entire store as JSON
nsl-graph export json -s test-dbs/demo.db > backup.json

# Network summary (counts / detail)
nsl-graph print summary -s test-dbs/demo.db

# Compare an SNMP scan against an SSH scan of the same host
nsl-graph compare --help
```

---

## 9. Running the API server

The web UI talks to a Go HTTP API. Start it with:

```bash
nsl-graph server --port 8081 -s test-dbs/demo.db
```

Then serve the PHP frontend against it (see the **[Web UI Usage
Guide](usage-webui.md)**):

```bash
php -S localhost:8091 -t frontend/php
```

The API endpoints are documented in **[Frontend API](frontend-api.md)**;
logging/observability flags (`--log-level`, `--log-format`, `--log-file`) are on
the `server` command.
