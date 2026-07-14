# Web UI Usage Guide

A task-oriented walkthrough of the NSL-Graph web interface (the PHP frontend).
For the HTTP endpoints behind each action see **[Frontend API](frontend-api.md)**;
for the equivalent commands see the **[CLI Usage Guide](usage-cli.md)**; for terms
see the **[Glossary](glossary.md)**.

- [1. Starting the stack](#1-starting-the-stack)
- [2. The app bar (every page)](#2-the-app-bar-every-page)
- [3. Dashboard — manage data & view the diagram](#3-dashboard--manage-data--view-the-diagram)
- [4. Import devices — scan → configure → import](#4-import-devices--scan--configure--import)
- [5. Scan connections — discover L2/L1 links](#5-scan-connections--discover-l2l1-links)
- [6. The credential vault](#6-the-credential-vault)

---

## 1. Starting the stack

The web UI is a thin PHP frontend that calls the Go HTTP API. Start both:

```bash
# 1. API server (reads/writes the CloverDB store)
nsl-graph server --port 8081 -s demo.db

# 2. PHP frontend (talks to the API)
php -S localhost:8091 -t frontend/php
```

Open **http://localhost:8091**. The frontend defaults to the API at
`http://localhost:8081`; change it at runtime with the **API Base URL** control
in the app bar (see below). All state is stored by the API in the `-s` store —
the frontend holds nothing but your session preferences.

---

## 2. The app bar (every page)

Every page shares one header:

| Control | What it does |
|---------|--------------|
| **Section nav** | Switches between the three sections: **Dashboard**, **Import devices**, **Scan connections**. The active one is highlighted. |
| **Vault pill + Unlock/Lock** | Status of the [credential vault](#6-the-credential-vault) for stored SSH secrets, with unlock/lock actions. |
| **API Base URL + Update** | Point the frontend at a different API instance; the field is remembered for your session. |
| **Dark / Light** | Toggle dark mode (remembered in the browser). |

---

## 3. Dashboard — manage data & view the diagram

The **Dashboard** (`main.php`) is the CRUD console plus a live diagram.

### CRUD forms (left)

Two dropdowns drive the form:

1. **Choose action** — `Get`, `Add`, `Update`, `Delete`.
2. **Choose entity** — Brand, Model Type, OS Type, Owner, Zone Type, Zone,
   Model, Device, ModelPort, DevicePort, connections, connectiontypes, VLAN.

Changing either selector reloads the matching form (e.g. *Add* × *Device*). The
form fields map directly to the entity (see the [Glossary](glossary.md)). A
typical order to populate a fresh store is: Brand → Model Type → Model → Zone
Type → Owner → Zone → Device, then Ports and Connections.

### Inline "create object" for composed fields

When an entity references a parent that doesn't exist yet (e.g. adding a *Model*
needs a *Brand*), you don't have to leave the form: each composed field has a
**create-object** affordance that opens an inline sub-form to create the parent
(recursively, if the parent itself has composed fields). When you finish, the new
object is selected back in the original form. This is powered by
`inline-create.js`.

### Diagram pane (right)

A live diagram of the current store, with controls that re-render it on change:

| Control | Options |
|---------|---------|
| **Diagram Format** | `Ports` (list every port per device) or `Connections` (link-focused). |
| **VLAN Display** | Color by VLAN, or not. |
| **Color Ports with VLANs** | Color the port nodes too, or leave them plain. |
| **Show All Ports** | Include ports with no connection, or only connected ones. |
| **VLAN scope** | `Untagged only` (one line per link) or `All (incl. tagged)`. |
| **Only VLANs** | A comma-separated allow-list (e.g. `2,4,99`) — overrides VLAN scope. |

The image is fetched from the API's `/diagram` endpoint; if the store has no
devices the pane shows an "empty" placeholder. (The CLI renders the same diagram,
including ASCII output — see [CLI Usage §7](usage-cli.md#7-diagrams-svg--ascii).)

---

## 4. Import devices — scan → configure → import

The **Import devices** page (`import.php` → `action/importscan.php`) is the
guided scan-to-import workflow. It mirrors the CLI `scan run`/`scan import`
round-trip (see [CLI Usage §4](usage-cli.md#4-scanning-devices-scan--edit--import)),
but with editable forms instead of a JSON file.

### a. (Optional) load or create a profile

- **Load into form** — pick a saved [scan profile](#6-the-credential-vault) to
  pre-fill the scan parameters; existing profiles can be deleted here too.
- **Create profile** — four one-click types set up the right fields:
  **Device · SNMP**, **Device · SSH**, **Generic · SNMP**, **Generic · SSH**
  (Generic = reusable credentials with no host). SSH passwords are encrypted by
  the credential vault.

### b. Live scan

In the **Live scan** box:

1. **Method** — `snmp` or `ssh` (the form toggles the relevant fields).
2. **Target** — a single IP, a CIDR, or a comma-separated list
   (e.g. `10.0.2.245` or `10.0.2.0/26`).
3. **SNMP**: community/version/port; **SSH**: choose an SSH profile for the
   credentials and os-type.
4. **Auto-import** (checkbox) — import discovered devices immediately, skipping
   the per-device review.
5. **Scan** — runs the scan **asynchronously**. A *Scanning…* spinner polls the
   API for progress (`scan-status.js`) and reloads when complete. **Clean scan**
   clears the current results.

### c. Discovered devices → configure each one

Completed scans list the **discovered devices**. For each, **Configure & import
→** analyzes it into an editable plan:

- A per-interface table shows the proposed **IP → VLAN** mappings, each with a
  **confidence/reason**. Edit the **subnet** and **VLAN id** cells to correct the
  heuristics.
- **Import device** executes the (edited) plan: it creates the brand/model/zone
  as needed and adds the device, its ports, and VLANs.

Already-imported devices are marked so you don't import twice.

### d. Upload a scan-result JSON

The **Upload scan-result JSON** box imports a file produced by the CLI
(`nsl-graph scan ... > file.json`) — the same format `scan import <file>` accepts
(a raw `ScanResult` object or an edited import-plan array). This is the bridge
between a CLI scan and a UI import.

---

## 5. Scan connections — discover L2/L1 links

The **Scan connections** page (`connections.php` → `action/connections.php`)
discovers physical/L2 links between known hosts. It is **observe-only**: it reads
LLDP/CDP/bridge tables but never configures the targets (these must already be
enabled). It mirrors the CLI `scan connections`
([CLI Usage §6](usage-cli.md#6-discovering-connections)).

### a. Run discovery

In **Run discovery**:

- **Targets** — which hosts to query (DB devices, a swept subnet, or saved
  profiles).
- **Collector** — leave blank to use **all sources merged** (LLDP via SNMP/SSH,
  CDP, bridge FDB), or pick exactly one (`snmp-lldp`, `ssh-lldp`, `snmp-fdb`, …).
- **SSH credentials** — for hosts without a device profile (inline user/key, or a
  generic profile via the vault).
- **Discover** — runs the gather.

### b. Review the gather

After discovery you see:

- **Gather (N hosts)** — what each host reported. Hosts lacking a scan profile can
  be given one inline with **Assign**.
- **Intermediary devices detected in the middle** — transparent switches that
  don't speak LLDP/SNMP. You can **import a placeholder unmanaged device** so the
  link still shows a node in the middle.
- **Discovered topology** — a preview of the inferred graph.
- **Derived edges** — the candidate links in a table (**Import · DB status · Mark
  · From · To · Via · Discard**). Untick or **Discard** any edge you don't want
  (discarded rows are grayed and moved to the end). **Import selected
  connections** commits the ticked edges to the DB.
- **Discrepancies** — conflicting or unresolved evidence is surfaced here for you
  to resolve before importing.

---

## 6. The credential vault

SSH passwords and uploaded private keys (in scan profiles) are encrypted at rest
by the **credential vault** (AES-256-GCM under a random data key, which is itself
wrapped by your master passphrase). The vault control in the app bar shows its
state:

- **First use** — you set a master passphrase.
- **Locked** — click **Unlock** and enter the passphrase to let scans use stored
  SSH secrets; the data key is held in server memory only while unlocked.
- **Lock** — drops the in-memory key.

SNMP-only scans never need the vault. The vault is the same one the CLI uses, so a
profile created in the UI works from the CLI and vice-versa.

---

## See also

- **[CLI Usage Guide](usage-cli.md)** — the same workflows from the command line.
- **[Frontend API](frontend-api.md)** — the HTTP endpoints each page calls.
- **[Glossary](glossary.md)** — canonical terms.
