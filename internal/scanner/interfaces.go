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

// IPVLANMapping represents a proposed mapping between an IP address and a VLAN
type IPVLANMapping struct {
	IP           string `json:"ip"`
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
	VLANNumber     string   `json:"vlan_number"`
	VLANName       string   `json:"vlan_name"`
	IPSegmentIDs   []string `json:"ip_segment_ids"`
	Action         string   `json:"action"` // "create" or "update"
	ExistingSegments []string `json:"existing_segments,omitempty"`
}

// DeviceImportPlan represents the complete import plan for a device
type DeviceImportPlan struct {
	Device         DiscoveredDevice      `json:"device"`
	InterfacePlans []InterfaceImportPlan `json:"interface_plans"`
	HasConflicts   bool                  `json:"has_conflicts"`
	RequiresInput  bool                  `json:"requires_input"`
	Summary        string                `json:"summary"`
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
	InteractiveVLANs     bool   `json:"interactive_vlans"`     // New: Always confirm VLAN mappings
	AutoApproveHeuristic bool   `json:"auto_approve_heuristic"` // New: Auto-approve high-confidence mappings
	VLANAccuracyLevel    int    `json:"vlan_accuracy_level"`   // New: VLAN detection accuracy level (1=interface names only, 2=include IP heuristics)

	// Configuration parsing options
	ConfigSource         string `json:"config_source"`          // "none", "ssh", "file", "manual"
	ConfigFile           string `json:"config_file,omitempty"`  // Path to config file when using file source
	DeviceType           string `json:"device_type,omitempty"`  // Device OS type override (opnsense, openwrt, fortinet, cisco)
	SSHUsername          string `json:"ssh_username,omitempty"` // SSH username for config retrieval
	SSHPassword          string `json:"ssh_password,omitempty"` // SSH password for config retrieval
	SSHKeyFile           string `json:"ssh_key_file,omitempty"` // SSH private key file path
	SSHPort              int    `json:"ssh_port,omitempty"`     // SSH port (default: 22)
	DiscrepancyAction    string `json:"discrepancy_action"`     // "fail", "prefer-snmp", "prefer-config"
	MergeWithConfig      bool   `json:"merge_with_config"`      // Merge SNMP data with config data
	ParseConfigTimeout   int    `json:"parse_config_timeout"`   // Timeout for config parsing in seconds
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
