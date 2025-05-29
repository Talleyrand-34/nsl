// Package domain set ups the interface for implementations for operations with the repository
package domain

/*
  Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU Affero General Public License as published
  by the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  GNU Affero General Public License for more details.

  You should have received a copy of the GNU Affero General Public License
  along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

import e "nsl-graph/internal/repository/entities"

// Type repository is an interface for interaction with all the db logic needed from application
type repository interface {
	// Brand interaction
	AddBrand(brandName string) error
	GetBrands() ([]e.Brand, error)
	UpdateBrand(brandId string, newBrandName string) error
	DeleteBrand(brandName string) error

	// DeviceClass interaction
	AddDeviceClass(deviceClassName string) error
	GetDeviceClasses() ([]e.DevClass, error)
	UpdateDeviceClass(deviceClassId string, newDeviceClassName string) error
	DeleteDeviceClass(deviceClassName string) error

	// ZoneType interaction
	AddZoneType(zoneTypeName string) error
	GetZonetypes() ([]e.ZoneType, error)
	UpdateZoneType(zoneTypeId string, newZoneTypeName string) error
	DeleteZoneType(zoneTypeName string) error

	// Proprietary interaction
	AddProprietary(proprietaryName string) error
	GetProperties() ([]e.Proprietary, error)
	UpdateProprietary(proprietaryId string, newProprietaryName string) error
	DeleteProprietary(proprietaryName string) error

	// Zone interaction
	AddZone(
		zoneName string,
		fatherZoneId string,
		fatherZoneName string,
		proprietaryName string,
		zoneTypeName string,
	) error
	GetZones() ([]e.Zone, error)
	UpdateZone(
		zoneId string,
		newZoneName string,
		newFatherZoneId string,
		newZoneTypeId string,
		newProprietaryId string,
	) error
	DeleteZone(zoneId string) error

	// Model interaction
	AddModel(
		modelName string,
		brandName string,
		deviceClassName string,
	) error
	GetModels() ([]e.ModelDevice, error)
	UpdateModel(
		modelId string,
		newModelName string,
		newBrandId string,
		newDeviceClassId string,
	) error
	DeleteModel(modelId string) error

	// Device interaction
	AddDevice(
		deviceLabel string,
		modelName string,
		zoneId string,
		zoneName string,
		proprietaryName string,
	) error
	GetDevices() ([]e.Device, error)
	UpdateDevice(
		deviceId string,
		newDeviceLabel string,
		newModelId string,
		newZoneId string,
		newProprietaryId string,
	) error
	DeleteDevice(deviceId string) error

	// ModelPort interaction
	AddModelPort(
		portName string,
		positionX string,
		positionY string,
		modelName string,
	) error
	GetModelPorts() ([]e.ModelPort, error)
	UpdateModelPort(
		modelPortId string,
		newPortName string,
		newPositionX string,
		newPositionY string,
		newModelId string,
	) error
	DeleteModelPort(modelPortId string) error

	// DevicePort interaction
	AddDevicePort(deviceId string, modelPortId string, macAddress string) error
	GetDevicePorts() ([]e.DevicePort, error)
	DeleteDevicePort(deviceId string, modelPortId string) error

	// ConnectionType interaction
	AddConnectionType(connectionTypeName string) error
	GetConnectionTypes() ([]e.ConnectionType, error)
	UpdateConnectionType(connectionTypeId string, newConnectionTypeName string) error
	DeleteConnectionType(connectionTypeName string) error

	// Connection interaction
	AddConnection(
		fromDeviceId string,
		fromModelPortId string,
		fromIPSegment string,
		toDeviceId string,
		toModelPortId string,
		toIPSegment string,
	) error
	GetConnections() ([]e.Connection, error)
	UpdateConnection(
		connectionId string,
		newFromDeviceId string,
		newFromModelPortId string,
		newFromIPSegment string,
		newToDeviceId string,
		newToModelPortId string,
		newToIPSegment string,
	) error
	DeleteConnection(connectionId string) error

	// get all the ports mapped
	GetAllPortsAll() ([]e.DevicePort, error)
	// get all the ports mapped for a device
	GetAllPortsDevice(deviceid string) ([]e.DevicePort, error)
	// Export info
	ExportAllStructs() (e.All, error)
}

// type NetRepository exposes the interface for implementation
type NetRepository repository
