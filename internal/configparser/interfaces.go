package configparser

import (
	"time"

	s "nsl-graph/internal/scanner"
)

// ConfigSource defines how device configuration is obtained
type ConfigSource int

const (
	ConfigSourceSSH ConfigSource = iota
	ConfigSourceFile
	ConfigSourceManual
)

func (cs ConfigSource) String() string {
	switch cs {
	case ConfigSourceSSH:
		return "ssh"
	case ConfigSourceFile:
		return "file"
	case ConfigSourceManual:
		return "manual"
	default:
		return "unknown"
	}
}

// DiscrepancyAction defines how to handle conflicts between SNMP and config data
type DiscrepancyAction int

const (
	DiscrepancyActionFail DiscrepancyAction = iota
	DiscrepancyActionPreferSNMP
	DiscrepancyActionPreferConfig
)

func (da DiscrepancyAction) String() string {
	switch da {
	case DiscrepancyActionFail:
		return "fail"
	case DiscrepancyActionPreferSNMP:
		return "prefer-snmp"
	case DiscrepancyActionPreferConfig:
		return "prefer-config"
	default:
		return "unknown"
	}
}

// SSHCredentials contains SSH authentication information
type SSHCredentials struct {
	Username      string        `json:"username"`
	Password      string        `json:"password,omitempty"`
	KeyFile       string        `json:"key_file,omitempty"`
	PrivateKey    string        `json:"private_key,omitempty"` // PEM key content (in-memory; takes precedence over KeyFile)
	KeyPassphrase string        `json:"key_passphrase,omitempty"`
	Port          int           `json:"port"`
	Timeout       time.Duration `json:"timeout"`
}

// ConfigInterface represents a network interface as defined in device configuration
type ConfigInterface struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Enabled     bool         `json:"enabled"`
	IPAddresses []string     `json:"ip_addresses"`
	MACAddress  string       `json:"mac_address,omitempty"`
	MTU         int          `json:"mtu,omitempty"`
	VLANs       []ConfigVLAN `json:"vlans"`
	// Type: "physical", "vlan", "bridge", "tunnel", "wifi-radio", "wifi-iface", etc.
	Type   string `json:"type"`
	Parent string `json:"parent,omitempty"` // Parent interface for VLANs/subinterfaces
	// WiFi fields — populated for "wifi-radio" and "wifi-iface" types
	WifiBand     string `json:"wifi_band,omitempty"`     // "2.4GHz", "5GHz", "6GHz" — wifi-radio only
	WifiSSID     string `json:"wifi_ssid,omitempty"`     // SSID name — wifi-iface only
	WifiSecurity string `json:"wifi_security,omitempty"` // "open", "wpa2", "wpa3" — wifi-iface only
	WifiRadio    string `json:"wifi_radio,omitempty"`    // parent radio name — wifi-iface only
}

// ConfigVLAN represents a VLAN as defined in device configuration
type ConfigVLAN struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	IPRange     string `json:"ip_range,omitempty"` // CIDR notation
	Gateway     string `json:"gateway,omitempty"`
	Tagged      bool   `json:"tagged"`
}

// ConfigRoute represents routing information from configuration
type ConfigRoute struct {
	Network     string `json:"network"`   // Destination network (CIDR)
	Gateway     string `json:"gateway"`   // Next hop
	Interface   string `json:"interface"` // Outgoing interface
	Metric      int    `json:"metric"`    // Route metric/priority
	Description string `json:"description,omitempty"`
}

// ConfigFirewallRule represents a firewall rule from configuration
type ConfigFirewallRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Enabled     bool     `json:"enabled"`
	Action      string   `json:"action"`    // "allow", "deny", "reject"
	Direction   string   `json:"direction"` // "in", "out", "forward"
	SourceZone  string   `json:"source_zone,omitempty"`
	DestZone    string   `json:"dest_zone,omitempty"`
	Source      []string `json:"source"`      // IP addresses/networks
	Destination []string `json:"destination"` // IP addresses/networks
	Ports       []string `json:"ports"`       // Port numbers/ranges
	Protocol    string   `json:"protocol"`    // "tcp", "udp", "icmp", etc.
}

// PortVLANInfo represents VLAN membership for a switch port
type PortVLANInfo struct {
	VID    string `json:"vid"`    // VLAN ID (e.g., "1", "60")
	Tagged bool   `json:"tagged"` // true if port is tagged on this VLAN
}

// SwitchPortInfo represents physical switch port information from board.json/swconfig
type SwitchPortInfo struct {
	PortNumber int            `json:"port_number"` // Port index (0, 1, 2, etc.)
	PortName   string         `json:"port_name"`   // Human-readable name (e.g., "lan1", "wan0")
	LinkStatus string         `json:"link_status"` // "up", "down"
	Role       string         `json:"role"`        // "lan", "wan"
	VLANs      []PortVLANInfo `json:"vlans"`       // VLAN membership per port (from swconfig)
}

// ConfigData contains the parsed configuration data from a network device
type ConfigData struct {
	DeviceType    string               `json:"device_type"`  // "opnsense", "openwrt", "fortinet", "cisco"
	DeviceModel   string               `json:"device_model"` // Specific model/version info
	Hostname      string               `json:"hostname"`
	Domain        string               `json:"domain,omitempty"`
	ConfigVersion string               `json:"config_version"` // Configuration version/timestamp
	Source        ConfigSource         `json:"source"`         // How config was obtained
	Interfaces    []ConfigInterface    `json:"interfaces"`
	VLANs         []ConfigVLAN         `json:"vlans"`
	Routes        []ConfigRoute        `json:"routes"`
	FirewallRules []ConfigFirewallRule `json:"firewall_rules"`
	SwitchPorts   []SwitchPortInfo     `json:"switch_ports,omitempty"` // Physical switch ports (OpenWrt)
	Raw           string               `json:"raw,omitempty"`          // Raw configuration content
	ParsedAt      time.Time            `json:"parsed_at"`
}

// ConfigDiscrepancy represents a conflict between SNMP and configuration data
type ConfigDiscrepancy struct {
	Type        string `json:"type"`         // "interface", "vlan", "ip", "status"
	Object      string `json:"object"`       // Object identifier (interface name, VLAN ID, etc.)
	SNMPValue   string `json:"snmp_value"`   // Value from SNMP
	ConfigValue string `json:"config_value"` // Value from configuration
	Description string `json:"description"`  // Human-readable description
	Severity    string `json:"severity"`     // "warning", "error"
}

// EnhancedDiscoveredDevice extends scanner.DiscoveredDevice with configuration data
type EnhancedDiscoveredDevice struct {
	s.DiscoveredDevice                     // Embedded original struct
	ConfigData         *ConfigData         `json:"config_data,omitempty"`
	Discrepancies      []ConfigDiscrepancy `json:"discrepancies,omitempty"`
	EnhancedBy         []string            `json:"enhanced_by"` // Sources of enhancement: ["snmp", "config"]
}

// ConfigParser defines the interface that all device-specific configuration parsers must implement
type ConfigParser interface {
	// GetDeviceType returns the device type this parser handles (e.g., "opnsense", "fortinet")
	GetDeviceType() string

	// SupportsDevice returns true if this parser can handle the given device
	SupportsDevice(device s.SNMPDevice) bool

	// ParseConfig parses raw configuration data and returns structured ConfigData
	ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*ConfigData, error)

	// GetConfigViaSSH retrieves configuration from device via SSH
	GetConfigViaSSH(ip string, creds SSHCredentials) (string, error)

	// ValidateConfig performs basic validation on parsed configuration
	ValidateConfig(config *ConfigData) []error
}

// ConfigParserOptions contains options for configuration parsing
type ConfigParserOptions struct {
	Source            ConfigSource      `json:"source"`
	FilePath          string            `json:"file_path,omitempty"`
	DeviceType        string            `json:"device_type,omitempty"` // Manual OS override (opnsense, openwrt, fortinet, cisco)
	SSHCredentials    *SSHCredentials   `json:"ssh_credentials,omitempty"`
	DiscrepancyAction DiscrepancyAction `json:"discrepancy_action"`
	MergeWithSNMP     bool              `json:"merge_with_snmp"`
	ParseTimeout      time.Duration     `json:"parse_timeout"`
}

// ConfigParserRegistry manages available configuration parsers
type ConfigParserRegistry struct {
	parsers map[string]ConfigParser
}

// NewConfigParserRegistry creates a new parser registry
func NewConfigParserRegistry() *ConfigParserRegistry {
	return &ConfigParserRegistry{
		parsers: make(map[string]ConfigParser),
	}
}

// RegisterParser adds a parser to the registry
func (r *ConfigParserRegistry) RegisterParser(parser ConfigParser) {
	r.parsers[parser.GetDeviceType()] = parser
}

// GetParser returns the appropriate parser for a device type
func (r *ConfigParserRegistry) GetParser(deviceType string) (ConfigParser, bool) {
	parser, exists := r.parsers[deviceType]
	return parser, exists
}

// GetParserForDevice returns the appropriate parser for a discovered device
func (r *ConfigParserRegistry) GetParserForDevice(device s.SNMPDevice) (ConfigParser, bool) {
	for _, parser := range r.parsers {
		if parser.SupportsDevice(device) {
			return parser, true
		}
	}
	return nil, false
}

// GetParserForDeviceWithType returns parser by manual type or auto-detection
func (r *ConfigParserRegistry) GetParserForDeviceWithType(device s.SNMPDevice, deviceType string) (ConfigParser, bool) {
	// Use manual type if specified and valid
	if deviceType != "" {
		if parser, exists := r.parsers[deviceType]; exists {
			return parser, true
		}
	}

	// Fallback to auto-detection
	return r.GetParserForDevice(device)
}

// ListParsers returns all registered parser device types
func (r *ConfigParserRegistry) ListParsers() []string {
	var types []string
	for deviceType := range r.parsers {
		types = append(types, deviceType)
	}
	return types
}

// DefaultRegistry is the global parser registry. Parsers self-register via init().
var DefaultRegistry = NewConfigParserRegistry()
