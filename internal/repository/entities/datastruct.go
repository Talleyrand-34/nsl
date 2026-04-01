// Package entities contains the structs needed for the processing of sql queries
//
// datastruct.go: Contains core data structures for SQL query processing.
package entities

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

// Zone This struct containd the info about a zone
type Zone struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	FatherID     string `json:"fatherid"`
	Father       string `json:"father"`
	LocationType string `json:"location_type"`
	Proprietary  string `json:"proprietary"`
}

// ModelDevice This struct contains the info about a model
type ModelDevice struct {
	ID    string `json:"id"`
	Model string `json:"model"`
	Brand string `json:"brand"`
	Class string `json:"class"`
}

// ModelPort This struct contains the info about a port from the model perspective
type ModelPort struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Positionx                int    `json:"positionx"`
	Positiony                int    `json:"positiony"`
	Model                    string `json:"model"`
	Brand                    string `json:"brand"`
	AllowMultipleConnections bool   `json:"allow_multiple_connections"`
	PortType                 string `json:"port_type,omitempty"` // "" / "ethernet" = wired; "wifi" = radio
	Band                     string `json:"band,omitempty"`      // for wifi ports: "2.4GHz", "5GHz", "6GHz"
}

// Device This struct contains the info about a device
type Device struct {
	ID          string   `json:"id"`
	Name        string   `json:"label"`
	Model       string   `json:"model"`
	Brand       string   `json:"brand"`
	ZoneID      string   `json:"zoneid"`
	ZoneName    string   `json:"zonename"`
	ZoneFather  string   `json:"zonefathername"`
	Proprietary string   `json:"proprietary"`
	IPs         []string `json:"ips"` // List of IP addresses (e.g., "192.168.1.1")
}

// PortVlanConfig represents a VLAN configuration on a port (tagged or untagged)
type PortVlanConfig struct {
	VlanNumber string `json:"vlan_number"` // VLAN number (e.g., "100")
	Tagged     bool   `json:"tagged"`      // true = tagged, false = untagged
}

// DevicePort This struct contains the info about ports asociated to a device since the information is contained in the model
type DevicePort struct {
	DeviceID    string           `json:"devid"`
	ModelID     string           `json:"modelid"`
	MacAddress  string           `json:"mac_address"`
	PortName    string           `json:"portname"`
	DevLabel    string           `json:"devname"`
	Positionx   int              `json:"positionx"`
	Positiony   int              `json:"positiony"`
	VlanConfigs []PortVlanConfig `json:"vlan_configs"`
}

// DeviceInterface represents a logical interface (with VLANs) on a device.
// Multiple logical interfaces can be mapped to one or more physical DevicePorts
// via InterfacePort join records.
type DeviceInterface struct {
	ID           string           `json:"id"`
	DeviceID     string           `json:"device_id"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Parent       string           `json:"parent,omitempty"` // physical parent interface for VLAN/subinterfaces
	VlanConfigs  []PortVlanConfig `json:"vlan_configs"`
	IPAddresses  []string         `json:"ip_addresses"`
	WifiSSID     string           `json:"wifi_ssid,omitempty"`
	WifiSecurity string           `json:"wifi_security,omitempty"` // "open", "wpa2", "wpa3"
}

// InterfacePort is a join record linking a DeviceInterface to a physical DevicePort.
// It also stores the VLAN configurations for that specific interface-to-port link.
type InterfacePort struct {
	ID          string           `json:"id"`
	InterfaceID string           `json:"interface_id"`
	DeviceID    string           `json:"device_id"`
	ModelPortID string           `json:"model_port_id"`
	VlanConfigs []PortVlanConfig `json:"vlan_configs"`
}

// Connection This struct contains the info about a connection
type Connection struct {
	ID            string `json:"id"`
	FromDevice    string `json:"fromdevice"` // Name of the FromDevice
	FromModelPort string `json:"frommodel"`  // Name of the port on the model
	FromZoneName  string `json:"fromzonename"`
	FromZoneID    string `json:"fromzoneid"`
	ToDevice      string `json:"todevice"` // Name of the ToDevice
	ToModelPort   string `json:"tomodel"`  // Name of the port on the model
	ToZoneName    string `json:"tozonename"`
	ToZoneID      string `json:"tozoneid"`
}

type Vlan struct {
	ID        string `json:"id"`
	VlanID    string `json:"vlanid"`
	VlanName  string `json:"vlanname"`
	IPSegment string `json:"ip_segment"` // Single IP segment associated with this VLAN
}

// LocalVlan represents a device-specific VLAN name mapping
type LocalVlan struct {
	ID       string `json:"id"`
	VlanID   string `json:"vlanid"`   // VLAN number (e.g., "100")
	DeviceID string `json:"deviceid"` // Device database ID
	VlanName string `json:"vlanname"` // Local VLAN name on this specific device
}

// Brand represents a brand
type Brand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DevClass represents the class of a device
type DevClass struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Proprietary represents a proprietary of a device o zone
type Proprietary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ZoneType represents the type of zone a zone can be
type ZoneType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ConnectionType represents the type of connection a connection can be
type ConnectionType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
