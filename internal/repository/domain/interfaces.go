// Package domain set ups the interface for implementations for operations with the repository
package domain

/*
  Copyright © 2025 Talleyrand-34 (t34@t34.dev)

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
		isUnmanaged bool,
		isInvisible bool,
	) error
	GetDevices() ([]e.Device, error)
	UpdateDevice(
		deviceId string,
		newDeviceLabel string,
		newModelId string,
		newZoneId string,
		newProprietaryId string,
		isUnmanaged *bool,
	) error
	UpdateDeviceIPs(deviceId string, ips []string) error
	MigrateDeviceModel(deviceId string, newModelId string, portMap map[string]string) error
	DeleteDevice(deviceId string) error

	// ModelPort interaction
	AddModelPort(
		portName string,
		positionX string,
		positionY string,
		modelName string,
		allowMultipleConnections bool,
		portType string,
		band string,
	) error
	GetModelPorts() ([]e.ModelPort, error)
	UpdateModelPort(
		modelPortId string,
		newPortName string,
		newPositionX string,
		newPositionY string,
		newModelId string,
		newAllowMultipleConnections bool,
	) error
	DeleteModelPort(modelPortId string) error

	// DevicePort interaction
	DevicePortExists(deviceId string, modelPortId string) (bool, error)
	AddDevicePort(deviceId string, modelPortId string, macAddress string, vlanConfigs []e.PortVlanConfig) (string, error)
	GetDevicePorts() ([]e.DevicePort, error)
	GetDevicePortByIDs(deviceId string, modelPortId string) (*e.DevicePort, error)
	UpdateDevicePort(deviceId string, modelPortId string, macAddress string, vlanConfigs []e.PortVlanConfig) error
	DeleteDevicePort(deviceId string, modelPortId string) error

	// ConnectionType interaction
	AddConnectionType(connectionTypeName string) error
	GetConnectionTypes() ([]e.ConnectionType, error)
	UpdateConnectionType(connectionTypeId string, newConnectionTypeName string) error
	DeleteConnectionType(connectionTypeName string) error

	// Connection interaction
	AddConnection(
		fromDeviceportID string,
		toDeviceportID string,
		connectionType string,
		discoveredVia ...string,
	) error
	GetConnections() ([]e.Connection, error)
	UpdateConnection(
		connectionId string,
		newFromDeviceportID string,
		newToDeviceportID string,
		connectionType string,
	) error
	DeleteConnection(connectionId string) error

	// VLAN interaction
	AddVlan(vlanID string, vlanName string, ipSegment string) error
	GetVlans() ([]e.Vlan, error)
	UpdateVlan(vlanId string, newVlanID string, newVlanName string) error
	UpdateVlanIPSegment(vlanId string, ipSegment string) error
	DeleteVlan(vlanId string) error

	// Cascade deletion methods
	DeleteBrandCascade(brandName string) error
	DeleteDeviceClassCascade(deviceClassName string) error
	DeleteZoneTypeCascade(zoneTypeName string) error
	DeleteProprietaryCascade(proprietaryName string) error
	DeleteZoneCascade(zoneId string) error
	DeleteModelCascade(modelId string) error
	DeleteDeviceCascade(deviceId string) error
	DeleteModelPortCascade(modelPortId string) error
	DeleteDevicePortCascade(deviceId string, modelPortId string) error
	DeleteConnectionCascade(connectionId string) error
	DeleteVlanCascade(vlanId string) error

	// DeviceInterface interaction
	AddDeviceInterface(deviceID, name, description, parent string, vlanConfigs []e.PortVlanConfig, ips []string, wifiSSID, wifiSecurity string) error
	GetDeviceInterfaces(deviceID string) ([]e.DeviceInterface, error)
	GetAllDeviceInterfaces() ([]e.DeviceInterface, error)
	UpdateDeviceInterface(id string, vlanConfigs []e.PortVlanConfig) error
	DeleteDeviceInterface(id string) error

	// InterfacePort interaction
	AddInterfacePort(interfaceID, deviceID, modelPortID string) error
	GetInterfacePortsByInterface(interfaceID string) ([]e.InterfacePort, error)
	GetInterfacePortsByPort(deviceID, modelPortID string) ([]e.InterfacePort, error)
	GetAllInterfacePorts() ([]e.InterfacePort, error)
	DeleteInterfacePort(interfaceID, deviceID, modelPortID string) error

	// get all the ports mapped
	GetAllPortsAll() ([]e.DevicePort, error)
	// get all the ports mapped for a device
	GetAllPortsDevice(deviceid string) ([]e.DevicePort, error)
	// Export info
	ExportAllStructs() (e.All, error)

	// ScanProfile interaction
	AddScanProfile(p e.ScanProfile) error
	GetScanProfiles() ([]e.ScanProfile, error)
	GetScanProfileByName(name string) (*e.ScanProfile, error)
	GetScanProfileByHost(host string) (*e.ScanProfile, error)
	UpdateScanProfile(p e.ScanProfile) error
	DeleteScanProfile(name string) error
}

// type NetRepository exposes the interface for implementation
type NetRepository repository
