# Frontend API Documentation

NSL-Graph provides both a PHP web interface and HTTP REST API for managing network data. The system runs on two services that work together.

## Architecture

### Services
1. **Go HTTP API Server**: Backend API service (default port 8081)
2. **PHP Web Interface**: Frontend interface (default port 8091)

### Startup
Use the provided script to start both services:
```bash
./run-server.sh
```

This starts:
- Go API server: `http://localhost:8081`
- PHP frontend: `http://localhost:8091`

## PHP Web Interface

### Access
Open your browser to: `http://localhost:8091`

### Features
The web interface provides:
- **Dynamic Entity Management**: Add/view network entities through dropdown menus
- **Live Network Diagram**: Real-time SVG diagram that updates as you modify data
- **Two-Panel Layout**: 
  - Left: Action forms and data display
  - Right: Resizable network diagram viewer

### Available Actions

#### Action Types
- **Get**: Retrieve and display existing data
- **Add**: Create new network entities

#### Supported Entities
| Entity | Get | Add | Description |
|--------|-----|-----|-------------|
| Brand | ✓ | ✓ | Equipment manufacturers |
| Device Class | ✓ | ✓ | Equipment types (router, switch, etc.) |
| Proprietary | ✓ | ✓ | Ownership entities |
| Zone Type | ✓ | ✓ | Zone classifications |
| Zone | ✓ | ✓ | Network zones |
| Model | ✓ | ✓ | Device models |
| Device | ✓ | ✓ | Device instances |
| Model Port | ✓ | ✓ | Port definitions |
| Connections | ✓ | ✓ | Network connections |
| Connection Type | ✓ | ✓ | Connection classifications |

### Interface Workflow
1. Select action type (Get/Add)
2. Choose entity type
3. Form automatically loads for selected combination
4. Submit data or view results
5. Network diagram updates automatically

## HTTP REST API

### Base URL
```
http://localhost:8081
```

### Authentication
No authentication required (development/local use).

### Content Type
- **Request**: `application/json` (for POST/PUT)
- **Response**: `application/json`

### Endpoints

#### Root Endpoint
```http
GET /
```
Returns list of all available endpoints.

#### Brands
```http
GET    /brands
POST   /brands
DELETE /brands
```

**POST Example:**
```bash
curl -X POST http://localhost:8081/brands \
  -H "Content-Type: application/json" \
  -d '{"brand": "Cisco"}'
```

**DELETE Example:**
```bash
curl -X DELETE "http://localhost:8081/brands?brand=Cisco"
```

#### Device Classes
```http
GET    /deviceclasses
POST   /deviceclasses
DELETE /deviceclasses
```

**POST Example:**
```bash
curl -X POST http://localhost:8081/deviceclasses \
  -H "Content-Type: application/json" \
  -d '{"name": "router"}'
```

#### Proprietaries
```http
GET    /proprietaries
POST   /proprietaries
DELETE /proprietaries
```

#### Zone Types
```http
GET    /zonetypes
POST   /zonetypes
DELETE /zonetypes
```

#### Zones
```http
GET    /zones
POST   /zones
DELETE /zones
```

**POST Example:**
```bash
curl -X POST http://localhost:8081/zones \
  -H "Content-Type: application/json" \
  -d '{
    "name": "datacenter",
    "zonetype": "physical",
    "proprietary": "MyCompany"
  }'
```

#### Models
```http
GET    /models
POST   /models
DELETE /models
```

**POST Example:**
```bash
curl -X POST http://localhost:8081/models \
  -H "Content-Type: application/json" \
  -d '{
    "model": "ISR4431",
    "brand": "Cisco",
    "class": "router"
  }'
```

#### Devices
```http
GET    /devices
POST   /devices
PUT    /devices
DELETE /devices
```

**POST Example** (create). `model_name`, `zone_id`/`zone_name`, and
`proprietary` are required-ish; `ips` and `profile` are optional. `profile` is a
scan-profile **name** (device or generic) the device is associated with — also set
automatically when a device is imported from a scan (see *Glossary › Profile*):
```bash
curl -X POST http://localhost:8081/devices \
  -H "Content-Type: application/json" \
  -d '{
    "label": "router01",
    "model_name": "ISR4431",
    "zone_id": "<zone-id>",
    "zone_name": "datacenter",
    "proprietary": "MyCompany",
    "ips": ["10.0.0.1"],
    "profile": "gen-ssh"
  }'
```

**PUT Example** (partial update; omitted fields are left unchanged). Use
`is_unmanaged` to toggle managed/unmanaged VLAN mode; `profile` is nullable —
omit it to leave unchanged, set `""` to clear. A bare `{id, profile}` PUT just
(re)assigns the scan profile:
```bash
curl -X PUT http://localhost:8081/devices \
  -H "Content-Type: application/json" \
  -d '{
    "id": "<device-id>",
    "is_unmanaged": true,
    "profile": "dev-snmp"
  }'
```

#### Model Ports
```http
GET    /modelports
POST   /modelports
DELETE /modelports
```

**POST Example:**
```bash
curl -X POST http://localhost:8081/modelports \
  -H "Content-Type: application/json" \
  -d '{
    "name": "GigE0/0/0",
    "posx": "0",
    "posy": "0",
    "modelname": "ISR4431"
  }'
```

#### Device Ports
```http
GET    /deviceports
POST   /deviceports
DELETE /deviceports
```

**POST Example:**
```bash
curl -X POST http://localhost:8081/deviceports \
  -H "Content-Type: application/json" \
  -d '{
    "deviceid": "1",
    "modelportid": "1"
  }'
```

#### Connections
```http
GET    /connections
POST   /connections
DELETE /connections
PUT    /connections
```

**POST Example** (`connection_type` is required and must already exist; create one
via `POST /connectiontypes`):
```bash
curl -X POST http://localhost:8081/connections \
  -H "Content-Type: application/json" \
  -d '{
    "from_deviceport_id": "<deviceport-id>",
    "to_deviceport_id": "<deviceport-id>",
    "connection_type": "ethernet"
  }'
```

**PUT Example:**
```bash
curl -X PUT http://localhost:8081/connections \
  -H "Content-Type: application/json" \
  -d '{
    "id": "1",
    "from_device": "1",
    "from_port": "2",
    "to_device": "3",
    "to_port": "1"
  }'
```

#### Connection Types
```http
GET    /connectiontypes
POST   /connectiontypes
```

#### Special Endpoints

##### All Ports for Device
```http
GET /allports/device?deviceid=<id>
```
Returns all ports associated with a specific device.

##### All Ports (All Devices)
```http
GET /allports/all
```
Returns all device ports across all devices.

##### Network Diagram
```http
GET /diagram
```
Returns SVG network diagram.

**Example:**
```bash
curl http://localhost:8081/diagram > network.svg
```

#### Network Scanning

```http
POST /scan/host          {"ip": "...", "community": "...", "snmp_version": "...", "snmp_port": 0, "profile": "..."}
POST /scan/network       {"subnet": "...", "community": "...", ..., "profile": "..."}
POST /scan/host-ssh      {"profile": "...", "ip": "...(optional override)"}
POST /scan/analyze       {"device": <DiscoveredDevice>}
POST /scan/execute       {"plan": <DeviceImportPlan>, "options": {...}}
POST /scan/import        {"devices": [...], "options": {...}}
POST /scan/import-file   <raw scanner.ScanResult JSON>
```

**Connection (L2/L1 link) discovery:**

```http
POST /scan/connections              <ConnectionScanOptions>   (async; returns {scan_id})
POST /scan/connections/import       {"edges": [<ConnectionEdge>...]}
POST /scan/connections/placeholders {"intermediaries": [<Intermediary>...]}
```

`/scan/connections` correlates LLDP/CDP/bridge-FDB evidence into edges and detects
**intermediaries** — MACs seen by 2+ hosts that don't speak LLDP and aren't in the DB
(unknown device(s) "in the middle").

In **from-db** mode every device must have an associated scan **profile** (its own
`device.profile`, else one auto-matched by host). Devices without one are excluded
and reported as a `device-no-profile` discrepancy (the device labels in
`provenance`); the web review then offers a per-device profile dropdown that
assigns it (`PUT /devices {id, profile}`) so you can re-run.

`/scan/connections/placeholders` materializes a
single shared **placeholder unmanaged device** (in an "Unknown infrastructure" zone)
for the selected intermediaries and wires each observing `device:port` endpoint to it,
for VLAN attestation/documentation of the gap. Returns
`{zone, device, connections, message?}`.

The optional **`profile`** field on `/scan/host` and `/scan/network` applies a saved scan
profile (see below). When omitted, a profile whose `host` matches the `ip`/`subnet` is
auto-applied. Any SNMP field present in the request overrides the profile. The SNMP scan path
never uses the SSH password.

**SSH scan + interactive VLAN import** (used by the web UI's editable flow):

- `POST /scan/host-ssh` — scans a host over **SSH/config**. Credentials come from a saved
  profile (it must set `device_type`, `ssh_user`, and an encrypted `ssh_password`); the
  profile's secret is decrypted by the **credential vault**, which must be unlocked (see
  *Credential vault* below). Returns a `DiscoveredDevice`. Errors: `404` unknown profile,
  `400` vault locked / missing `device_type`.
- `POST /scan/analyze` — returns the `DeviceImportPlan` for a discovered device: per-interface
  **IP / ip-segment(subnet) / VLAN-id** mappings (`ip_mappings`) plus suggested VLAN create/update
  plans, with a confidence/reason per mapping.
- `POST /scan/execute` — imports a (possibly **edited**) plan. The server recomputes the VLAN
  create/update plans from the final `ip_mappings`, so edits to vlan-id / ip-segment are applied
  consistently, then creates the device, ports, interfaces and VLANs.

#### Scan Profiles

Reusable per-host scan parameters. `GET` never returns the SSH password — only a
`has_ssh_password` boolean.

```http
GET    /scan/profiles                      # list (no passwords)
POST   /scan/profiles   {profile fields}   # create
PUT    /scan/profiles   {profile fields}   # update (omit ssh_password to keep existing)
DELETE /scan/profiles?name=<name>          # delete
```

To store an SSH password, unlock the **credential vault** (below) and send
`ssh_password`; the server encrypts it under the vault's data key (AES-256-GCM)
and stores only the ciphertext. Creating a profile with an SSH secret while the
vault is locked fails with `400 invalid_profile`.

#### Credential vault

A single server-side vault protects all stored SSH secrets. A master passphrase
unlocks a random **data key** (held in memory, wrapped at rest by the passphrase
via scrypt); while unlocked, profile secrets are encrypted/decrypted without
re-entering it. It auto-locks after an idle timeout and on server restart.

```http
GET  /vault/status   # {"initialized": bool, "unlocked": bool}
POST /vault/init     {"passphrase": "..."}   # set the master passphrase (fresh vault); leaves it unlocked. 409 if already initialized
POST /vault/unlock   {"passphrase": "..."}   # unlock. 401 on a wrong passphrase
POST /vault/lock                             # clear the in-memory data key
```

```bash
# create an SNMP profile
curl -X POST http://localhost:8081/scan/profiles \
  -H 'Content-Type: application/json' \
  -d '{"name":"ow","host":"10.0.2.245","snmp_community":"public"}'

# scan using it (auto-matched by ip)
curl -X POST http://localhost:8081/scan/host -d '{"ip":"10.0.2.245"}'
```

### Error Responses

The API returns standard HTTP status codes:

- **200**: Success
- **400**: Bad Request (invalid data)
- **404**: Not Found
- **500**: Internal Server Error

Error response format:
```json
{
  "error": "Error description",
  "details": "Additional error details"
}
```

### Data Formats

#### Brand
```json
{
  "id": 1,
  "brand": "Cisco"
}
```

#### Device Class
```json
{
  "id": 1,
  "name": "router"
}
```

#### Model Device
```json
{
  "id": 1,
  "model": "ISR4431",
  "brand": "Cisco",
  "class": "router"
}
```

#### Device
```json
{
  "id": 1,
  "label": "router01",
  "model": "ISR4431",
  "zone": "datacenter",
  "proprietary": "MyCompany"
}
```

#### Connection
```json
{
  "id": 1,
  "from_device": 1,
  "from_port": 1,
  "to_device": 2,
  "to_port": 1,
  "connection_type": "ethernet"
}
```

## PHP Action Files

The PHP frontend uses individual action files in the `frontend/php/action/` directory:

### Get Actions
- `getbrand.php`: Display all brands
- `getdevclass.php`: Display device classes
- `getproprietary.php`: Display proprietaries
- `getzonetype.php`: Display zone types
- `getzone.php`: Display zones
- `getmodeldevice.php`: Display models
- `getdevice.php`: Display devices
- `getmodelport.php`: Display model ports
- `getconnections.php`: Display connections
- `getconnectiontype.php`: Display connection types

### Add Actions
- `addbrand.php`: Add brand form
- `adddevclass.php`: Add device class form
- `addproprietary.php`: Add proprietary form
- `addzonetype.php`: Add zone type form
- `addzone.php`: Add zone form
- `addmodeldevice.php`: Add model form
- `adddevice.php`: Add device form
- `addmodelport.php`: Add model port form
- `addconnections.php`: Add connection form
- `addconnectiontype.php`: Add connection type form

## Configuration

### API Configuration
Edit `frontend/php/config.php` to modify:
- API base URL (default: `http://localhost:8081`)
- Individual endpoint URLs

### Server Ports
Modify `run-server.sh` to change:
- Go API port (default: 8081)
- PHP server port (default: 8091)

## Development

### Adding New Entities
1. Add API endpoint in Go backend
2. Create corresponding PHP action files:
   - `action/get{entity}.php`
   - `action/add{entity}.php`
3. Update `main.php` action mapping
4. Add endpoint constant in `config.php`

### Custom Styling
Modify CSS in `main.php` to customize:
- Grid layout
- Form styling
- Diagram container sizing
- Responsive behavior

## Troubleshooting

### Common Issues

**Cannot connect to API:**
- Verify Go server is running on port 8081
- Check `config.php` API_BASE_URL setting

**Diagram not loading:**
- Ensure `/diagram` endpoint is accessible
- Check browser console for errors

**Form submissions failing:**
- Verify POST data format matches API expectations
- Check API server logs for errors

**PHP server issues:**
- Ensure PHP is installed and accessible
- Verify port 8091 is not in use