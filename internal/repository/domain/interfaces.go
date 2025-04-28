
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
