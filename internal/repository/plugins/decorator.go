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
package plugins

import (
	"nsl-graph/internal/repository/domain"
	e "nsl-graph/internal/repository/entities"
)

// PluginAwareRepository wraps a base repository and applies plugins
// to transform data without modifying the underlying database implementation.
// It implements the NetRepository interface by delegating to the base repository
// and applying plugin transformations.
type PluginAwareRepository struct {
	baseRepo domain.NetRepository
	registry *Registry
}

// NewPluginAwareRepository creates a new plugin-aware repository decorator
func NewPluginAwareRepository(baseRepo domain.NetRepository, registry *Registry) *PluginAwareRepository {
	return &PluginAwareRepository{
		baseRepo: baseRepo,
		registry: registry,
	}
}

// GetConnections retrieves connections from the base repository and applies
// the active sorting plugin
func (p *PluginAwareRepository) GetConnections() ([]e.Connection, error) {
	// Get connections from base repository (CloverDB)
	connections, err := p.baseRepo.GetConnections()
	if err != nil {
		return nil, err
	}

	// Apply active sorting plugin
	sortedConnections := p.registry.ApplySorting(connections)

	return sortedConnections, nil
}

// All other methods delegate directly to the base repository without transformation

// Brand interaction
func (p *PluginAwareRepository) AddBrand(brandName string) error {
	return p.baseRepo.AddBrand(brandName)
}

func (p *PluginAwareRepository) GetBrands() ([]e.Brand, error) {
	return p.baseRepo.GetBrands()
}

func (p *PluginAwareRepository) UpdateBrand(brandId string, newBrandName string) error {
	return p.baseRepo.UpdateBrand(brandId, newBrandName)
}

func (p *PluginAwareRepository) DeleteBrand(brandName string) error {
	return p.baseRepo.DeleteBrand(brandName)
}

// DeviceClass interaction
func (p *PluginAwareRepository) AddDeviceClass(deviceClassName string) error {
	return p.baseRepo.AddDeviceClass(deviceClassName)
}

func (p *PluginAwareRepository) GetDeviceClasses() ([]e.DevClass, error) {
	return p.baseRepo.GetDeviceClasses()
}

func (p *PluginAwareRepository) UpdateDeviceClass(deviceClassId string, newDeviceClassName string) error {
	return p.baseRepo.UpdateDeviceClass(deviceClassId, newDeviceClassName)
}

func (p *PluginAwareRepository) DeleteDeviceClass(deviceClassName string) error {
	return p.baseRepo.DeleteDeviceClass(deviceClassName)
}

// ZoneType interaction
func (p *PluginAwareRepository) AddZoneType(zoneTypeName string) error {
	return p.baseRepo.AddZoneType(zoneTypeName)
}

func (p *PluginAwareRepository) GetZonetypes() ([]e.ZoneType, error) {
	return p.baseRepo.GetZonetypes()
}

func (p *PluginAwareRepository) UpdateZoneType(zoneTypeId string, newZoneTypeName string) error {
	return p.baseRepo.UpdateZoneType(zoneTypeId, newZoneTypeName)
}

func (p *PluginAwareRepository) DeleteZoneType(zoneTypeName string) error {
	return p.baseRepo.DeleteZoneType(zoneTypeName)
}

// Proprietary interaction
func (p *PluginAwareRepository) AddProprietary(proprietaryName string) error {
	return p.baseRepo.AddProprietary(proprietaryName)
}

func (p *PluginAwareRepository) GetProperties() ([]e.Proprietary, error) {
	return p.baseRepo.GetProperties()
}

func (p *PluginAwareRepository) UpdateProprietary(proprietaryId string, newProprietaryName string) error {
	return p.baseRepo.UpdateProprietary(proprietaryId, newProprietaryName)
}

func (p *PluginAwareRepository) DeleteProprietary(proprietaryName string) error {
	return p.baseRepo.DeleteProprietary(proprietaryName)
}

// Zone interaction
func (p *PluginAwareRepository) AddZone(zoneName string, fatherZoneId string, fatherZoneName string, proprietaryName string, zoneTypeName string) error {
	return p.baseRepo.AddZone(zoneName, fatherZoneId, fatherZoneName, proprietaryName, zoneTypeName)
}

func (p *PluginAwareRepository) GetZones() ([]e.Zone, error) {
	return p.baseRepo.GetZones()
}

func (p *PluginAwareRepository) UpdateZone(zoneId string, newZoneName string, newFatherZoneId string, newZoneTypeId string, newProprietaryId string) error {
	return p.baseRepo.UpdateZone(zoneId, newZoneName, newFatherZoneId, newZoneTypeId, newProprietaryId)
}

func (p *PluginAwareRepository) DeleteZone(zoneId string) error {
	return p.baseRepo.DeleteZone(zoneId)
}

// Model interaction
func (p *PluginAwareRepository) AddModel(modelName string, brandName string, deviceClassName string) error {
	return p.baseRepo.AddModel(modelName, brandName, deviceClassName)
}

func (p *PluginAwareRepository) GetModels() ([]e.ModelDevice, error) {
	return p.baseRepo.GetModels()
}

func (p *PluginAwareRepository) UpdateModel(modelId string, newModelName string, newBrandId string, newDeviceClassId string) error {
	return p.baseRepo.UpdateModel(modelId, newModelName, newBrandId, newDeviceClassId)
}

func (p *PluginAwareRepository) DeleteModel(modelId string) error {
	return p.baseRepo.DeleteModel(modelId)
}

// Device interaction
func (p *PluginAwareRepository) AddDevice(deviceLabel string, modelName string, zoneId string, zoneName string, proprietaryName string, ips []string) error {
	return p.baseRepo.AddDevice(deviceLabel, modelName, zoneId, zoneName, proprietaryName, ips)
}

func (p *PluginAwareRepository) GetDevices() ([]e.Device, error) {
	return p.baseRepo.GetDevices()
}

func (p *PluginAwareRepository) UpdateDevice(deviceId string, newDeviceLabel string, newModelId string, newZoneId string, newProprietaryId string) error {
	return p.baseRepo.UpdateDevice(deviceId, newDeviceLabel, newModelId, newZoneId, newProprietaryId)
}

func (p *PluginAwareRepository) UpdateDeviceIPs(deviceId string, ips []string) error {
	return p.baseRepo.UpdateDeviceIPs(deviceId, ips)
}

func (p *PluginAwareRepository) DeleteDevice(deviceId string) error {
	return p.baseRepo.DeleteDevice(deviceId)
}

// ModelPort interaction
func (p *PluginAwareRepository) AddModelPort(portName string, positionX string, positionY string, modelName string, allowMultipleConnections bool, portType string, band string) error {
	return p.baseRepo.AddModelPort(portName, positionX, positionY, modelName, allowMultipleConnections, portType, band)
}

func (p *PluginAwareRepository) GetModelPorts() ([]e.ModelPort, error) {
	return p.baseRepo.GetModelPorts()
}

func (p *PluginAwareRepository) UpdateModelPort(modelPortId string, newPortName string, newPositionX string, newPositionY string, newModelId string, newAllowMultipleConnections bool) error {
	return p.baseRepo.UpdateModelPort(modelPortId, newPortName, newPositionX, newPositionY, newModelId, newAllowMultipleConnections)
}

func (p *PluginAwareRepository) DeleteModelPort(modelPortId string) error {
	return p.baseRepo.DeleteModelPort(modelPortId)
}

// DevicePort interaction
func (p *PluginAwareRepository) DevicePortExists(deviceId string, modelPortId string) (bool, error) {
	return p.baseRepo.DevicePortExists(deviceId, modelPortId)
}

func (p *PluginAwareRepository) AddDevicePort(deviceId string, modelPortId string, macAddress string, vlanConfigs []e.PortVlanConfig) error {
	return p.baseRepo.AddDevicePort(deviceId, modelPortId, macAddress, vlanConfigs)
}

func (p *PluginAwareRepository) GetDevicePorts() ([]e.DevicePort, error) {
	return p.baseRepo.GetDevicePorts()
}

func (p *PluginAwareRepository) GetDevicePortByIDs(deviceId string, modelPortId string) (*e.DevicePort, error) {
	return p.baseRepo.GetDevicePortByIDs(deviceId, modelPortId)
}

func (p *PluginAwareRepository) UpdateDevicePort(deviceId string, modelPortId string, macAddress string, vlanConfigs []e.PortVlanConfig) error {
	return p.baseRepo.UpdateDevicePort(deviceId, modelPortId, macAddress, vlanConfigs)
}

func (p *PluginAwareRepository) DeleteDevicePort(deviceId string, modelPortId string) error {
	return p.baseRepo.DeleteDevicePort(deviceId, modelPortId)
}

// ConnectionType interaction
func (p *PluginAwareRepository) AddConnectionType(connectionTypeName string) error {
	return p.baseRepo.AddConnectionType(connectionTypeName)
}

func (p *PluginAwareRepository) GetConnectionTypes() ([]e.ConnectionType, error) {
	return p.baseRepo.GetConnectionTypes()
}

func (p *PluginAwareRepository) UpdateConnectionType(connectionTypeId string, newConnectionTypeName string) error {
	return p.baseRepo.UpdateConnectionType(connectionTypeId, newConnectionTypeName)
}

func (p *PluginAwareRepository) DeleteConnectionType(connectionTypeName string) error {
	return p.baseRepo.DeleteConnectionType(connectionTypeName)
}

// Connection interaction
func (p *PluginAwareRepository) AddConnection(fromDeviceId string, fromModelPortId string, toDeviceId string, toModelPortId string, allowVLANUnion bool) error {
	return p.baseRepo.AddConnection(fromDeviceId, fromModelPortId, toDeviceId, toModelPortId, allowVLANUnion)
}

func (p *PluginAwareRepository) UpdateConnection(connectionId string, newFromDeviceId string, newFromModelPortId string, newToDeviceId string, newToModelPortId string, allowVLANUnion bool) error {
	return p.baseRepo.UpdateConnection(connectionId, newFromDeviceId, newFromModelPortId, newToDeviceId, newToModelPortId, allowVLANUnion)
}

func (p *PluginAwareRepository) DeleteConnection(connectionId string) error {
	return p.baseRepo.DeleteConnection(connectionId)
}

// VLAN interaction
func (p *PluginAwareRepository) AddVlan(vlanID string, vlanName string, ipSegment string) error {
	return p.baseRepo.AddVlan(vlanID, vlanName, ipSegment)
}

func (p *PluginAwareRepository) GetVlans() ([]e.Vlan, error) {
	return p.baseRepo.GetVlans()
}

func (p *PluginAwareRepository) UpdateVlan(vlanId string, newVlanID string, newVlanName string) error {
	return p.baseRepo.UpdateVlan(vlanId, newVlanID, newVlanName)
}

func (p *PluginAwareRepository) UpdateVlanIPSegment(vlanId string, ipSegment string) error {
	return p.baseRepo.UpdateVlanIPSegment(vlanId, ipSegment)
}

func (p *PluginAwareRepository) DeleteVlan(vlanId string) error {
	return p.baseRepo.DeleteVlan(vlanId)
}

// Local VLAN interaction
func (p *PluginAwareRepository) AddLocalVlan(vlanID string, deviceID string, vlanName string) error {
	return p.baseRepo.AddLocalVlan(vlanID, deviceID, vlanName)
}

func (p *PluginAwareRepository) GetLocalVlans() ([]e.LocalVlan, error) {
	return p.baseRepo.GetLocalVlans()
}

func (p *PluginAwareRepository) GetLocalVlansByDevice(deviceID string) ([]e.LocalVlan, error) {
	return p.baseRepo.GetLocalVlansByDevice(deviceID)
}

func (p *PluginAwareRepository) GetLocalVlansByVlanID(vlanID string) ([]e.LocalVlan, error) {
	return p.baseRepo.GetLocalVlansByVlanID(vlanID)
}

func (p *PluginAwareRepository) UpdateLocalVlan(localVlanId string, newVlanID string, newDeviceID string, newVlanName string) error {
	return p.baseRepo.UpdateLocalVlan(localVlanId, newVlanID, newDeviceID, newVlanName)
}

func (p *PluginAwareRepository) UpdateLocalVlanByMapping(vlanID string, deviceID string, newVlanName string) error {
	return p.baseRepo.UpdateLocalVlanByMapping(vlanID, deviceID, newVlanName)
}

func (p *PluginAwareRepository) DeleteLocalVlan(localVlanId string) error {
	return p.baseRepo.DeleteLocalVlan(localVlanId)
}

func (p *PluginAwareRepository) DeleteLocalVlansByDevice(deviceID string) error {
	return p.baseRepo.DeleteLocalVlansByDevice(deviceID)
}

func (p *PluginAwareRepository) DeleteLocalVlansByVlanID(vlanID string) error {
	return p.baseRepo.DeleteLocalVlansByVlanID(vlanID)
}

// Get all ports methods
func (p *PluginAwareRepository) GetAllPortsAll() ([]e.DevicePort, error) {
	return p.baseRepo.GetAllPortsAll()
}

func (p *PluginAwareRepository) GetAllPortsDevice(deviceid string) ([]e.DevicePort, error) {
	return p.baseRepo.GetAllPortsDevice(deviceid)
}

// Cascade deletion methods
func (p *PluginAwareRepository) DeleteBrandCascade(brandName string) error {
	return p.baseRepo.DeleteBrandCascade(brandName)
}

func (p *PluginAwareRepository) DeleteDeviceClassCascade(deviceClassName string) error {
	return p.baseRepo.DeleteDeviceClassCascade(deviceClassName)
}

func (p *PluginAwareRepository) DeleteZoneTypeCascade(zoneTypeName string) error {
	return p.baseRepo.DeleteZoneTypeCascade(zoneTypeName)
}

func (p *PluginAwareRepository) DeleteProprietaryCascade(proprietaryName string) error {
	return p.baseRepo.DeleteProprietaryCascade(proprietaryName)
}

func (p *PluginAwareRepository) DeleteZoneCascade(zoneId string) error {
	return p.baseRepo.DeleteZoneCascade(zoneId)
}

func (p *PluginAwareRepository) DeleteModelCascade(modelId string) error {
	return p.baseRepo.DeleteModelCascade(modelId)
}

func (p *PluginAwareRepository) DeleteDeviceCascade(deviceId string) error {
	return p.baseRepo.DeleteDeviceCascade(deviceId)
}

func (p *PluginAwareRepository) DeleteModelPortCascade(modelPortId string) error {
	return p.baseRepo.DeleteModelPortCascade(modelPortId)
}

func (p *PluginAwareRepository) DeleteDevicePortCascade(deviceId string, modelPortId string) error {
	return p.baseRepo.DeleteDevicePortCascade(deviceId, modelPortId)
}

func (p *PluginAwareRepository) DeleteConnectionCascade(connectionId string) error {
	return p.baseRepo.DeleteConnectionCascade(connectionId)
}

func (p *PluginAwareRepository) DeleteVlanCascade(vlanId string) error {
	return p.baseRepo.DeleteVlanCascade(vlanId)
}

// DeviceInterface interaction
func (p *PluginAwareRepository) AddDeviceInterface(deviceID, name, description string, vlanConfigs []e.PortVlanConfig, ips []string, wifiSSID, wifiSecurity string) error {
	return p.baseRepo.AddDeviceInterface(deviceID, name, description, vlanConfigs, ips, wifiSSID, wifiSecurity)
}

func (p *PluginAwareRepository) GetDeviceInterfaces(deviceID string) ([]e.DeviceInterface, error) {
	return p.baseRepo.GetDeviceInterfaces(deviceID)
}

func (p *PluginAwareRepository) GetAllDeviceInterfaces() ([]e.DeviceInterface, error) {
	return p.baseRepo.GetAllDeviceInterfaces()
}

func (p *PluginAwareRepository) UpdateDeviceInterface(id string, vlanConfigs []e.PortVlanConfig) error {
	return p.baseRepo.UpdateDeviceInterface(id, vlanConfigs)
}

func (p *PluginAwareRepository) DeleteDeviceInterface(id string) error {
	return p.baseRepo.DeleteDeviceInterface(id)
}

// InterfacePort interaction
func (p *PluginAwareRepository) AddInterfacePort(interfaceID, deviceID, modelPortID string) error {
	return p.baseRepo.AddInterfacePort(interfaceID, deviceID, modelPortID)
}

func (p *PluginAwareRepository) GetInterfacePortsByInterface(interfaceID string) ([]e.InterfacePort, error) {
	return p.baseRepo.GetInterfacePortsByInterface(interfaceID)
}

func (p *PluginAwareRepository) GetInterfacePortsByPort(deviceID, modelPortID string) ([]e.InterfacePort, error) {
	return p.baseRepo.GetInterfacePortsByPort(deviceID, modelPortID)
}

func (p *PluginAwareRepository) GetAllInterfacePorts() ([]e.InterfacePort, error) {
	return p.baseRepo.GetAllInterfacePorts()
}

func (p *PluginAwareRepository) DeleteInterfacePort(interfaceID, deviceID, modelPortID string) error {
	return p.baseRepo.DeleteInterfacePort(interfaceID, deviceID, modelPortID)
}

// Export info
func (p *PluginAwareRepository) ExportAllStructs() (e.All, error) {
	return p.baseRepo.ExportAllStructs()
}
