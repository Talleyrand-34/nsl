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
// package application set ups an interface for access to database and transforms outputs to json
package application

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	d "nsl-graph/internal/repository/domain"
	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
)

type NetService struct {
	netRepo d.NetRepository
}

type NetServiceInt interface {
	// Brand operations
	AddBrand(brandName string) error
	GetBrands() ([]e.Brand, error)
	UpdateBrand(brandId string, newBrandName string) error
	DeleteBrand(brandName string) error

	// DeviceClass operations
	AddDeviceClass(deviceClassName string) error
	GetDeviceClasses() ([]e.DevClass, error)
	UpdateDeviceClass(deviceClassId string, newDeviceClassName string) error
	DeleteDeviceClass(deviceClassName string) error

	// ZoneType operations
	AddZoneType(zoneTypeName string) error
	GetZonetypes() ([]e.ZoneType, error)
	UpdateZoneType(zoneTypeId string, newZoneTypeName string) error
	DeleteZoneType(zoneTypeName string) error

	// Proprietary operations
	AddProprietary(proprietaryName string) error
	GetProperties() ([]e.Proprietary, error)
	UpdateProprietary(proprietaryId string, newProprietaryName string) error
	DeleteProprietary(proprietaryName string) error

	// Zone operations
	AddZone(
		zoneName string,
		fatherZoneName string,
		fatherZoneId string,
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

	// Model operations
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

	// Device operations
	AddDevice(
		deviceLabel string,
		modelName string,
		zoneId string,
		zoneName string,
		proprietaryName string,
		ips []string,
	) error
	GetDevices() ([]e.Device, error)
	UpdateDevice(
		deviceId string,
		newDeviceLabel string,
		newModelId string,
		newZoneId string,
		newProprietaryId string,
	) error
	UpdateDeviceIPs(deviceId string, ips []string) error
	DeleteDevice(deviceId string) error

	// ModelPort operations
	AddModelPort(
		portName string,
		positionX string,
		positionY string,
		modelName string,
		allowMultipleConnections bool,
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

	// DevicePort operations
	AddDevicePort(deviceId string, modelPortId string, macAddress string, vlanConfigs []e.PortVlanConfig) error
	GetDevicePorts() ([]e.DevicePort, error)
	UpdateDevicePort(deviceId string, modelPortId string, macAddress string, vlanConfigs []e.PortVlanConfig) error
	DeleteDevicePort(deviceId string, modelPortId string) error

	// ConnectionType operations
	AddConnectionType(connectionTypeName string) error
	GetConnectionTypes() ([]e.ConnectionType, error)
	UpdateConnectionType(connectionTypeId string, newConnectionTypeName string) error
	DeleteConnectionType(connectionTypeName string) error

	// Connection operations
	AddConnection(
		fromDeviceId string,
		fromModelPortId string,
		toDeviceId string,
		toModelPortId string,
		allowVLANUnion bool,
	) error
	GetConnections() ([]e.Connection, error)
	UpdateConnection(
		connectionId string,
		newFromDeviceId string,
		newFromModelPortId string,
		newToDeviceId string,
		newToModelPortId string,
		allowVLANUnion bool,
	) error
	DeleteConnection(connectionId string) error

	// VLAN operations
	AddVlan(vlanID string, vlanName string, ipSegmentIDs []string) error
	GetVlans() ([]e.Vlan, error)
	UpdateVlan(vlanId string, newVlanID string, newVlanName string) error
	UpdateVlanIPSegments(vlanId string, ipSegmentIDs []string) error
	DeleteVlan(vlanId string) error

	// Special operations
	GetAllPortsDevice(deviceId string) ([]e.DevicePort, error)
	GetAllPortsAll() ([]e.DevicePort, error)
	ExportAllStructs() []byte

	// Network scanning operations
	ScanNetwork(subnet string, options s.ScanOptions) (*s.ScanResult, error)
	ScanDevice(ip string, options s.ScanOptions) (*s.SNMPDevice, error)
	DiscoverDevices(scanResult *s.ScanResult) ([]s.DiscoveredDevice, error)
	ImportScanResults(devices []s.DiscoveredDevice, options s.ImportOptions) error
	ImportDiscoveredDevices(devices []s.DiscoveredDevice, options s.ImportOptions) error
}

func NewNetService(netRepository d.NetRepository) NetServiceInt {
	return &NetService{netRepo: netRepository}
}

func (ns *NetService) AddBrand(brandName string) error {
	return ns.netRepo.AddBrand(brandName)
}

func (ns *NetService) GetBrands() ([]e.Brand, error) {
	return ns.netRepo.GetBrands()
}

func (ns *NetService) UpdateBrand(brandId string, newBrandName string) error {
	return ns.netRepo.UpdateBrand(brandId, newBrandName)
}

func (ns *NetService) AddDeviceClass(deviceClassName string) error {
	return ns.netRepo.AddDeviceClass(deviceClassName)
}

func (ns *NetService) GetDeviceClasses() ([]e.DevClass, error) {
	return ns.netRepo.GetDeviceClasses()
}

func (ns *NetService) AddZoneType(zoneTypeName string) error {
	return ns.netRepo.AddZoneType(zoneTypeName)
}

func (ns *NetService) GetZonetypes() ([]e.ZoneType, error) {
	return ns.netRepo.GetZonetypes()
}

func (ns *NetService) AddProprietary(proprietary string) error {
	return ns.netRepo.AddProprietary(proprietary)
}

func (ns *NetService) GetProperties() ([]e.Proprietary, error) {
	return ns.netRepo.GetProperties()
}

func (ns *NetService) AddZone(
	name string,
	fatherid string,
	father string,
	proprietary string,
	zonename string,
) error {
	return ns.netRepo.AddZone(name, fatherid, father, proprietary, zonename)
}

func (ns *NetService) GetZones() ([]e.Zone, error) {
	return ns.netRepo.GetZones()
}

func (ns *NetService) GetModels() ([]e.ModelDevice, error) {
	return ns.netRepo.GetModels()
}

func (ns *NetService) AddModelPort(
	name string,
	posx string,
	posy string,
	modelName string,
	allowMultipleConnections bool,
) error {
	return ns.netRepo.AddModelPort(name, posx, posy, modelName, allowMultipleConnections)
}

func (ns *NetService) GetModelPorts() ([]e.ModelPort, error) {
	return ns.netRepo.GetModelPorts()
}

func (ns *NetService) AddModel(
	modelName string,
	brandName string,
	className string,
) error {
	return ns.netRepo.AddModel(modelName, brandName, className)
}

func (ns *NetService) GetDevices() ([]e.Device, error) {
	return ns.netRepo.GetDevices()
}

func (ns *NetService) AddDevice(
	label string,
	model string,
	zoneId string,
	zoneName string,
	proprietary string,
	ips []string,
) error {
	return ns.netRepo.AddDevice(label, model, zoneId, zoneName, proprietary, ips)
}

func (ns *NetService) UpdateDeviceIPs(deviceId string, ips []string) error {
	return ns.netRepo.UpdateDeviceIPs(deviceId, ips)
}

func (ns *NetService) GetDevicePorts() ([]e.DevicePort, error) {
	zones, err := ns.netRepo.GetDevicePorts()
	return zones, err
}

func (ns *NetService) AddDevicePort(
	deviceid string,
	modelportid string,
	macAddress string,
	vlanConfigs []e.PortVlanConfig,
) error {
	return ns.netRepo.AddDevicePort(deviceid, modelportid, macAddress, vlanConfigs)
}

func (ns *NetService) UpdateDevicePort(deviceid string, modelportid string, macAddress string, vlanConfigs []e.PortVlanConfig) error {
	return ns.netRepo.UpdateDevicePort(deviceid, modelportid, macAddress, vlanConfigs)
}

func (ns *NetService) AddConnectionType(connectionTypeName string) error {
	return ns.netRepo.AddConnectionType(connectionTypeName)
}

func (ns *NetService) GetConnectionTypes() ([]e.ConnectionType, error) {
	return ns.netRepo.GetConnectionTypes()
}

func (ns *NetService) GetConnections() ([]e.Connection, error) {
	return ns.netRepo.GetConnections()
}

func (ns *NetService) AddConnection(
	fromDevice string,
	fromModelPort string,
	toDevice string,
	toModelPort string,
	allowVLANUnion bool,
) error {
	// Business logic: Ensure device ports exist before creating connection
	// Check if "from" device port exists, create if it doesn't
	fromExists, err := ns.netRepo.DevicePortExists(fromDevice, fromModelPort)
	if err != nil {
		return fmt.Errorf("error checking from device port: %w", err)
	}
	if !fromExists {
		// Try to create the device port with empty VLANs
		if err := ns.netRepo.AddDevicePort(fromDevice, fromModelPort, "", nil); err != nil {
			return fmt.Errorf("error creating from device port: %w", err)
		}
	}

	// Check if "to" device port exists, create if it doesn't
	toExists, err := ns.netRepo.DevicePortExists(toDevice, toModelPort)
	if err != nil {
		return fmt.Errorf("error checking to device port: %w", err)
	}
	if !toExists {
		// Try to create the device port with empty VLANs
		if err := ns.netRepo.AddDevicePort(toDevice, toModelPort, "", nil); err != nil {
			return fmt.Errorf("error creating to device port: %w", err)
		}
	}

	// Create the connection (VLAN validation happens in repository layer)
	return ns.netRepo.AddConnection(fromDevice, fromModelPort, toDevice, toModelPort, allowVLANUnion)
}

func (ns *NetService) GetAllPortsAll() ([]e.DevicePort, error) {
	return ns.netRepo.GetAllPortsAll()
}

func (ns *NetService) GetAllPortsDevice(deviceid string) ([]e.DevicePort, error) {
	return ns.netRepo.GetAllPortsDevice(deviceid)
}

func (ns *NetService) ExportAllStructs() []byte {
	export, err := ns.netRepo.ExportAllStructs()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(export)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) DeleteBrand(brand string) error {
	return ns.netRepo.DeleteBrand(brand)
}

func (ns *NetService) DeleteDeviceClass(name string) error {
	return ns.netRepo.DeleteDeviceClass(name)
}

func (ns *NetService) DeleteZoneType(locationType string) error {
	return ns.netRepo.DeleteZoneType(locationType)
}

func (ns *NetService) DeleteProprietary(proprietary string) error {
	return ns.netRepo.DeleteProprietary(proprietary)
}

func (ns *NetService) DeleteZone(id string) error {
	return ns.netRepo.DeleteZone(id)
}

func (ns *NetService) DeleteModel(id string) error {
	return ns.netRepo.DeleteModel(id)
}

func (ns *NetService) DeleteDevice(id string) error {
	return ns.netRepo.DeleteDevice(id)
}

func (ns *NetService) DeleteModelPort(id string) error {
	return ns.netRepo.DeleteModelPort(id)
}

func (ns *NetService) DeleteDevicePort(deviceID, modelPortID string) error {
	return ns.netRepo.DeleteDevicePort(deviceID, modelPortID)
}

func (ns *NetService) DeleteConnection(id string) error {
	return ns.netRepo.DeleteConnection(id)
}

func (ns *NetService) UpdateDeviceClass(deviceClassId string, newDeviceClassName string) error {
	return ns.netRepo.UpdateDeviceClass(deviceClassId, newDeviceClassName)
}

func (ns *NetService) UpdateZoneType(zoneTypeId string, newZoneTypeName string) error {
	return ns.netRepo.UpdateZoneType(zoneTypeId, newZoneTypeName)
}

func (ns *NetService) UpdateProprietary(proprietaryId string, newProprietaryName string) error {
	return ns.netRepo.UpdateProprietary(proprietaryId, newProprietaryName)
}

func (ns *NetService) UpdateZone(
	zoneId string,
	newZoneName string,
	newFatherZoneId string,
	newZoneTypeId string,
	newProprietaryId string,
) error {
	return ns.netRepo.UpdateZone(zoneId, newZoneName, newFatherZoneId, newZoneTypeId, newProprietaryId)
}

func (ns *NetService) UpdateModel(
	modelId string,
	newModelName string,
	newBrandId string,
	newDeviceClassId string,
) error {
	return ns.netRepo.UpdateModel(modelId, newModelName, newBrandId, newDeviceClassId)
}

func (ns *NetService) UpdateDevice(
	deviceId string,
	newDeviceLabel string,
	newModelId string,
	newZoneId string,
	newProprietaryId string,
) error {
	return ns.netRepo.UpdateDevice(deviceId, newDeviceLabel, newModelId, newZoneId, newProprietaryId)
}

func (ns *NetService) UpdateModelPort(
	modelPortId string,
	newPortName string,
	newPositionX string,
	newPositionY string,
	newModelId string,
	newAllowMultipleConnections bool,
) error {
	return ns.netRepo.UpdateModelPort(modelPortId, newPortName, newPositionX, newPositionY, newModelId, newAllowMultipleConnections)
}

func (ns *NetService) UpdateConnectionType(connectionTypeId string, newConnectionTypeName string) error {
	return ns.netRepo.UpdateConnectionType(connectionTypeId, newConnectionTypeName)
}

func (ns *NetService) DeleteConnectionType(connectionTypeName string) error {
	return ns.netRepo.DeleteConnectionType(connectionTypeName)
}

func (ns *NetService) UpdateConnection(
	connectionId string,
	newFromDeviceId string,
	newFromModelPortId string,
	newToDeviceId string,
	newToModelPortId string,
	allowVLANUnion bool,
) error {
	// Business logic: Ensure device ports exist before updating connection
	// Check if "from" device port exists, create if it doesn't
	fromExists, err := ns.netRepo.DevicePortExists(newFromDeviceId, newFromModelPortId)
	if err != nil {
		return fmt.Errorf("error checking from device port: %w", err)
	}
	if !fromExists {
		// Try to create the device port with empty VLANs
		if err := ns.netRepo.AddDevicePort(newFromDeviceId, newFromModelPortId, "", nil); err != nil {
			return fmt.Errorf("error creating from device port: %w", err)
		}
	}

	// Check if "to" device port exists, create if it doesn't
	toExists, err := ns.netRepo.DevicePortExists(newToDeviceId, newToModelPortId)
	if err != nil {
		return fmt.Errorf("error checking to device port: %w", err)
	}
	if !toExists {
		// Try to create the device port with empty VLANs
		if err := ns.netRepo.AddDevicePort(newToDeviceId, newToModelPortId, "", nil); err != nil {
			return fmt.Errorf("error creating to device port: %w", err)
		}
	}

	// Update the connection (VLAN validation happens in repository layer)
	return ns.netRepo.UpdateConnection(connectionId, newFromDeviceId, newFromModelPortId, newToDeviceId, newToModelPortId, allowVLANUnion)
}

func (ns *NetService) AddVlan(vlanID string, vlanName string, ipSegmentIDs []string) error {
	return ns.netRepo.AddVlan(vlanID, vlanName, ipSegmentIDs)
}

func (ns *NetService) UpdateVlanIPSegments(vlanId string, ipSegmentIDs []string) error {
	return ns.netRepo.UpdateVlanIPSegments(vlanId, ipSegmentIDs)
}

func (ns *NetService) GetVlans() ([]e.Vlan, error) {
	return ns.netRepo.GetVlans()
}

func (ns *NetService) UpdateVlan(vlanId string, newVlanID string, newVlanName string) error {
	return ns.netRepo.UpdateVlan(vlanId, newVlanID, newVlanName)
}

func (ns *NetService) DeleteVlan(vlanId string) error {
	return ns.netRepo.DeleteVlan(vlanId)
}

// Network scanning method implementations

func (ns *NetService) ScanNetwork(subnet string, options s.ScanOptions) (*s.ScanResult, error) {
	scanner := s.NewSNMPScanner()

	options.Subnet = subnet
	if options.SNMP.Community == "" {
		options.SNMP.Community = "public"
	}
	if options.SNMP.Version == "" {
		options.SNMP.Version = "v2c"
	}

	result, err := scanner.Scan(options)
	if err != nil {
		return nil, fmt.Errorf("network scan failed: %w", err)
	}

	return result, nil
}

func (ns *NetService) ScanDevice(ip string, options s.ScanOptions) (*s.SNMPDevice, error) {
	scanner := s.NewSNMPScanner()

	if options.SNMP.Community == "" {
		options.SNMP.Community = "public"
	}
	if options.SNMP.Version == "" {
		options.SNMP.Version = "v2c"
	}

	device, err := scanner.ScanDevice(ip, options)
	if err != nil {
		return nil, fmt.Errorf("device scan failed: %w", err)
	}

	return device, nil
}

func (ns *NetService) DiscoverDevices(scanResult *s.ScanResult) ([]s.DiscoveredDevice, error) {
	discoverer := s.NewDeviceDiscoverer()

	var devices []s.DiscoveredDevice
	for _, dev := range scanResult.Devices {
		if !dev.Reachable {
			continue
		}
		brand, model, class := discoverer.ClassifyDevice(dev)
		devices = append(devices, s.DiscoveredDevice{
			Device:        dev,
			Brand:         brand,
			Model:         model,
			DeviceClass:   class,
			SuggestedName: discoverer.GenerateDeviceName(dev, class),
			SuggestedZone: discoverer.SuggestZone(dev),
		})
	}

	log.Printf("Discovered %d devices via SNMP", len(devices))
	return devices, nil
}

func (ns *NetService) ImportScanResults(devices []s.DiscoveredDevice, options s.ImportOptions) error {
	return ns.ImportDiscoveredDevices(devices, options)
}

func (ns *NetService) ImportDiscoveredDevices(devices []s.DiscoveredDevice, options s.ImportOptions) error {
	if err := ns.ensureRequiredEntities(options); err != nil {
		return fmt.Errorf("failed to setup required entities: %w", err)
	}

	for _, device := range devices {
		if options.SkipExisting && ns.deviceExists(device.Device.IP) {
			continue
		}

		if err := ns.importSingleDevice(device, options); err != nil {
			log.Printf("Failed to import device %s: %v", device.SuggestedName, err)
			if !options.ReviewMode {
				return err
			}
		}
	}

	return nil
}

func (ns *NetService) ensureRequiredEntities(options s.ImportOptions) error {
	brands, _ := ns.GetBrands()
	brandExists := func(name string) bool {
		for _, b := range brands {
			if b.Name == name {
				return true
			}
		}
		return false
	}

	if !brandExists("Generic") {
		ns.AddBrand("Generic")
	}
	if !brandExists("Discovered") {
		ns.AddBrand("Discovered")
	}
	if !brandExists("Linux") {
		ns.AddBrand("Linux")
	}
	if !brandExists("Cisco") {
		ns.AddBrand("Cisco")
	}

	classes, _ := ns.GetDeviceClasses()
	classExists := func(name string) bool {
		for _, c := range classes {
			if c.Name == name {
				return true
			}
		}
		return false
	}

	requiredClasses := []string{"Switch", "Router", "Server", "Workstation", "Printer", "Generic", "Access Point", "Firewall"}
	for _, className := range requiredClasses {
		if !classExists(className) {
			ns.AddDeviceClass(className)
		}
	}

	proprietaries, _ := ns.GetProperties()
	proprietaryExists := func(name string) bool {
		for _, p := range proprietaries {
			if p.Name == name {
				return true
			}
		}
		return false
	}

	if !proprietaryExists("Discovered") {
		ns.AddProprietary("Discovered")
	}

	if options.CreateZones {
		zones, _ := ns.GetZones()
		zoneExists := func(name string) bool {
			for _, z := range zones {
				if z.Name == name {
					return true
				}
			}
			return false
		}

		requiredZones := []string{"Discovered", "LAN", "Internal", "External", "Private"}
		for _, zoneName := range requiredZones {
			if !zoneExists(zoneName) {
				ns.AddZone(zoneName, "", "", "Discovered", "Office")
			}
		}
	}

	return nil
}

func (ns *NetService) deviceExists(ip string) bool {
	devices, err := ns.GetDevices()
	if err != nil {
		return false
	}

	for _, device := range devices {
		for _, deviceIP := range device.IPs {
			if deviceIP == ip {
				return true
			}
		}
	}

	return false
}

func (ns *NetService) importSingleDevice(discovered s.DiscoveredDevice, options s.ImportOptions) error {
	zoneName := discovered.SuggestedZone
	if options.DefaultZone != "" {
		zoneName = options.DefaultZone
	}

	brandName := discovered.Brand
	if options.DefaultBrand != "" {
		brandName = options.DefaultBrand
	}
	_ = brandName

	err := ns.AddDevice(
		discovered.SuggestedName,
		discovered.Model,
		"",
		zoneName,
		"Discovered",
		[]string{discovered.Device.IP},
	)
	if err != nil {
		return fmt.Errorf("failed to add device: %w", err)
	}

	if len(discovered.Device.Interfaces) > 0 {
		devices, _ := ns.GetDevices()
		var deviceID string
		for _, device := range devices {
			if device.Name == discovered.SuggestedName {
				deviceID = device.ID
				break
			}
		}

		if deviceID != "" {
			if err := ns.createDevicePortsForDevice(deviceID, discovered); err != nil {
				log.Printf("Failed to create ports for device %s: %v", discovered.SuggestedName, err)
			}
		}
	}

	return nil
}

func (ns *NetService) createDevicePortsForDevice(deviceID string, discovered s.DiscoveredDevice) error {
	models, err := ns.GetModels()
	if err != nil {
		return err
	}

	var modelName string
	for _, model := range models {
		if model.Model == discovered.Model {
			modelName = model.Model
			break
		}
	}

	if modelName == "" {
		modelName = discovered.Model
		if modelName == "" {
			modelName = "Generic Model"
		}
		if err := ns.AddModel(modelName, discovered.Brand, discovered.DeviceClass); err != nil {
			return err
		}
	}

	for i, iface := range discovered.Device.Interfaces {
		portName := iface.Name
		if portName == "" {
			portName = fmt.Sprintf("eth%d", i)
		}

		if err := ns.AddModelPort(portName, fmt.Sprintf("%d", i), "0", modelName, false); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}

		modelPorts, _ := ns.GetModelPorts()
		var modelPortID string
		for _, mp := range modelPorts {
			if mp.Name == portName && mp.Model == modelName {
				modelPortID = mp.ID
				break
			}
		}

		if modelPortID == "" {
			continue
		}

		// Build VLAN configs from SNMP data
		vlanConfigs := make([]e.PortVlanConfig, 0, len(iface.VLANs))
		for _, v := range iface.VLANs {
			vlanConfigs = append(vlanConfigs, e.PortVlanConfig{
				VlanNumber: v.VLANNumber,
				Tagged:     v.Tagged,
			})
		}

		if err := ns.AddDevicePort(deviceID, modelPortID, iface.MAC, vlanConfigs); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}

	return nil
}
