-- name: GetBrands :many
SELECT brand
FROM brand;

-- name: AddBrand :exec
INSERT INTO brand (
	brand
) VALUES (
	?
);
	
-- name: GetDeviceClasses :many
SELECT name
FROM DeviceClass;

-- name: AddDeviceClass :exec
INSERT INTO DeviceClass(
	name
) VALUES (
  ?	
);

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

-- name: GetZones :many
SELECT z1.name,z2.name as father,Zonetype.location_type,proprietary.proprietary
FROM Zone z1 JOIN proprietary JOIN ZoneType JOIN Zone z2 on z1.father=z2.id;

-- name: GetZoneId :one
SELECT id
FROM Zone
WHERE name=? LIMIT 1;

-- name: AddZone :exec
INSERT INTO Zone(
name,father,location_type,proprietary
) VALUES (
?,?,?,?	
);

