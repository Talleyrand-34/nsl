 --  Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)
 -- 
 --  This program is free software: you can redistribute it and/or modify
 --  it under the terms of the GNU Affero General Public License as published
 --  by the Free Software Foundation, either version 3 of the License, or
 --  (at your option) any later version.
 -- 
 --  This program is distributed in the hope that it will be useful,
 --  but WITHOUT ANY WARRANTY; without even the implied warranty of
 --  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 --  GNU Affero General Public License for more details.
 -- 
 --  You should have received a copy of the GNU Affero General Public License
 --  along with this program. If not, see <https://www.gnu.org/licenses/>.

PRAGMA foreign_keys = ON;

-- 1. Device Classes
CREATE TABLE DeviceClass (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR NOT NULL UNIQUE
);

-- 2. Brand
CREATE TABLE Brand (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    brand VARCHAR NOT NULL UNIQUE
);

-- 3. Proprietary
CREATE TABLE Proprietary (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    proprietary VARCHAR NOT NULL UNIQUE
);

-- 4. ZoneType
CREATE TABLE ZoneType (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location_type VARCHAR NOT NULL UNIQUE
);

-- 5. Zone
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

-- 6. ModelDevice
CREATE TABLE ModelDevice (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model VARCHAR NOT NULL,
    brand INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    FOREIGN KEY (brand) REFERENCES Brand (id),
    FOREIGN KEY (class_id) REFERENCES DeviceClass (id)
);

-- 7. ModelPort
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

-- 8. Device
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

-- 9. DevicePort
CREATE TABLE DevicePort (
    model_port_id INTEGER NOT NULL,
    device_id INTEGER NOT NULL,
    mac_address VARCHAR,
    PRIMARY KEY (model_port_id, device_id),
    FOREIGN KEY (model_port_id) REFERENCES ModelPort (id),
    FOREIGN KEY (device_id) REFERENCES Device (id)
);

-- 10. ConnectionType
CREATE TABLE ConnectionType (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    connection_type VARCHAR NOT NULL UNIQUE
);

-- 11. Connection
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

-- 12. Policy
CREATE TABLE Policy (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL, 
    associated_connection INTEGER DEFAULT NULL,
    TODO VARCHAR DEFAULT NULL,
    FOREIGN KEY (associated_connection) REFERENCES Connection (id)
);

-- 13. View: DevicePortValidation
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
