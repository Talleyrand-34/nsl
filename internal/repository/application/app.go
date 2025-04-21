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
	GetBrands() []string
	// deviceclass
	AddDeviceClass(brand string) error
	// deviceclass
	GetDeviceClasses() []string
	// deviceclass
	AddZoneType(name string) error
	// deviceclass
	GetZonetypes() []string
	// deviceclass
	AddProprietary(name string) error
	// deviceclass
	GetProperties() []string
	// deviceclass
	AddZone(
		name string,
		father string,
		proprietary string,
		zonename string,
	) error
	// deviceclass
	GetZones() []byte
	// GetZone(name string) int
}

func NewNetService(netRepository d.NetRepository) NetServiceInt {
	return &NetService{netRepo: netRepository}
}

//	func (ns *NetService) GetBrands() []byte {
//	    brands := ns.netRepo.GetBrands()
//	    jsonData, err := json.Marshal(brands)
//	    if err != nil {
//	        log.Printf("Error marshaling brands to JSON: %v", err)
//	        return []byte("[]") // Return empty JSON array on error
//	    }
//	    return jsonData
//	}
func (ns *NetService) AddBrand(brand string) error {
	return ns.netRepo.AddBrand(brand)
}

func (ns *NetService) GetBrands() []string {
	return ns.netRepo.GetBrands()
}

func (ns *NetService) AddDeviceClass(deviceClassName string) error {
	return ns.netRepo.AddDeviceClass(deviceClassName)
}

func (ns *NetService) GetDeviceClasses() []string {
	return ns.netRepo.GetDeviceClasses()
}

func (ns *NetService) AddZoneType(zoneTypeName string) error {
	return ns.netRepo.AddZoneType(zoneTypeName)
}

func (ns *NetService) GetZonetypes() []string {
	return ns.netRepo.GetZonetypes()
}

func (ns *NetService) AddProprietary(proprietary string) error {
	return ns.netRepo.AddProprietary(proprietary)
}

func (ns *NetService) GetProperties() []string {
	return ns.netRepo.GetProperties()
}

func (ns *NetService) AddZone(
	name string,
	father string,
	proprietary string,
	zonename string,
) error {
	return ns.netRepo.AddZone(name, father, proprietary, zonename)
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

// deviceclass
