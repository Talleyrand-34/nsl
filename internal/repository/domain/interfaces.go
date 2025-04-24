// Package domain set ups the interface for implementations for operations with the repository
package domain

import e "nsl-graph/internal/repository/entities"

// Comment
type NetRepository interface {
	// Brand
	AddBrand(brand string) error
	// Brand
	GetBrands() []e.Brand
	// deviceclass
	AddDeviceClass(devclass string) error
	// deviceclass
	GetDeviceClasses() []e.DevClass
	// deviceclass
	AddZoneType(name string) error
	// deviceclass
	GetZonetypes() []e.ZoneType
	// deviceclass
	AddProprietary(name string) error
	// deviceclass
	GetProperties() []e.Proprietary
	// deviceclass
	AddZone(
		name string,
		fatherid string,
		father string,
		proprietary string,
		zonename string,
	) error
	// deviceclass
	GetZones() []e.Zone
	// GetZone(name string) int
	AddModel(
		modelName string,
		brandName string,
		className string,
	) error

	GetModels() []e.ModelDevice
	GetDevices() []e.Device
	AddDevice(
		label string,
		model string,
		zoneId string,
		zoneName string,
		proprietary string,
	) error
	GetModelPorts() []e.ModelPort

	AddModelPort(
		name string,
		posx string,
		posy string,
		modelName string,
	) error
	GetDevicePorts() []e.DevicePort
	AddDevicePort(deviceid string, modelportid string) error
	GetConnections() []e.Connection
	AddConnection(
		fromDevice string,
		fromModelPort string,
		toDevice string,
		toModelPort string,
	) error
}
