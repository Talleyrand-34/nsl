// Comment
package repository

// Comment
type NetRepository interface {
	// Brand
	AddBrand(brand string) error
	// Brand
	GetBrands() []string
	// GetBrandByID(id int) (string, error)

	// UpdateBrand(id int, brand string) error
	// DeleteBrand(id int) error

	// DeviceClasses
	// AddDeviceClass(name string) error
	// GetDeviceClasses() []string
	// GetDeviceClassByID(id int) (string, error)
	// UpdateDeviceClass(id int, name string) error
	// DeleteDeviceClass(id int) error

	// // ModelDevices
	// AddModelDevice(model string, brandID int, classID int) error
	// GetModelDevices() []string
	// GetModelDeviceByID(id int) (string, int, int, error) // model, brandID, classID
	// UpdateModelDevice(id int, model string, brandID int, classID int) error
	// DeleteModelDevice(id int) error
	//
	// // ModelPorts
	// AddModelPort(name string, positionX int, positionY int, modelID int) error
	// GetModelPorts() []string
	// GetModelPortByID(id int) (string, int, int, int, error) // name, posX, posY, modelID
	// UpdateModelPort(id int, name string, positionX int, positionY int, modelID int) error
	// DeleteModelPort(id int) error
	//
	// // Devices
	// AddDevice(label string, modelID int, zoneID int, propietary int) error
	// GetDevices() []string
	// GetDeviceByID(id int) (string, int, int, int, error) // label, modelID, zoneID, propietary
	// UpdateDevice(id int, label string, modelID int, zoneID int, propietary int) error
	// DeleteDevice(id int) error
	//
	// // Property
	// AddProperty(name string) error
	// GetProperties() []string
	// GetPropertyByID(id int) (string, error)
	// UpdateProperty(id int, name string) error
	// DeleteProperty(id int) error
	//
	// // DevicePorts
	// AddDevicePort(modelPortID int, deviceID int) error
	// GetDevicePorts() [][2]int // slice of [modelPortID, deviceID]
	// DeleteDevicePort(modelPortID int, deviceID int) error
	//
	// // Connections
	// AddConnection(
	// 	fromModelPortID int,
	// 	fromDeviceID int,
	// 	fromIPSegment string,
	// 	toModelPortID int,
	// 	toDeviceID int,
	// 	toIPSegment string,
	// 	connectionType int,
	// ) error
	// GetConnections() []int // IDs
	// GetConnectionByID(id int) (
	// 	int, int, string, int, int, string, int, error,
	// ) // fromModelPortID, fromDeviceID, fromIPSegment, toModelPortID, toDeviceID, toIPSegment, connectionType
	// UpdateConnection(
	// 	id int,
	// 	fromModelPortID int,
	// 	fromDeviceID int,
	// 	fromIPSegment string,
	// 	toModelPortID int,
	// 	toDeviceID int,
	// 	toIPSegment string,
	// 	connectionType int,
	// ) error
	// DeleteConnection(id int) error
	//
	// // ConnectionType
	// AddConnectionType(connectionType string) error
	// GetConnectionTypes() []string
	// GetConnectionTypeByID(id int) (string, error)
	// UpdateConnectionType(id int, connectionType string) error
	// DeleteConnectionType(id int) error
	//
	// // Zone
	// AddZone(
	// 	name string,
	// 	father int,
	// 	granularity int,
	// 	proprietary int,
	// 	locationType string,
	// ) error
	// GetZones() []string
	// GetZoneByID(
	// 	id int,
	// ) (string, int, int, int, string, error) // name, father, granularity, proprietary, locationType
	// UpdateZone(
	// 	id int,
	// 	name string,
	// 	father int,
	// 	granularity int,
	// 	proprietary int,
	// 	locationType string,
	// ) error
	// DeleteZone(id int) error
	//
	// // ZoneType
	// AddZoneType(locationType string) error
	// GetZoneTypes() []string
	// GetZoneTypeByID(id int) (string, error)
	// UpdateZoneType(id int, locationType string) error
	// DeleteZoneType(id int) error
	//
	// // Policies
	// AddPolicy(name string, description string, associatedConnection int, todo string) error
	// GetPolicies() []string
	// GetPolicyByID(
	// 	id int,
	// ) (string, string, int, string, error) // name, description, associatedConnection, todo
	// UpdatePolicy(
	// 	id int,
	// 	name string,
	// 	description string,
	// 	associatedConnection int,
	// 	todo string,
	// ) error
	// DeletePolicy(id int) error
}
