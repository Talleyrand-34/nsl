# Database Schema

NSL-Graph uses SQLite as its database backend. The schema defines network infrastructure entities and their relationships.

## Database Structure

The database consists of 12 main tables and 1 view, organized to represent a complete network topology.

### Core Entity Tables

#### 1. Brand
Stores equipment manufacturers.

```sql
CREATE TABLE Brand (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    brand VARCHAR NOT NULL UNIQUE
);
```

**Purpose**: Track equipment vendors (Cisco, Juniper, HP, etc.)

#### 2. DeviceClass  
Defines types of network equipment.

```sql
CREATE TABLE DeviceClass (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR NOT NULL UNIQUE
);
```

**Purpose**: Categorize equipment types (router, switch, firewall, server, etc.)

#### 3. Proprietary
Defines ownership or management entities.

```sql
CREATE TABLE Proprietary (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    proprietary VARCHAR NOT NULL UNIQUE
);
```

**Purpose**: Track who owns/manages different network segments

#### 4. ZoneType
Defines zone categorization.

```sql
CREATE TABLE ZoneType (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location_type VARCHAR NOT NULL UNIQUE
);
```

**Purpose**: Classify zones (physical, logical, security, management, etc.)

### Hierarchical Tables

#### 5. Zone
Represents network segments with hierarchical relationships.

```sql
CREATE TABLE Zone (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR NOT NULL,
    father INTEGER DEFAULT NULL,
    granularity INTEGER DEFAULT NULL, 
    proprietary INTEGER DEFAULT NULL,
    location_type INTEGER DEFAULT NULL,
    FOREIGN KEY (father) REFERENCES Zone (id),
    FOREIGN KEY (proprietary) REFERENCES Proprietary (id),
    FOREIGN KEY (location_type) REFERENCES ZoneType (id)
);
```

**Relationships**:
- Self-referencing: Zones can contain sub-zones
- Links to ZoneType for classification
- Links to Proprietary for ownership

**Purpose**: Model network hierarchy (datacenter → building → floor → rack)

#### 6. ModelDevice
Device models with specifications.

```sql
CREATE TABLE ModelDevice (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model VARCHAR NOT NULL,
    brand INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    FOREIGN KEY (brand) REFERENCES Brand (id),
    FOREIGN KEY (class_id) REFERENCES DeviceClass (id)
);
```

**Relationships**:
- Links to Brand (manufacturer)
- Links to DeviceClass (equipment type)

**Purpose**: Define specific equipment models (Cisco ISR4431, Juniper EX4300, etc.)

### Port and Device Tables

#### 7. ModelPort
Defines port layouts for device models.

```sql
CREATE TABLE ModelPort (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR NOT NULL,
    positionx INTEGER NOT NULL,
    positiony INTEGER NOT NULL,
    model_id INTEGER NOT NULL,
    FOREIGN KEY (model_id) REFERENCES ModelDevice (id),
    UNIQUE (positionx, positiony, model_id),
    UNIQUE (name, model_id)
);
```

**Constraints**:
- Each position (x,y) is unique per model
- Each port name is unique per model

**Purpose**: Define physical port layouts for equipment models

#### 8. Device
Individual equipment instances.

```sql
CREATE TABLE Device (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    label VARCHAR NOT NULL,
    model_id INTEGER NOT NULL,
    zone_id INTEGER DEFAULT NULL,
    proprietary INTEGER DEFAULT NULL,
    FOREIGN KEY (model_id) REFERENCES ModelDevice (id),
    FOREIGN KEY (zone_id) REFERENCES Zone (id),
    FOREIGN KEY (proprietary) REFERENCES Proprietary (id)
);
```

**Relationships**:
- Links to ModelDevice (equipment model)
- Links to Zone (physical/logical location)
- Links to Proprietary (ownership)

**Purpose**: Represent actual equipment instances

#### 9. DevicePort
Associates model ports with device instances.

```sql
CREATE TABLE DevicePort (
    model_port_id INTEGER NOT NULL,
    device_id INTEGER NOT NULL,
    PRIMARY KEY (model_port_id, device_id),
    FOREIGN KEY (model_port_id) REFERENCES ModelPort (id),
    FOREIGN KEY (device_id) REFERENCES Device (id)
);
```

**Composite Primary Key**: (model_port_id, device_id)

**Purpose**: Enable/activate specific ports on device instances

### Connection Tables

#### 10. ConnectionType
Categorizes connection types.

```sql
CREATE TABLE ConnectionType (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    connection_type VARCHAR NOT NULL UNIQUE
);
```

**Purpose**: Define connection mediums (ethernet, fiber, wireless, etc.)

#### 11. Connection
Physical connections between device ports.

```sql
CREATE TABLE Connection (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_device_port_model_port_id INTEGER NOT NULL,
    from_device_port_device_id INTEGER NOT NULL,
    from_ip_segment VARCHAR,
    to_device_port_model_port_id INTEGER NOT NULL,
    to_device_port_device_id INTEGER NOT NULL,
    to_ip_segment VARCHAR,
    connection_type INTEGER,
    FOREIGN KEY (from_device_port_model_port_id, from_device_port_device_id) 
        REFERENCES DevicePort (model_port_id, device_id),
    FOREIGN KEY (to_device_port_model_port_id, to_device_port_device_id) 
        REFERENCES DevicePort (model_port_id, device_id),
    FOREIGN KEY (connection_type) REFERENCES ConnectionType (id)
);
```

**Complex Foreign Keys**: References DevicePort composite keys

**Purpose**: Model physical network connections with IP segment information

### Policy Table

#### 12. Policy
Network policies and rules (future enhancement).

```sql
CREATE TABLE Policy (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL, 
    associated_connection INTEGER DEFAULT NULL,
    TODO VARCHAR DEFAULT NULL,
    FOREIGN KEY (associated_connection) REFERENCES Connection (id)
);
```

**Purpose**: Define network policies associated with connections

### Views

#### DevicePortValidation
Validates device-port model compatibility.

```sql
CREATE VIEW DevicePortValidation AS
SELECT 
   dp.model_port_id AS ModelPortId, 
   dp.device_id AS DeviceId, 
   d.model_id AS DeviceModelId, 
   mp.model_id AS ModelPortDeviceModelId, 
   CASE 
       WHEN d.model_id = mp.model_id THEN 1 
       ELSE 0 
   END AS IsValidConfiguration
FROM DevicePort dp
JOIN Device d ON dp.device_id = d.id
JOIN ModelPort mp ON dp.model_port_id = mp.id;
```

**Purpose**: Ensure ports assigned to devices match the device's model

## Entity Relationships

### Key Relationships

1. **Brand → ModelDevice**: One-to-many (brands have multiple models)
2. **DeviceClass → ModelDevice**: One-to-many (classes have multiple models)
3. **ModelDevice → ModelPort**: One-to-many (models have multiple ports)
4. **ModelDevice → Device**: One-to-many (models have multiple instances)
5. **Zone → Zone**: Hierarchical self-reference (zones contain sub-zones)
6. **Zone → Device**: One-to-many (zones contain multiple devices)
7. **DevicePort ↔ Connection**: Many-to-many (ports can have multiple connections)

### Data Integrity

The schema enforces several integrity constraints:

1. **Referential Integrity**: All foreign keys are enforced
2. **Unique Constraints**: Prevent duplicate brands, device classes, etc.
3. **Composite Keys**: Ensure unique device-port associations
4. **Position Constraints**: Prevent overlapping port positions on models

### Example Data Flow

```
Brand: "Cisco"
    ↓
DeviceClass: "router" + Brand: "Cisco"
    ↓
ModelDevice: "ISR4431"
    ↓
ModelPort: "GigE0/0/0" (position 0,0)
    ↓
Device: "router01" (instance of ISR4431)
    ↓
DevicePort: Associates "router01" with "GigE0/0/0"
    ↓
Connection: Links two DevicePorts
```

## Database Operations

### Common Queries

#### Get all devices in a zone
```sql
SELECT d.label, md.model, b.brand
FROM Device d
JOIN ModelDevice md ON d.model_id = md.id
JOIN Brand b ON md.brand = b.id
WHERE d.zone_id = ?;
```

#### Get all connections for a device
```sql
SELECT c.id, c.from_ip_segment, c.to_ip_segment
FROM Connection c
WHERE c.from_device_port_device_id = ? 
   OR c.to_device_port_device_id = ?;
```

#### Validate port assignments
```sql
SELECT * FROM DevicePortValidation 
WHERE IsValidConfiguration = 0;
```

### Performance Considerations

- **Indexes**: Primary keys are automatically indexed
- **Foreign Keys**: Enable efficient joins
- **Composite Keys**: Optimize device-port lookups
- **View**: Pre-computed validation logic

## Database Files

- **Default**: `test.db`
- **Location**: Current working directory
- **Format**: SQLite 3
- **Encoding**: UTF-8
- **Foreign Keys**: Enabled (`PRAGMA foreign_keys = ON`)

The database file is created automatically when first accessed by the application.