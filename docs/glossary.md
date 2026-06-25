# Glossary

The canonical vocabulary for NSL-Graph, shared across the **code**, **CLI**, and
**web UI**. When a term is spelled differently in different layers (a JSON key, a
DB collection, a CLI flag, a UI label), the canonical term is given first and the
layer-specific spellings are listed as *aliases*. Use the canonical term in new
code, docs, and UI text.

> Backend note: the data store is **CloverDB** (a document/NoSQL store), not
> SQLite. See [database-schema.md](database-schema.md).

---

## Core data model

### Brand
A hardware **manufacturer** (e.g. Cisco, Netgear, OpenWrt).
- Collection `brands`; API `/brands`; CLI `add/update/delete brand`; UI "Brand".

### Device Class
The **category** of equipment — router, switch, firewall, access point, …
- Collection `devclasses`; API `/deviceclasses`; UI entity `devclass` ("Device Class").
- Aliases: `device_class` / `device_class_name` (JSON), `class` (model form field),
  `DevClass` (Go type). **Canonical: "Device Class".**

### Model
A specific **product**: a Brand + a Device Class under a model name (e.g.
"Catalyst 9300"). A Model is a *template* — Devices are instances of it, and it
owns the Model Ports that define the device's physical ports.
- Collection `models`; API `/models`; UI entity `modeldevice` ("Model").
- Alias: `modeldevice` is the UI/route value; the user-facing label is **Model**.

### Device
A concrete piece of equipment: an **instance of a Model**, placed in a Zone, owned
by a Proprietary, optionally carrying IP addresses, a scan **Profile**, and the
`unmanaged` / `invisible` flags.
- Collection `devices`; API `/devices`; CLI `add/update/delete device`; UI "Device".
- A device's human identifier is its **label** (Go `Device.Name`, JSON `label`).
  **Canonical: "label"** (not "name") for the device's display string.

### Zone
A **network segment** (logical or physical) devices live in. Zones form a
hierarchy via a parent (`father`) and each has a Zone Type and a Proprietary.
- Collection `zones`; API `/zones`; CLI `add/delete zone`; UI "Zone".

### Zone Type
The **category of a Zone** (e.g. building, rack, VLAN domain).
- Collection `zonetypes`; API `/zonetypes`; UI entity `zonetype`.
- Alias: the field on a Zone is `location_type`. **Canonical UI label: "Zone Type"**
  (do not label it "Location Type").

### Proprietary
The **owner / administrative entity** responsible for a Zone or Device (an
organization or admin domain). The word is used as a noun meaning "proprietor".
- Collection `proprietaries`; API `/proprietaries`; UI "Proprietary".
- Known awkward term; kept for compatibility (see *Terminology notes*).

### VLAN
A virtual LAN, identified by its VLAN id, optionally named. Membership on a port
is **tagged** or **untagged**.
- Collection `vlans`; API `/vlans`; CLI `add/delete vlan`; UI "VLAN".

### Model Port
A port defined on a **Model** (the template): a name, an (x, y) diagram position,
and whether it may hold multiple connections.
- Collection `modelports`; API `/modelports`; UI entity `modelport` ("ModelPort").

### Device Port
A **physical port on a Device** — an instance of a Model Port — carrying a MAC and
per-port VLAN configs (tagged/untagged). Keyed by `device_id` + `model_port_id`.
- Collection `deviceports`; API `/deviceports`; UI entity `deviceport` ("DevicePort").
- CLI alias: the `add` subcommand is `devport`.

### Device Interface
A **logical interface** on a device (e.g. `eth0.1`, a bridge, a sub-interface),
with its own VLAN configs, IP addresses, and optional Wi-Fi SSID/security.
- Collection `deviceinterfaces`. Joined to Device Ports through Interface Ports.

### Interface Port
A pure **join** between a Device Interface and a Device Port (`interface_id` +
`model_port_id`, with `device_id` as a denormalized query key).
- Collection `interfaceports`.

### Connection
A persisted **link between two Device Ports**, with a Connection Type. Its
effective/`missing` VLANs are computed at read time by intersecting both ends.
- Collection `connections`; API `/connections`; UI entity `connections`.
- Contrast with **Edge** (a *candidate* link from a scan, below).

### Connection Type
The **medium/category** of a Connection (e.g. ethernet, fiber, wifi).
- Collection `connectiontypes`; API `/connectiontypes`; UI entity `connectiontype`.

---

## Scanning & topology

### Scan profile
Reusable, named scan parameters/credentials. Identified by a **unique name** and
characterized by two independent axes:
- **kind**: `device` (bound to a Host, auto-matched by IP) or `generic`
  (credential-only, reusable across hosts, never auto-matched).
- **scan source**: `snmp` or `ssh`.
The web Create-profile form presents these as four buttons (device/generic ×
snmp/ssh). Collection `scanprofiles`; API `/scan/profiles`.

### Profile (device association)
The scan profile a **Device** is tied to (`Device.profile`, a profile *name*). Set
on the dashboard device form, and **auto-assigned on import** to whatever profile a
scan used. Scan-connections (from-db) requires every device to have one.

### Device type
The **operating system / firmware family** used to pick the right config parser
over SSH — `openwrt`, `opnsense`, `fortinet`, `cisco`, … It is **not** the hardware
Model. UI label: "OS / firmware type"; CLI flag `--device-type`; field `device_type`.
**Canonical meaning: OS/firmware**, distinct from Model.

### Scan (device scan / import)
Discovering one or more **devices** (SNMP or SSH) for import into the DB — the
*Import devices* page and `scan host` / `scan network` / `scan run`.

### Scan connections
Discovering **layer-2/1 links** between known devices by correlating LLDP / CDP /
bridge-FDB evidence — the *Scan connections* page and `scan connections`. Read-only;
never configures targets.

### Discovered Device
A device returned by a scan (`DiscoveredDevice`), pre-import: classification
(brand/model/class), suggested name/zone, and the **profile** it was scanned with.

### Edge
A **candidate** connection derived from a connections scan (LLDP/CDP/FDB evidence),
shown in the review table before you import it. Confidence: `confirmed` |
`candidate` | `weak`. Importing an Edge creates a **Connection**.

### Collector / source
A single evidence source for a connections scan: `snmp-lldp`, `snmp-cdp`,
`snmp-fdb`, `ssh-lldp`, `ssh-fdb`, `local-lldp`. By default all are merged.

### Discrepancy
A non-fatal **warning** from a connections scan, with a `kind`: e.g.
`host-not-in-db`, `unknown-remote`, `multi-neighbor`, `ssh-no-credentials`,
`ssh-key-missing`, `device-no-profile` (a from-db device excluded for having no
scan profile).

### Intermediary / Placeholder
An **unknown middle device** — a globally-administered MAC seen by ≥2 hosts that
isn't in the DB (an *intermediary*). Opting in materializes a single shared
*placeholder* unmanaged device (in an "Unknown infrastructure" zone) linking the
real devices through it.

### Unmanaged (device)
A device whose switch **replicates all VLANs through all ports** (`is_unmanaged`).
Distinct from a scan profile of kind "generic". UI: "Unmanaged VLANs".

### Invisible (device)
A device hidden from normal views (`is_invisible`) — used for placeholders.

---

## Credentials & security

### Credential vault
A single **server-side vault**: one master passphrase unlocks a random data key
(held in memory) that AES-256-GCM-encrypts stored SSH secrets. Auto-locks on idle
and on restart. API `/vault/{status,init,unlock,lock}`; UI control in the app bar.
See [frontend-api.md](frontend-api.md).

---

## Terminology notes (known aliases kept for compatibility)

These spellings differ across layers but mean the same thing; the canonical term
is on the left. Renaming the storage/API forms would be a breaking change, so they
are documented rather than churned:

| Canonical | Aliases / where they appear |
|-----------|------------------------------|
| Device Class | `devclass` (UI entity), `deviceclasses` (API), `device_class`/`class` (JSON), `DevClass` (Go) |
| Model | `modeldevice` (UI entity/route value) |
| Zone Type | `location_type` (Zone field) — **don't** surface as "Location Type" |
| Proprietary | "owner/administrative entity" (no rename; pervasive in DB/API/UI) |
| Device label | `Device.Name` (Go), `label` (JSON/UI) |
| Device type (OS/firmware) | `--device-type` (CLI), `device_type` (JSON) — **not** the hardware Model |
| Device Port | `devport` (CLI `add` subcommand), `deviceport` (UI) |

Internal note: a `cmd/modify/` command group duplicates `cmd/update/` and is not
registered in the CLI (`update` is the live one); treat `update` as canonical.
