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
	Owner        string `json:"owner"`
}

// ModelDevice This struct contains the info about a model
type ModelDevice struct {
	ID        string `json:"id"`
	Model     string `json:"model"`
	Brand     string `json:"brand"`
	ModelType string `json:"model_type"`
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
	Owner       string   `json:"owner"`
	IsUnmanaged bool     `json:"is_unmanaged"`
	IsInvisible bool     `json:"is_invisible"`
	Ips         []string `json:"ips"`
	Profile     string   `json:"profile"` // associated scan-profile name (device or generic)
}

// PortVlanConfig represents a VLAN configuration on a port (tagged or untagged)
type PortVlanConfig struct {
	VlanNumber string `json:"vlan_number"` // VLAN number (e.g., "100")
	Tagged     bool   `json:"tagged"`      // true = tagged, false = untagged
}

// DevicePort This struct contains the info about ports asociated to a device since the information is contained in the model
type DevicePort struct {
	ID          string           `json:"id"`
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

// InterfacePort is a pure join record linking a DeviceInterface to a physical
// DevicePort. VLAN configurations and IP addresses belong to the DeviceInterface
// (resolved via InterfaceID); they are intentionally NOT stored here to avoid
// duplicating the logical interface's data.
type InterfacePort struct {
	ID          string `json:"id"`
	InterfaceID string `json:"interface_id"`
	DeviceID    string `json:"device_id"`
	ModelPortID string `json:"model_port_id"`
}

// ConnectionVlanInfo represents a VLAN with its tagging status on a connection
type ConnectionVlanInfo struct {
	VLANID string `json:"vlan_id"` // VLAN number (e.g., "1", "2")
	Tagged bool   `json:"tagged"`  // true=tagged, false=untagged
}

// Connection This struct contains the info about a connection
type Connection struct {
	ID             string               `json:"id"`
	FromDevice     string               `json:"fromdevice"` // Name of the FromDevice
	FromModelPort  string               `json:"frommodel"`  // Name of the port on the model
	FromZoneName   string               `json:"fromzonename"`
	FromZoneID     string               `json:"fromzoneid"`
	ToDevice       string               `json:"todevice"` // Name of the ToDevice
	ToModelPort    string               `json:"tomodel"`  // Name of the port on the model
	ToZoneName     string               `json:"tozonename"`
	ToZoneID       string               `json:"tozoneid"`
	ConnectionType string               `json:"connection_type,omitempty"` // Name of the connection type (e.g. "ethernet", "wifi")
	Vlans          []ConnectionVlanInfo `json:"vlans,omitempty"`           // Intersection: VLANs present on BOTH ports
	MissingVlans   []ConnectionVlanInfo `json:"missing_vlans,omitempty"`   // Symmetric difference: VLANs on ONLY ONE port (not both)
	DiscoveredVia  []string             `json:"discovered_via,omitempty"`  // provenance: sources that observed this link (e.g. "ssh-lldp@opnsense:igc1")
}

type Vlan struct {
	ID        string `json:"id"`
	VlanID    string `json:"vlanid"`
	VlanName  string `json:"vlanname"`
	IPSegment string `json:"ip_segment"` // Single IP segment associated with this VLAN
}

// Brand represents a brand
type Brand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ScanProfile stores reusable per-host scanning parameters so they need not be
// re-entered on every scan. All fields are stored in clear except SSHPassword
// and SSHKey, which hold AES-256-GCM blobs encrypted under the credential
// vault's data key and are never serialized to API clients (json:"-").
type ScanProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Kind is "device" (default, IP-bound — matched by Host) or "generic" (not
	// bound to a host: only SSH credentials, used as a fallback for any host).
	Kind          string `json:"kind"`
	Host          string `json:"host"`
	SNMPCommunity string `json:"snmp_community"`
	SNMPVersion   string `json:"snmp_version"`
	SNMPPort      int    `json:"snmp_port"`
	TimeoutSec    int    `json:"timeout_sec"`
	ScanSource    string `json:"scan_source"`   // "snmp" | "ssh"
	ConfigSource  string `json:"config_source"` // "none" | "ssh" | "file" | "manual"
	ConfigFile    string `json:"config_file"`
	DeviceType    string `json:"device_type"` // opnsense | openwrt | fortinet | cisco
	SSHUser       string `json:"ssh_user"`
	SSHPassword   string `json:"-"` // encrypted blob; never exposed to clients
	SSHKeyFile    string `json:"ssh_key_file"`
	SSHKey        string `json:"-"` // encrypted PEM private-key content (uploaded); never exposed
	SSHPort       int    `json:"ssh_port"`

	DiscrepancyAction string `json:"discrepancy_action"`
	MergeConfigs      bool   `json:"merge_configs"`
	ConfigTimeout     int    `json:"config_timeout"`
	VLANAccuracy      int    `json:"vlan_accuracy"`

	// HasSSHPassword / HasSSHKey are computed, read-only flags for listing — true
	// when an encrypted SSH password / private key is stored. Set by the service
	// on sanitized reads (the secrets themselves are never serialized).
	HasSSHPassword bool `json:"has_ssh_password"`
	HasSSHKey      bool `json:"has_ssh_key"`
}

// ModelType represents the class of a device
type ModelType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Owner represents a owner of a device o zone
type Owner struct {
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
