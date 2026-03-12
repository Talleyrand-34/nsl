package scanner

import "time"

type SNMPOptions struct {
	Community string `json:"community"` // default "public"
	Version   string `json:"version"`   // "v1", "v2c" — default "v2c"
	Port      uint16 `json:"port"`      // default 161
}

type ScanOptions struct {
	Subnet  string        `json:"subnet"`
	Timeout time.Duration `json:"timeout"`
	SNMP    SNMPOptions   `json:"snmp"`
}

// DeviceInterface represents one physical or logical interface as reported by the device's ifTable.
type DeviceInterface struct {
	Index       int              `json:"index"`
	Name        string           `json:"name"`        // ifDescr
	MAC         string           `json:"mac"`         // ifPhysAddress
	AdminStatus int              `json:"admin_status"` // 1=up 2=down
	OperStatus  int              `json:"oper_status"`
	IPAddresses []string         `json:"ip_addresses"`
	VLANs       []VLANMembership `json:"vlans"`
}

// VLANMembership describes a VLAN assignment on an interface.
type VLANMembership struct {
	VLANNumber string `json:"vlan_number"`
	Tagged     bool   `json:"tagged"` // false = untagged/native, true = tagged/trunk
}

// DirectNeighbor is a neighbor explicitly advertised by LLDP or CDP — no inference.
type DirectNeighbor struct {
	LocalPort  string `json:"local_port"`
	RemoteIP   string `json:"remote_ip"`
	RemoteMAC  string `json:"remote_mac"`
	RemotePort string `json:"remote_port"`
	RemoteName string `json:"remote_name"`
	Protocol   string `json:"protocol"` // "lldp" or "cdp"
}

// SNMPDevice is the complete result of querying one device via SNMP.
type SNMPDevice struct {
	IP          string            `json:"ip"`
	SysName     string            `json:"sys_name"`
	SysDescr    string            `json:"sys_descr"`
	SysObjectID string            `json:"sys_object_id"`
	SysLocation string            `json:"sys_location"`
	SysContact  string            `json:"sys_contact"`
	Reachable   bool              `json:"reachable"`
	Interfaces  []DeviceInterface `json:"interfaces"`
	Neighbors   []DirectNeighbor  `json:"neighbors"`
}

// ScanResult aggregates all devices discovered in a scan.
type ScanResult struct {
	ID        string       `json:"id"`
	Subnet    string       `json:"subnet"`
	StartTime time.Time    `json:"start_time"`
	EndTime   time.Time    `json:"end_time"`
	Devices   []SNMPDevice `json:"devices"`
}

// DiscoveredDevice pairs an SNMPDevice with classification for import.
type DiscoveredDevice struct {
	Device        SNMPDevice `json:"device"`
	Brand         string     `json:"brand"`
	Model         string     `json:"model"`
	DeviceClass   string     `json:"device_class"`
	SuggestedName string     `json:"suggested_name"`
	SuggestedZone string     `json:"suggested_zone"`
}

// ImportOptions controls device import behaviour.
type ImportOptions struct {
	AutoImport   bool   `json:"auto_import"`
	MergeIPs     bool   `json:"merge_ips"`
	CreateZones  bool   `json:"create_zones"`
	DefaultZone  string `json:"default_zone"`
	DefaultBrand string `json:"default_brand"`
	SkipExisting bool   `json:"skip_existing"`
	ReviewMode   bool   `json:"review_mode"`
}

// NetworkScanner is the interface implemented by SNMPScanner.
type NetworkScanner interface {
	Scan(options ScanOptions) (*ScanResult, error)
	ScanDevice(ip string, options ScanOptions) (*SNMPDevice, error)
}
