---- Getters and setters for each table

--- Brand

-- name: GetBrands :many
SELECT brand
FROM brand;

-- name: AddBrand :exec
INSERT INTO brand (
	brand
) VALUES (
	?
);

-- name: GetBrandId :one
SELECT id FROM Brand WHERE brand = ? LIMIT 1;
	
-- name: DeleteBrand :exec
DELETE FROM Brand WHERE brand = ?;
--- DeviceClass

-- name: GetDeviceClasses :many
SELECT name
FROM DeviceClass;

-- name: AddDeviceClass :exec
INSERT INTO DeviceClass(
	name
) VALUES (
  ?	
);

-- name: GetClassId :one
SELECT id FROM DeviceClass WHERE name = ? LIMIT 1;

-- name: DeleteDeviceClass :exec
DELETE FROM DeviceClass WHERE name = ?;

--- ZoneTypes

-- name: GetZoneTypes :many
SELECT location_type
FROM Zonetype;

-- name: GetZoneType :one
SELECT id
FROM Zonetype
WHERE location_type=? LIMIT 1;

-- name: AddZoneType :exec
INSERT INTO Zonetype(
	location_type
) VALUES (
  ?	
);

-- name: DeleteZoneType :exec
DELETE FROM Zonetype WHERE location_type = ?;

--- Proprietary

-- name: GetProprietaries :many
SELECT proprietary 
FROM Proprietary;

-- name: GetProprietary :one
SELECT id 
FROM Proprietary
WHERE proprietary=? LIMIT 1;
-- name: AddProprietary :exec
INSERT INTO Proprietary(
	proprietary 
) VALUES (
  ?	
);

-- name: DeleteProprietary :exec
DELETE FROM Proprietary WHERE proprietary = ?;

--- Zone

-- name: GetZones :many
SELECT z1.id as id,z1.name as name,z2.id as fatherid,z2.name as father,Zonetype.location_type,proprietary.proprietary
FROM Zone z1 JOIN proprietary JOIN ZoneType LEFT JOIN Zone z2 on z1.father=z2.id;

-- name: GetZoneId :one
SELECT id
FROM Zone
WHERE name=? LIMIT 1;

-- name: CountZones :one
SELECT count(*)
FROM Zone
WHERE name=?;

-- name: AddZone :exec
INSERT INTO Zone(
name,father,location_type,proprietary
) VALUES (
?,?,?,?	
);

-- name: DeleteZone :exec
DELETE FROM Zone WHERE id = ?;

--- Model

-- name: GetModels :many
SELECT 
    m.id,
    m.model,
    b.brand,
    c.name AS class_name
FROM ModelDevice m
JOIN Brand b ON m.brand = b.id
JOIN DeviceClass c ON m.class_id = c.id;

-- name: GetModelId :one
SELECT id
FROM ModelDevice
WHERE model = ? LIMIT 1;

-- name: CountModels :one
SELECT count(*)
FROM ModelDevice
WHERE model = ?;

-- name: AddModel :exec
INSERT INTO ModelDevice(
    model,
    brand,
    class_id
) VALUES (
    ?,?,?
);

-- name: DeleteModel :exec
DELETE FROM ModelDevice WHERE id = ?;

--- Device

-- name: GetDevices :many
SELECT 
    d.id,
    d.label,
    md.model,
    b.brand,
    z.name as zonename,
    z.id as zoneid,
    z2.name as zonefathername,
    p.proprietary
FROM Device d 
LEFT JOIN ModelDevice md on d.model_id=md.id
LEFT JOIN Proprietary p on d.proprietary=p.id
LEFT JOIN Zone z on d.zone_id=z.id
LEFT JOIN Brand b on md.brand=b.id
LEFT JOIN Zone z2 on z.father=z2.id;

-- name: GetDeviceId :one
SELECT id
FROM Device
WHERE label = ? LIMIT 1;

-- name: CountDevices :one
SELECT count(*)
FROM Device
WHERE label = ?;

-- name: AddDevice :exec
INSERT INTO Device(
    label,
    model_id, 
    zone_id,
    proprietary
) VALUES (
    ?,?,?,?
);

-- name: DeleteDevice :exec
DELETE FROM Device WHERE id = ?;

--- ModelPort

-- name: GetModelPorts :many
SELECT 
    mp.id,
    mp.name,
    mp.positionx,
    mp.positiony,
    m.model,
    b.brand
FROM ModelPort mp
LEFT JOIN ModelDevice m on mp.model_id=m.id
LEFT JOIN brand b on m.brand=b.id;

-- name: GetModelPortId :one
SELECT mp.id
FROM ModelPort mp
LEFT JOIN ModelDevice m on mp.model_id=m.id
WHERE mp.name=? and m.model=? LIMIT 1;

-- name: CountModelPorts :one
SELECT count(*)
FROM ModelPort mp
LEFT JOIN ModelDevice m on mp.model_id=m.id
WHERE mp.name=? and m.model=? LIMIT 1;

-- name: AddModelPort :exec
INSERT INTO ModelPort(
    name,
    positionx,
    positiony,
    model_id
) VALUES (
    ?,?,?,?
);


-- name: DeleteModelPort :exec
DELETE FROM ModelPort WHERE id = ?;

--- DevicePort

-- name: GetDevicePorts :many
SELECT dp.device_id,dp.model_port_id,mp.name,d.label,mp.positionx,mp.positiony
FROM DevicePort dp
LEFT JOIN ModelPort mp on dp.model_port_id=mp.id 
LEFT JOIN Device d on dp.device_id=d.id 
;

-- name: AddDevicePort :exec
INSERT INTO DevicePort (
	device_id,model_port_id
) VALUES (
	?,?
);

-- name: DeleteDevicePort :exec
DELETE FROM DevicePort WHERE device_id = ? AND model_port_id = ?;

--- Connection

-- name: GetConnections :many
SELECT
    c.id,
    d.label as fromdevname,
    mp.name as frommodelportname,
    d2.label as todevname,
    mp2.name as tomodelportname,
    ct.connection_type,
    z.id as fromzoneid,
    z.name as fromzonename,
    z2.id as tozoneid,
    z2.name as tozonename
FROM Connection c
LEFT JOIN DevicePort dp on c.from_device_port_device_id=dp.device_id and c.from_device_port_model_port_id=dp.model_port_id
LEFT JOIN DevicePort dp2 on c.to_device_port_device_id=dp2.device_id and c.to_device_port_model_port_id=dp2.model_port_id
LEFT JOIN ModelPort mp on dp.model_port_id=mp.id 
LEFT JOIN Device d on dp.device_id=d.id 
LEFT JOIN ModelPort mp2 on dp2.model_port_id=mp2.id 
LEFT JOIN Device d2 on dp2.device_id=d2.id 
LEFT JOIN ConnectionType ct on c.connection_type=ct.id
LEFT JOIN Zone z on d.zone_id=z.id
LEFT JOIN Zone z2 on d2.zone_id=z2.id

;

-- name: AddConnection :exec
INSERT INTO Connection (
	from_device_port_device_id,
	from_device_port_model_port_id,
	to_device_port_device_id,
	to_device_port_model_port_id,
	connection_type
) VALUES (
	?,?,?,?,?
);

-- name: GetConnectionId :many
SELECT 
    id,
    from_device_port_device_id,
    from_device_port_model_port_id,
    to_device_port_device_id,
    to_device_port_model_port_id,
    connection_type
FROM Connection
WHERE 
    from_device_port_device_id = ? AND
    from_device_port_model_port_id = ? AND
    to_device_port_device_id = ? AND
    to_device_port_model_port_id = ?
;

-- name: UpdateConnection :exec
UPDATE connection
SET
    from_device_port_device_id = ?,
    from_device_port_model_port_id = ?,
    to_device_port_device_id = ?,
    to_device_port_model_port_id = ?
WHERE
    id = ?;
-- name: DeleteConnection :exec
DELETE FROM Connection 
WHERE
    id = ?;
---- Special 


-- name: GetPossiblePortsAll :many
SELECT Device.id as deviceid, ModelPort.id as modelid
FROM Device
CROSS JOIN ModelPort;


-- name: GetPossiblePortsDevice :many
SELECT Device.id as deviceid, ModelPort.id as modelid
FROM Device
CROSS JOIN ModelPort
WHERE Device.id = ?;



--- Basic getter


-- name: BasicGetDeviceClasses :many
SELECT id, name
FROM DeviceClass;

-- name: BasicGetBrands :many
SELECT id, brand
FROM Brand;

-- name: BasicGetProprietaries :many
SELECT id, proprietary
FROM Proprietary;

-- name: BasicGetZoneTypes :many
SELECT id, location_type
FROM ZoneType;

-- name: BasicGetZones :many
SELECT id, name, father, granularity, proprietary, location_type
FROM Zone;

-- name: BasicGetModelDevices :many
SELECT id, model, brand, class_id
FROM ModelDevice;

-- name: BasicGetModelPorts :many
SELECT id, name, positionx, positiony, model_id
FROM ModelPort;

-- name: BasicGetDevices :many
SELECT id, label, model_id, zone_id, proprietary
FROM Device;

-- name: BasicGetDevicePorts :many
SELECT model_port_id, device_id
FROM DevicePort;

-- name: BasicGetConnectionTypes :many
SELECT id, connection_type
FROM ConnectionType;

-- name: BasicGetConnections :many
SELECT id, from_device_port_model_port_id, from_device_port_device_id, from_ip_segment, to_device_port_model_port_id, to_device_port_device_id, to_ip_segment, connection_type
FROM Connection;

-- name: BasicGetPolicies :many
SELECT id, name, description, associated_connection, TODO
FROM Policy;
























