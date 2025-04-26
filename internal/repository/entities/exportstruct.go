// Package entities contains the structs needed for the processing of sql queries
package entities

type All struct {
	Brands          []BasicBrand
	ConnectionTypes []BasicConnectiontype
	Connections     []BasicConnection
	DeviceClasses   []BasicDeviceclass
	DevicePorts     []BasicDeviceport
	Devices         []BasicDevice
	ModelDevices    []BasicModeldevice
	ModelPorts      []BasicModelport
	Policies        []BasicPolicy
	Proprietaries   []BasicProprietary
	ZoneTypes       []BasicZonetype
	Zones           []BasicZone
}

type BasicBrand struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type BasicConnectiontype struct {
	ID             int64  `json:"id"`
	ConnectionType string `json:"connection_type"`
}

type BasicConnection struct {
	ID                        int64  `json:"id"`
	FromDevicePortModelPortID int64  `json:"from_device_port_model_port_id"`
	FromDevicePortDeviceID    int64  `json:"from_device_port_device_id"`
	FromIPSegment             string `json:"from_ip_segment"` // "" if null
	ToDevicePortModelPortID   int64  `json:"to_device_port_model_port_id"`
	ToDevicePortDeviceID      int64  `json:"to_device_port_device_id"`
	ToIPSegment               string `json:"to_ip_segment"`   // "" if null
	ConnectionType            int64  `json:"connection_type"` // -1 if null
}

type BasicDeviceclass struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type BasicDeviceport struct {
	ModelPortID int64 `json:"model_port_id"`
	DeviceID    int64 `json:"device_id"`
}

type BasicDevice struct {
	ID          int64  `json:"id"`
	Label       string `json:"label"`
	ModelID     int64  `json:"model_id"`
	ZoneID      int64  `json:"zone_id"`     // -1 if null
	Proprietary int64  `json:"proprietary"` // -1 if null
}

type BasicModeldevice struct {
	ID      int64  `json:"id"`
	Model   string `json:"model"`
	Brand   int64  `json:"brand"`
	ClassID int64  `json:"class_id"`
}

type BasicModelport struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Positionx int64  `json:"positionx"`
	Positiony int64  `json:"positiony"`
	ModelID   int64  `json:"model_id"`
}

type BasicPolicy struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	AssociatedConnection int64  `json:"associated_connection"` // -1 if null
	TODO                 string `json:"todo"`                  // "" if null
}

type BasicProprietary struct {
	ID          int64  `json:"id"`
	Proprietary string `json:"proprietary"`
}

type BasicZonetype struct {
	ID           int64  `json:"id"`
	LocationType string `json:"location_type"`
}

type BasicZone struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Father       int64  `json:"father"`        // -1 if null
	Granularity  int64  `json:"granularity"`   // -1 if null
	Proprietary  int64  `json:"proprietary"`   // -1 if null
	LocationType int64  `json:"location_type"` // -1 if null
}
