-- Database Schema for NetworkSchema
PRAGMA foreign_keys = ON;

-- Table: DeviceClasses
CREATE TABLE DeviceClass (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR NOT NULL UNIQUE
);

-- Table: Brand
CREATE TABLE Brand (
    id INTEGER PRIMARY KEY,
    brand VARCHAR NOT NULL UNIQUE
);

-- Table: ModelDevices
CREATE TABLE ModelDevice (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model VARCHAR NOT NULL,
    brand INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    FOREIGN KEY (brand) REFERENCES Brand (id),
    FOREIGN KEY (class_id) REFERENCES DeviceClasses (id)
);

-- Table: ModelPorts
CREATE TABLE ModelPort (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR NOT NULL,
    positionx INTEGER NOT NULL,
    positiony INTEGER NOT NULL,
    model_id INTEGER NOT NULL,
    FOREIGN KEY (model_id) REFERENCES ModelDevices (id)
);

-- Table: Devices
CREATE TABLE Device (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    label VARCHAR NOT NULL,
    model_id INTEGER NOT NULL,
    zone_id INTEGER DEFAULT NULL,
    proprietary INTEGER DEFAULT NULL,
    FOREIGN KEY (model_id) REFERENCES ModelDevices (id),
    FOREIGN KEY (zone_id) REFERENCES Zone (id),
    FOREIGN KEY (proprietary) REFERENCES Proprietary (id)
);

-- Table: Proprietary
CREATE TABLE Proprietary(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    proprietary VARCHAR NOT NULL UNIQUE
);

-- Table: DevicePorts
CREATE TABLE DevicePort (
    model_port_id INTEGER NOT NULL,
    device_id INTEGER NOT NULL,
    PRIMARY KEY (model_port_id, device_id), -- Composite primary key
    FOREIGN KEY (model_port_id) REFERENCES ModelPort (id),
    FOREIGN KEY (device_id) REFERENCES Devices (id)
);

-- Table: Connections
CREATE TABLE Connection (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_device_port_model_port_id INTEGER NOT NULL,
    from_device_port_device_id INTEGER NOT NULL,
    from_ip_segment VARCHAR,
    to_device_port_model_port_id INTEGER NOT NULL,
    to_device_port_device_id INTEGER NOT NULL,
    to_ip_segment VARCHAR,
    connection_type INTEGER NOT NULL,
    FOREIGN KEY (from_device_port_model_port_id, from_device_port_device_id) REFERENCES DevicePort (model_port_id, device_id),
    FOREIGN KEY (to_device_port_model_port_id, to_device_port_device_id) REFERENCES DevicePort (model_port_id, device_id),
    FOREIGN KEY (connection_type) REFERENCES ConnectionType (id)
);

-- Table: ConnectionType
CREATE TABLE ConnectionType (
    id INTEGER PRIMARY KEY,
    connection_type VARCHAR NOT NULL UNIQUE
);

-- Table: Zone
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

-- Table: ZoneType
CREATE TABLE ZoneType (
    id INTEGER PRIMARY KEY,
    location_type VARCHAR NOT NULL UNIQUE
);

-- Table: Policies
CREATE TABLE Policy (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL, 
    associated_connection INTEGER DEFAULT NULL,
    TODO VARCHAR DEFAULT NULL,
    FOREIGN KEY (associated_connection) REFERENCES Connections (id)
);
-- View: Check Port Validity
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
JOIN Device d 
   ON dp.device_id = d.id
JOIN ModelPort mp 
   ON dp.model_port_id = mp.id;
