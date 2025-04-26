// Package domain set ups the interface for implementations for operations with the repository
package domain

import e "nsl-graph/internal/repository/entities"

// Comment
type Repository interface {
	// Brand
	AddBrand(brand string) error
	// Brand
	GetBrands() ([]e.Brand, error)
	// deviceclass
	AddDeviceClass(devclass string) error
	// deviceclass
	GetDeviceClasses() ([]e.DevClass, error)
	// deviceclass
	AddZoneType(name string) error
	// deviceclass
	GetZonetypes() ([]e.ZoneType, error)
	// deviceclass
	AddProprietary(name string) error
	// deviceclass
	GetProperties() ([]e.Proprietary, error)
	// deviceclass
	AddZone(
		name string,
		fatherid string,
		father string,
		proprietary string,
		zonename string,
	) error
	// deviceclass
	GetZones() ([]e.Zone, error)
	// GetZone(name string) int
	AddModel(
		modelName string,
		brandName string,
		className string,
	) error

	GetModels() ([]e.ModelDevice, error)
	GetDevices() ([]e.Device, error)
	AddDevice(
		label string,
		model string,
		zoneId string,
		zoneName string,
		proprietary string,
	) error
	GetModelPorts() ([]e.ModelPort, error)

	AddModelPort(
		name string,
		posx string,
		posy string,
		modelName string,
	) error
	GetDevicePorts() ([]e.DevicePort, error)
	AddDevicePort(deviceid string, modelportid string) error
	GetConnections() ([]e.Connection, error)
	AddConnection(
		fromDevice string,
		fromModelPort string,
		toDevice string,
		toModelPort string,
	) error
	GetAllPortsDevice(deviceid string) ([]e.DevicePort, error)
	GetAllPortsAll() ([]e.DevicePort, error)
	ExportAllStructs() (e.All, error)
}

type NetRepository Repository
