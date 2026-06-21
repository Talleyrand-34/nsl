# Database Schema

NSL-Graph uses **CloverDB** as its database backend - a document-based (NoSQL) store. Documents are stored as JSON with automatic ID assignment and query capabilities.

## Database Structure

The database consists of **14 collections** organized to represent a complete network topology:

- **Base entities**: brands, devclasses, proprietaries, zonetypes
- **VLAN definitions**: vlans
- **Zone hierarchy**: zones
- **Model/Template collections**: models, modelports
- **Device collections**: devices, deviceports, deviceinterfaces, interfaceports
- **Connection collections**: connections, connectiontypes

## Schema Definition

The complete database schema is documented in DBML format:

**See: [database-schema.dbml](./database-schema.dbml)**

The DBML file contains:
- All collection definitions with field names and types
- Relationship references between collections
- Notes explaining the purpose of each collection
- Nested array field definitions for complex types

## Document Structure Examples

### DevicePort (Physical port with VLANs)
```json
{
  "_id": "abc123",
  "device_id": "device-uuid",
  "model_port_id": "port-uuid",
  "mac_address": "00:e0:b4:60:e9:b1",
  "vlan_configs": [
    {"vlan_number": "1", "tagged": false},
    {"vlan_number": "10", "tagged": true}
  ]
}
```

### DeviceInterface (Logical interface)
```json
{
  "_id": "iface-uuid",
  "device_id": "device-uuid",
  "name": "eth0.1",
  "description": "VLAN 1 subinterface",
  "parent": "eth0",
  "vlan_configs": [
    {"vlan_number": "1", "tagged": false}
  ],
  "ip_addresses": ["10.0.1.1/24"],
  "wifi_ssid": "",
  "wifi_security": ""
}
```

### Connection (with VLAN info)
```json
{
  "_id": "conn-uuid",
  "from_device_id": "device1-uuid",
  "from_model_port_id": "port1-uuid",
  "to_device_id": "device2-uuid",
  "to_model_port_id": "port2-uuid",
  "vlans": [
    {"vlan_id": "1", "tagged": false}
  ],
  "missing_vlans": [
    {"vlan_id": "2", "tagged": false}
  ]
}
```

**Note**: The `vlans` and `missing_vlans` fields on Connection are computed at retrieval time by intersecting VLAN memberships from both port ends.

## Key Relationships

```
devices ─────┬──── models (device instance of model)
             │
             ├──── zones (device in zone)
             │
             ├──── deviceports (physical ports with VLANs)
             │       └──── modelports (port template)
             │
             ├──── deviceinterfaces (logical interfaces)
             │       └──── interfaceports (interface to port mapping)
             │
             └──── localvlans (device-specific VLAN names)

connections ──┴──── deviceports (from/to port connections)

## Data Integrity

- **Referential Integrity**: Enforced at application layer through repository pattern
- **Unique Constraints**: Collection-based uniqueness on key fields
- **Composite Keys**: Junction tables use combined keys (e.g., deviceports uses device_id + model_port_id)
- **Validation**: Application layer validates data before storage

## Database Files

- **Location**: Specified via `-s` or `--source` flag (default: `test.db`)
- **Format**: CloverDB document store (directory with data files)
- **Encoding**: UTF-8
- **Auto-creation**: Database is created automatically on first access

## Entity Relationships

### One-to-Many
- Brand → Models
- DeviceClass → Models
- Model → ModelPorts
- Model → Devices
- Zone → Sub-zones (self-reference)
- Zone → Devices
- Device → DevicePorts
- Device → DeviceInterfaces

### Many-to-Many
- DevicePorts ↔ Connections (via from/to references)
- DeviceInterfaces ↔ DevicePorts (via InterfacePorts join table)

> **Note:** `interface_ports` is a *pure join* table. VLAN configurations and IP
> addresses are owned exclusively by `device_interfaces` and are resolved via
> `interface_id`; they are not duplicated on the join. (`device_id` is kept on the
> join only as a denormalized query key, since `model_port_id` is shared across
> devices of the same model.)
