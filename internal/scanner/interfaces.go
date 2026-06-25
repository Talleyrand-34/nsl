package scanner

import (
	"strings"
	"time"
)

type SNMPOptions struct {
	Community string `json:"community"` // default "public"
	Version   string `json:"version"`   // "v1", "v2c" — default "v2c"
	Port      uint16 `json:"port"`      // default 161
}

type ScanOptions struct {
	Subnet  string        `json:"subnet"`
	Timeout time.Duration `json:"timeout"`
	SNMP    SNMPOptions   `json:"snmp"`
	// OnProgress, when set, is called once per host as a subnet sweep completes
	// (concurrently) so callers can report live progress. It must be safe for
	// concurrent use. Not serialized.
	OnProgress func(done, total int, ip string, reachable bool) `json:"-"`
}

// SNMP ifType values for interface classification
const (
	IfTypeEthernetCsmacd = 6   // physical Ethernet port
	IfTypeLoopback       = 24  // software loopback
	IfTypePropVirtual    = 53  // VLAN sub-interfaces and other virtual interfaces
	IfTypeIEEE80211      = 71  // IEEE 802.11 wireless radio
	IfTypeTunnel         = 131 // tunnel interfaces
	IfTypeLag            = 161 // IEEE 802.3ad Link Aggregation (bond)
)

// DeviceInterface represents one physical or logical interface as reported by the device's ifTable.
type DeviceInterface struct {
	Index        int               `json:"index"`
	IfType       int               `json:"if_type"`      // SNMP ifType value
	Name         string            `json:"name"`         // ifDescr
	MAC          string            `json:"mac"`          // ifPhysAddress
	AdminStatus  int               `json:"admin_status"` // 1=up 2=down
	OperStatus   int               `json:"oper_status"`
	IPAddresses  []string          `json:"ip_addresses"`
	IPNetmasks   map[string]string `json:"ip_netmasks,omitempty"` // ip -> netmask; omitted when empty so a JSON round-trip (e.g. via PHP) can't turn {} into []
	VLANs        []VLANMembership  `json:"vlans"`
	Parent       string            `json:"parent,omitempty"`        // physical parent for VLAN/subinterfaces
	WifiBand     string            `json:"wifi_band,omitempty"`     // "2.4GHz", "5GHz", "6GHz" for wifi radios
	WifiSSID     string            `json:"wifi_ssid,omitempty"`     // SSID for wifi-iface type
	WifiSecurity string            `json:"wifi_security,omitempty"` // "open", "wpa2", "wpa3"
	IsBridge     bool              `json:"is_bridge,omitempty"`     // true if this is a bridge interface (e.g., eth0 on OpenWrt DSA)
}

// IsPhysicalPort returns true if the interface represents a physical port.
// It checks both ifType AND interface naming conventions since some devices
// (e.g., OpenWrt) report all interfaces with the same ifType.
func (d *DeviceInterface) IsPhysicalPort() bool {
	if d.IsBridge {
		return false
	}
	// First check ifType for devices that properly classify interfaces
	if d.IfType == IfTypeEthernetCsmacd || d.IfType == IfTypeLag || d.IfType == IfTypeIEEE80211 {
		// Further validate by checking naming conventions for known non-physical patterns
		if isLogicalInterfaceName(d.Name) {
			return false
		}
		return true
	}
	return false
}

// isLogicalInterfaceName checks if an interface name follows logical/virtual patterns
func isLogicalInterfaceName(name string) bool {
	// Bridge interfaces (Linux/OpenWrt)
	if strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "bridge") {
		return true
	}
	// VLAN subinterfaces (e.g., eth0.10, eth1.20, ens0.100)
	if strings.Contains(name, ".") && isVLANSubinterface(name) {
		return true
	}
	// OpenVPN interfaces
	if strings.HasPrefix(name, "ovpns") || strings.HasPrefix(name, "tun") ||
		strings.HasPrefix(name, "tap") {
		return true
	}
	// WireGuard
	if strings.HasPrefix(name, "wg") {
		return true
	}
	// ZeroTier
	if strings.HasPrefix(name, "zt") || strings.HasPrefix(name, "zt+") {
		return true
	}
	// Vlan prefix (BSD-style: vlan0, vlan10)
	if strings.HasPrefix(name, "vlan") {
		return true
	}
	// GRE/GRE6 tunnels
	if strings.HasPrefix(name, "gre") || strings.HasPrefix(name, "gretap") {
		return true
	}
	// VXLAN
	if strings.HasPrefix(name, "vxlan") {
		return true
	}
	if strings.HasPrefix(name, "lagg") {
		return true
	}
	// Bond interfaces (not LAG, but bonding slaves)
	if strings.HasPrefix(name, "bond") || strings.HasPrefix(name, "sl") {
		return true
	}
	// WiFi AP interfaces (e.g., phy0-ap0, phy1-ap1) - these are logical, not physical
	if isWifiAPInterface(name) {
		return true
	}
	return false
}

// isWifiAPInterface checks if the interface name is a WiFi AP interface
// e.g., phy0-ap0, phy1-ap1, wlan0, wlan0-1
func isWifiAPInterface(name string) bool {
	// OpenWrt WiFi AP interfaces: phy0-ap0, phy1-ap1, etc.
	if strings.HasPrefix(name, "phy") {
		// Match pattern: phy[0-9]-ap[0-9]+
		parts := strings.Split(name, "-")
		if len(parts) == 2 && strings.HasPrefix(parts[1], "ap") {
			return true
		}
	}
	// Standard wireless interfaces: wlan0, wlan0-1, wl0, etc.
	if strings.HasPrefix(name, "wlan") || strings.HasPrefix(name, "wl") ||
		strings.HasPrefix(name, "wifi") {
		return true
	}
	return false
}

// isWifiRadioName checks if the interface name is a WiFi radio (physical)
// e.g., phy0, phy1, wlan0, wlan0-1 (without -ap suffix)
func isWifiRadioName(name string) bool {
	// OpenWrt WiFi radios: phy0, phy1, etc.
	if strings.HasPrefix(name, "phy") && !strings.Contains(name, "-") {
		return true
	}
	// Standard wireless interfaces without AP suffix
	if strings.HasPrefix(name, "wlan") || strings.HasPrefix(name, "wl") ||
		strings.HasPrefix(name, "wifi") {
		// Exclude interfaces that are AP interfaces
		if !strings.Contains(name, "-ap") {
			return true
		}
	}
	return false
}

// isVLANSubinterface checks if the dot-separated name represents a VLAN subinterface
func isVLANSubinterface(name string) bool {
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return false
	}
	// First part should look like a physical interface name
	baseName := parts[0]
	vlanID := parts[1]

	// Base name should not be empty and should look like an interface
	if baseName == "" || len(baseName) < 2 {
		return false
	}

	// VLAN ID should be numeric
	for _, c := range vlanID {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// IsWifiRadio returns true if the interface is a WiFi radio port.
// It checks both ifType and naming conventions (e.g., phy0, phy1).
func (d *DeviceInterface) IsWifiRadio() bool {
	// Check ifType for devices that properly classify WiFi
	if d.IfType == IfTypeIEEE80211 {
		return true
	}
	// Also check naming conventions for OpenWrt-style radios
	if isWifiRadioName(d.Name) {
		return true
	}
	return false
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

// FDBEntry is one learned MAC in a bridge forwarding database: the MAC was seen
// on the local bridge port Port (an ifDescr), optionally on VLAN.
type FDBEntry struct {
	MAC     string `json:"mac"`
	Port    string `json:"port"` // local interface name (ifDescr) the MAC was learned on
	IfIndex int    `json:"if_index,omitempty"`
	VLAN    string `json:"vlan,omitempty"`
}

// SNMPDevice is the complete result of querying one device via SNMP.
type SNMPDevice struct {
	IP          string             `json:"ip"`
	SysName     string             `json:"sys_name"`
	SysDescr    string             `json:"sys_descr"`
	SysObjectID string             `json:"sys_object_id"`
	SysLocation string             `json:"sys_location"`
	SysContact  string             `json:"sys_contact"`
	Reachable   bool               `json:"reachable"`
	Interfaces  []DeviceInterface  `json:"interfaces"`
	Neighbors   []DirectNeighbor   `json:"neighbors"`
	SwitchPorts []PhysicalPortInfo `json:"switch_ports,omitempty"` // Physical switch ports (OpenWrt)
	BridgeFDB   []FDBEntry         `json:"bridge_fdb,omitempty"`   // learned MAC→port table (managed switches)
}

// ScanResult aggregates all devices discovered in a scan.
type ScanResult struct {
	ID        string       `json:"id"`
	Subnet    string       `json:"subnet"`
	StartTime time.Time    `json:"start_time"`
	EndTime   time.Time    `json:"end_time"`
	Devices   []SNMPDevice `json:"devices"`
}

// PhysicalPortInfo represents a physical switch port from device configuration (e.g., OpenWrt board.json)
type PhysicalPortInfo struct {
	Name       string         `json:"name"`        // "lan1", "wan0"
	PortNumber int            `json:"port_number"` // 2, 3, 5, 4 (switch port number)
	Role       string         `json:"role"`        // "lan", "wan"
	LinkStatus string         `json:"link_status"` // "up", "down"
	MAC        string         `json:"mac"`         // MAC address if available
	VLANs      []PortVLANInfo `json:"vlans"`       // VLAN membership per port (from swconfig)
}

// PortVLANInfo represents VLAN membership for a switch port
type PortVLANInfo struct {
	VID    string `json:"vid"`    // VLAN ID (e.g., "1", "60")
	Tagged bool   `json:"tagged"` // true if port is tagged on this VLAN
}

// DiscoveredDevice pairs an SNMPDevice with classification for import.
type DiscoveredDevice struct {
	Device        SNMPDevice `json:"device"`
	Brand         string     `json:"brand"`
	Model         string     `json:"model"`
	DeviceClass   string     `json:"device_class"`
	SuggestedName string     `json:"suggested_name"`
	SuggestedZone string     `json:"suggested_zone"`
	Profile       string     `json:"profile,omitempty"` // scan profile used to discover it (tied on import)
}

// DiscoveredDeviceInfo holds ALL discovered information from a scan.
// This is the canonical data structure passed between discovery, display, and import phases.
type DiscoveredDeviceInfo struct {
	IP            string `json:"ip"`
	SysName       string `json:"sys_name"`
	SysDescr      string `json:"sys_descr"`
	Brand         string `json:"brand"`
	Model         string `json:"model"`
	DeviceClass   string `json:"device_class"`
	SuggestedName string `json:"suggested_name"`
	SuggestedZone string `json:"suggested_zone"`

	// Physical ports from configuration (e.g., board.json switch ports)
	PhysicalPorts []PhysicalPortInfo `json:"physical_ports"`

	// Interfaces - logical interfaces only (bridges, VLAN subinterfaces, WiFi, etc.)
	// Physical ports are NOT included here
	Interfaces []DeviceInterface `json:"interfaces"`

	// Source tracking
	Source string `json:"source"` // "snmp", "ssh", "merged"
}

// IPVLANMapping represents a proposed mapping between an IP address and a VLAN
type IPVLANMapping struct {
	IP           string `json:"ip"`
	Subnet       string `json:"subnet"`        // network CIDR derived from SNMP netmask (e.g. "192.168.1.0/24")
	VLANNumber   string `json:"vlan_number"`   // VLAN ID (can be negative for special cases)
	Confidence   string `json:"confidence"`    // "exact", "heuristic", "suggested", "unknown"
	Reason       string `json:"reason"`        // Human-readable explanation for the mapping
	IsNewVLAN    bool   `json:"is_new_vlan"`   // true if this mapping would create a new VLAN
	OriginalVLAN string `json:"original_vlan"` // Original VLAN from interface if any
}

// InterfaceImportPlan represents the planned import actions for a single interface
type InterfaceImportPlan struct {
	Interface     DeviceInterface `json:"interface"`
	IPMappings    []IPVLANMapping `json:"ip_mappings"`
	VLANsToCreate []VLANPlan      `json:"vlans_to_create"`
	VLANsToUpdate []VLANPlan      `json:"vlans_to_update"`
}

// VLANPlan represents a VLAN that will be created or updated
type VLANPlan struct {
	VLANNumber       string   `json:"vlan_number"`
	VLANName         string   `json:"vlan_name"`
	IPSegmentIDs     []string `json:"ip_segment_ids"`
	Action           string   `json:"action"` // "create" or "update"
	ExistingSegments []string `json:"existing_segments,omitempty"`
}

// DeviceImportPlan represents the complete import plan for a device
type DeviceImportPlan struct {
	Device         DiscoveredDevice      `json:"device"`
	InterfacePlans []InterfaceImportPlan `json:"interface_plans"`
	PhysicalPlans  []PhysicalPortPlan    `json:"physical_plans,omitempty"`
	HasConflicts   bool                  `json:"has_conflicts"`
	RequiresInput  bool                  `json:"requires_input"`
	Summary        string                `json:"summary"`
}

// PhysicalPortPlan represents a physical port to be created during import
type PhysicalPortPlan struct {
	Port  PhysicalPortInfo `json:"port"`
	VLANs []VLANPlan       `json:"vlans"`
}

// MappingAction represents user choices for import plans
type MappingAction int

const (
	ActionApprove MappingAction = iota
	ActionEdit
	ActionSkip
	ActionSkipDevice
	ActionQuit
)

// ImportOptions controls device import behaviour.
type ImportOptions struct {
	AutoImport           bool   `json:"auto_import"`
	MergeIPs             bool   `json:"merge_ips"`
	CreateZones          bool   `json:"create_zones"`
	DefaultZone          string `json:"default_zone"`
	DefaultBrand         string `json:"default_brand"`
	SkipExisting         bool   `json:"skip_existing"`
	ReviewMode           bool   `json:"review_mode"`
	InteractiveVLANs     bool   `json:"interactive_vlans"`      // New: Always confirm VLAN mappings
	AutoApproveHeuristic bool   `json:"auto_approve_heuristic"` // New: Auto-approve high-confidence mappings
	VLANAccuracyLevel    int    `json:"vlan_accuracy_level"`    // New: VLAN detection accuracy level (1=interface names only, 2=include IP heuristics)

	// Configuration parsing options
	ConfigSource       string `json:"config_source"`          // "none", "ssh", "file", "manual"
	ConfigFile         string `json:"config_file,omitempty"`  // Path to config file when using file source
	DeviceType         string `json:"device_type,omitempty"`  // Device OS type override (opnsense, openwrt, fortinet, cisco)
	SSHUsername        string `json:"ssh_username,omitempty"` // SSH username for config retrieval
	SSHPassword        string `json:"ssh_password,omitempty"` // SSH password for config retrieval
	SSHKeyFile         string `json:"ssh_key_file,omitempty"` // SSH private key file path
	SSHPort            int    `json:"ssh_port,omitempty"`     // SSH port (default: 22)
	DiscrepancyAction  string `json:"discrepancy_action"`     // "fail", "prefer-snmp", "prefer-config"
	MergeWithConfig    bool   `json:"merge_with_config"`      // Merge SNMP data with config data
	ParseConfigTimeout int    `json:"parse_config_timeout"`   // Timeout for config parsing in seconds
}

// RequiresUserInput checks if an interface plan needs user input
func (plan *InterfaceImportPlan) RequiresUserInput() bool {
	for _, mapping := range plan.IPMappings {
		if mapping.Confidence == "unknown" || mapping.Confidence == "suggested" {
			return true
		}
	}
	return false
}

// NetworkScanner is the interface implemented by SNMPScanner.
type NetworkScanner interface {
	Scan(options ScanOptions) (*ScanResult, error)
	ScanDevice(ip string, options ScanOptions) (*SNMPDevice, error)
}
