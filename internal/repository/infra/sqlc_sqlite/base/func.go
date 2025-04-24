package sqlcbase

import e "nsl-graph/internal/repository/entities"

func (r SQLiteRepository) AddBrand(brand string) error {
	return r.basicops.AddBrand(brand)
}

func (r SQLiteRepository) GetBrands() []e.Brand {
	return r.basicops.GetBrands()
}

func (r SQLiteRepository) AddDeviceClass(devclass string) error {
	return r.basicops.AddDeviceClass(devclass)
}

func (r SQLiteRepository) GetDeviceClasses() []e.DevClass {
	return r.basicops.GetDeviceClasses()
}

func (r SQLiteRepository) AddZoneType(name string) error {
	return r.basicops.AddZoneType(name)
}

func (r SQLiteRepository) GetZonetypes() []e.ZoneType {
	return r.basicops.GetZonetypes()
}

func (r SQLiteRepository) AddProprietary(name string) error {
	return r.basicops.AddProprietary(name)
}

func (r SQLiteRepository) GetProperties() []e.Proprietary {
	return r.basicops.GetProperties()
}

func (r SQLiteRepository) AddZone(
	name string,
	fatherid string,
	father string,
	proprietary string,
	zonename string,
) error {
	return r.basicops.AddZone(name, fatherid, father, proprietary, zonename)
}

func (r SQLiteRepository) GetZones() []e.Zone {
	return r.basicops.GetZones()
}

func (r SQLiteRepository) AddModel(
	modelName string,
	brandName string,
	className string,
) error {
	return r.basicops.AddModel(modelName, brandName, className)
}

func (r SQLiteRepository) GetModels() []e.ModelDevice {
	return r.basicops.GetModels()
}

func (r SQLiteRepository) GetDevices() []e.Device {
	return r.basicops.GetDevices()
}

func (r SQLiteRepository) AddDevice(
	label string,
	model string,
	zoneId string,
	zoneName string,
	proprietary string,
) error {
	return r.basicops.AddDevice(label, model, zoneId, zoneName, proprietary)
}

func (r SQLiteRepository) GetModelPorts() []e.ModelPort {
	return r.basicops.GetModelPorts()
}

func (r SQLiteRepository) AddModelPort(
	name string,
	posx string,
	posy string,
	modelName string,
) error {
	return r.basicops.AddModelPort(name, posx, posy, modelName)
}

func (r SQLiteRepository) GetDevicePorts() []e.DevicePort {
	return r.basicops.GetDevicePorts()
}

func (r SQLiteRepository) AddDevicePort(deviceid string, modelportid string) error {
	return r.basicops.AddDevicePort(deviceid, modelportid)
}

func (r SQLiteRepository) GetConnections() []e.Connection {
	return r.basicops.GetConnections()
}

func (r SQLiteRepository) AddConnection(
	fromDevice string,
	fromModelPort string,
	toDevice string,
	toModelPort string,
) error {
	return r.basicops.AddConnection(fromDevice, fromModelPort, toDevice, toModelPort)
}

func (r SQLiteRepository) GetAllPortsAll() []e.DevicePort {
	return r.specops.GetAllPortsAll()
}

func (r SQLiteRepository) GetAllPortsDevice(deviceid string) []e.DevicePort {
	return r.specops.GetAllPortsDevice(deviceid)
}
