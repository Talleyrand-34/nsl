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
}

func NewNetService(netRepository d.NetRepository) NetServiceInt {
	return &NetService{netRepo: netRepository}
}

func (ns *NetService) AddBrand(brand string) error {
	return ns.netRepo.AddBrand(brand)
}

func (ns *NetService) GetBrands() []byte {
	brands := ns.netRepo.GetBrands()
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
	devclass := ns.netRepo.GetDeviceClasses()
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
	zonetypes := ns.netRepo.GetZonetypes()
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
	properties := ns.netRepo.GetProperties()
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
	zones := ns.netRepo.GetZones()
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) GetModels() []byte {
	zones := ns.netRepo.GetModels()
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
	zones := ns.netRepo.GetModelPorts()
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
	zones := ns.netRepo.GetDevices()
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
	zones := ns.netRepo.GetDevicePorts()
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
	zones := ns.netRepo.GetConnections()
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
	zones := ns.netRepo.GetAllPortsAll()
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) GetAllPortsDevice(deviceid string) []byte {
	zones := ns.netRepo.GetAllPortsDevice(deviceid)
	jsonData, err := json.Marshal(zones)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}
