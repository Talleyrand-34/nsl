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
	e "nsl-graph/internal/repository/entities"
)

type NetService struct {
	netRepo d.NetRepository
}

type NetServiceInt interface {
	// Brand
	AddBrand(brand string) error
	// Brand
	GetBrands() ([]e.Brand, error)
	// deviceclass
	AddDeviceClass(brand string) error
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
		father string,
		fatherid string,
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
	AddDevice(
		label string,
		model string,
		zoneId string,
		zoneName string,
		proprietary string,
	) error
	GetDevices() ([]e.Device, error)
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
	ExportAllStructs() []byte
}

func NewNetService(netRepository d.NetRepository) NetServiceInt {
	return &NetService{netRepo: netRepository}
}

func (ns *NetService) AddBrand(brand string) error {
	return ns.netRepo.AddBrand(brand)
}

func (ns *NetService) GetBrands() ([]e.Brand, error) {
	return ns.netRepo.GetBrands()
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
) error {
	return ns.netRepo.AddModelPort(name, posx, posy, modelName)
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
) error {
	return ns.netRepo.AddDevice(label, model, zoneId, zoneName, proprietary)
}

func (ns *NetService) GetDevicePorts() ([]e.DevicePort, error) {
	zones, err := ns.netRepo.GetDevicePorts()
	return zones, err
}

func (ns *NetService) AddDevicePort(
	deviceid string,
	modelportid string,
) error {
	return ns.netRepo.AddDevicePort(deviceid, modelportid)
}

func (ns *NetService) GetConnections() ([]e.Connection, error) {
	return ns.netRepo.GetConnections()
}

func (ns *NetService) AddConnection(
	fromDevice string,
	fromModelPort string,
	toDevice string,
	toModelPort string,
) error {
	return ns.netRepo.AddConnection(fromDevice, fromModelPort, toDevice, toModelPort)
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
