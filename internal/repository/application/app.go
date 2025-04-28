
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
	"log"

	d "nsl-graph/internal/repository/domain"
)

type NetService struct {
	netRepo d.NetRepository
}

type NetServiceInt interface {
	// Brand
	AddBrand(brand string) error
	// Brand
	GetBrands() []byte
	// deviceclass
	AddDeviceClass(brand string) error
	// deviceclass
	GetDeviceClasses() []byte
	// deviceclass
	AddZoneType(name string) error
	// deviceclass
	GetZonetypes() []byte
	// deviceclass
	AddProprietary(name string) error
	// deviceclass
	GetProperties() []byte
	// deviceclass
	AddZone(
		name string,
		father string,
		fatherid string,
		proprietary string,
		zonename string,
	) error
	// deviceclass
	GetZones() []byte
	// GetZone(name string) int
	AddModel(
		modelName string,
		brandName string,
		className string,
	) error

	GetModels() []byte
	AddDevice(
		label string,
		model string,
		zoneId string,
		zoneName string,
		proprietary string,
	) error
	GetDevices() []byte
	GetModelPorts() []byte

	AddModelPort(
		name string,
		posx string,
		posy string,
		modelName string,
	) error
	GetDevicePorts() []byte
	AddDevicePort(deviceid string, modelportid string) error
	GetConnections() []byte
	AddConnection(
		fromDevice string,
		fromModelPort string,
		toDevice string,
		toModelPort string,
	) error
	GetAllPortsDevice(deviceid string) []byte
	GetAllPortsAll() []byte
	ExportAllStructs() []byte
}

func NewNetService(netRepository d.NetRepository) NetServiceInt {
	return &NetService{netRepo: netRepository}
}

func (ns *NetService) AddBrand(brand string) error {
	return ns.netRepo.AddBrand(brand)
}

func (ns *NetService) GetBrands() []byte {
	brands, err := ns.netRepo.GetBrands()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(brands)
	if err != nil {
		log.Printf("Error marshaling properties to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) AddDeviceClass(deviceClassName string) error {
	return ns.netRepo.AddDeviceClass(deviceClassName)
}

func (ns *NetService) GetDeviceClasses() []byte {
	devclass, err := ns.netRepo.GetDeviceClasses()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(devclass)
	if err != nil {
		log.Printf("Error marshaling properties to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) AddZoneType(zoneTypeName string) error {
	return ns.netRepo.AddZoneType(zoneTypeName)
}

func (ns *NetService) GetZonetypes() []byte {
	zonetypes, err := ns.netRepo.GetZonetypes()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(zonetypes)
	if err != nil {
		log.Printf("Error marshaling properties to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) AddProprietary(proprietary string) error {
	return ns.netRepo.AddProprietary(proprietary)
}

func (ns *NetService) GetProperties() []byte {
	properties, err := ns.netRepo.GetProperties()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(properties)
	if err != nil {
		log.Printf("Error marshaling properties to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
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

func (ns *NetService) GetZones() []byte {
	zones, err := ns.netRepo.GetZones()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) GetModels() []byte {
	zones, err := ns.netRepo.GetModels()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) AddModelPort(
	name string,
	posx string,
	posy string,
	modelName string,
) error {
	return ns.netRepo.AddModelPort(name, posx, posy, modelName)
}

func (ns *NetService) GetModelPorts() []byte {
	zones, err := ns.netRepo.GetModelPorts()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) AddModel(
	modelName string,
	brandName string,
	className string,
) error {
	return ns.netRepo.AddModel(modelName, brandName, className)
}

func (ns *NetService) GetDevices() []byte {
	zones, err := ns.netRepo.GetDevices()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) AddDevice(
	label string,
	model string,
	zoneId string,
	zoneName string,
	proprietary string,
) error {
	return ns.netRepo.AddDevice(label, model, zoneId, zoneName, proprietary)
}

func (ns *NetService) GetDevicePorts() []byte {
	zones, err := ns.netRepo.GetDevicePorts()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) AddDevicePort(
	deviceid string,
	modelportid string,
) error {
	return ns.netRepo.AddDevicePort(deviceid, modelportid)
}

func (ns *NetService) GetConnections() []byte {
	zones, err := ns.netRepo.GetConnections()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) AddConnection(
	fromDevice string,
	fromModelPort string,
	toDevice string,
	toModelPort string,
) error {
	return ns.netRepo.AddConnection(fromDevice, fromModelPort, toDevice, toModelPort)
}

func (ns *NetService) GetAllPortsAll() []byte {
	zones, err := ns.netRepo.GetAllPortsAll()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) GetAllPortsDevice(deviceid string) []byte {
	zones, err := ns.netRepo.GetAllPortsDevice(deviceid)
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
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
