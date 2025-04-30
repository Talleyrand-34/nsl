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
// Package domain set ups the interface for implementations for operations with the repository
package domain

import e "nsl-graph/internal/repository/entities"

// Comment

type Repository interface {
	// Brand
	AddBrand(brand string) error
	GetBrands() ([]e.Brand, error)
	DeleteBrand(brand string) error

	// DeviceClass
	AddDeviceClass(devclass string) error
	GetDeviceClasses() ([]e.DevClass, error)
	DeleteDeviceClass(devclass string) error

	// ZoneType
	AddZoneType(name string) error
	GetZonetypes() ([]e.ZoneType, error)
	DeleteZoneType(name string) error

	// Proprietary
	AddProprietary(name string) error
	GetProperties() ([]e.Proprietary, error)
	DeleteProprietary(name string) error

	// Zone
	AddZone(
		name string,
		fatherid string,
		father string,
		proprietary string,
		zonename string,
	) error
	GetZones() ([]e.Zone, error)
	DeleteZone(name string) error

	// Model
	AddModel(
		modelName string,
		brandName string,
		className string,
	) error
	GetModels() ([]e.ModelDevice, error)
	DeleteModel(modelName string) error

	// Device
	AddDevice(
		label string,
		model string,
		zoneId string,
		zoneName string,
		proprietary string,
	) error
	GetDevices() ([]e.Device, error)
	DeleteDevice(deviceId string) error

	// ModelPort
	AddModelPort(
		name string,
		posx string,
		posy string,
		modelName string,
	) error
	GetModelPorts() ([]e.ModelPort, error)
	DeleteModelPort(modelPortId string) error

	// DevicePort
	AddDevicePort(deviceid string, modelportid string) error
	GetDevicePorts() ([]e.DevicePort, error)
	DeleteDevicePort(devicePortId string, modelportid string) error

	// Connection
	AddConnection(
		fromDevice string,
		fromModelPort string,
		toDevice string,
		toModelPort string,
	) error
	GetConnections() ([]e.Connection, error)
	DeleteConnection(connectionId string) error

	GetAllPortsDevice(deviceid string) ([]e.DevicePort, error)
	GetAllPortsAll() ([]e.DevicePort, error)
	UpdateConnection(
		id string,
		from_device string,
		from_port string,
		to_device string,
		to_port string,
	) error
	ExportAllStructs() (e.All, error)
}

type NetRepository Repository
