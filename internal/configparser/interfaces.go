package configparser

import (
	"sort"
	"strings"
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
	Network   string `json:"network"`   // Destination network (CIDR)
	Gateway   string `json:"gateway"`   // Next hop
	Interface string `json:"interface"` // Outgoing interface
	Metric    int    `json:"metric"`    // Route metric/priority
	// Protocol records how the route was learned: "static", "connected",
	// "ospf", "bgp", … Empty means the source did not say, which for a
	// configuration file (as opposed to a RIB dump) always means static.
	Protocol    string `json:"protocol,omitempty"`
	Description string `json:"description,omitempty"`
}

// Routing protocol identifiers. These are the values RoutingProtocol.Type takes,
// normalised across vendors so that a VyOS `protocols ospf`, a RouterOS
// `/routing/ospf/instance` and an FRR `router ospf` all report RoutingProtoOSPF.
const (
	RoutingProtoStatic = "static"
	RoutingProtoOSPF   = "ospf"
	RoutingProtoOSPFv3 = "ospfv3"
	RoutingProtoBGP    = "bgp"
	RoutingProtoRIP    = "rip"
	RoutingProtoISIS   = "isis"
)

// ConfigOSPFArea is one OSPF area of a RoutingProtocol instance.
type ConfigOSPFArea struct {
	// ID as the device states it — "0" and "0.0.0.0" are both seen in the wild
	// and are deliberately not normalised, because the device's own spelling is
	// what an operator will search for.
	ID string `json:"id"`
	// Type is "default"/"standard", "stub" or "nssa"; empty when unstated.
	Type string `json:"type,omitempty"`
	// Networks participating in this area, in CIDR.
	Networks []string `json:"networks,omitempty"`
	// Interfaces explicitly bound to this area, where the vendor binds by
	// interface rather than by network (Infix, and RouterOS templates that name
	// an interface instead of a prefix).
	Interfaces []string `json:"interfaces,omitempty"`
}

// ConfigBGPNeighbor is one configured BGP peer.
type ConfigBGPNeighbor struct {
	Address      string `json:"address"`                 // Peer address (often a loopback)
	RemoteAS     string `json:"remote_as"`               // Peer AS
	LocalAS      string `json:"local_as,omitempty"`      // Local AS, when stated per-neighbor
	Description  string `json:"description,omitempty"`   // Peer name/description
	UpdateSource string `json:"update_source,omitempty"` // Source interface/address for the session
	NextHopSelf  bool   `json:"next_hop_self,omitempty"`
}

// IsIBGP reports whether the session is internal — the peer is in our own AS.
// A neighbor whose remote AS is unknown, or whose local AS was never stated,
// cannot be classified and reports false.
func (n ConfigBGPNeighbor) IsIBGP(instanceAS string) bool {
	local := n.LocalAS
	if local == "" {
		local = instanceAS
	}
	return local != "" && n.RemoteAS != "" && local == n.RemoteAS
}

// ConfigRoutingProtocol is one routing protocol instance configured on a device.
//
// This is what distinguishes a routed topology from the flat, transparently
// bridged networks the tool was first built against: on a ring of five vendors
// the interfaces and addresses can be identical between the static, OSPF and
// iBGP variants, and the control plane is the only thing that differs. Recording
// which protocol carries the topology — and its router-id, areas and peers — is
// therefore the part that actually identifies the design.
type ConfigRoutingProtocol struct {
	Type    string `json:"type"`    // RoutingProto* constant
	Enabled bool   `json:"enabled"` // false for a configured-but-shutdown instance
	// Instance names the process where the vendor allows several ("default",
	// "os", an FRR VRF). Empty when the vendor has only one.
	Instance string `json:"instance,omitempty"`
	RouterID string `json:"router_id,omitempty"`
	// LocalAS is the BGP autonomous system this instance speaks for.
	LocalAS string `json:"local_as,omitempty"`
	VRF     string `json:"vrf,omitempty"`
	// Areas is OSPF-only; Neighbors is BGP-only.
	Areas     []ConfigOSPFArea    `json:"areas,omitempty"`
	Neighbors []ConfigBGPNeighbor `json:"neighbors,omitempty"`
	// Networks the instance originates or participates in, for protocols that
	// state them outside an area (BGP `network` statements, RIP).
	Networks []string `json:"networks,omitempty"`
	// Interfaces the protocol is enabled on, where stated directly.
	Interfaces []string `json:"interfaces,omitempty"`
	// Redistribute lists the protocols redistributed into this one
	// ("connected", "static", "ospf", …).
	Redistribute []string `json:"redistribute,omitempty"`
}

// HasIBGP reports whether any configured peer is an internal one.
func (p ConfigRoutingProtocol) HasIBGP() bool {
	for _, n := range p.Neighbors {
		if n.IsIBGP(p.LocalAS) {
			return true
		}
	}
	return false
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
	Protocol    string      `json:"protocol"`    // "tcp", "udp", "icmp", etc.
}

// ConfigAlias is a firewall alias
type ConfigAlias struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`         // "host", "network", "port", "url", "urltable"
	Content     []string `json:"content"`    // OPNsense table values; OpenWrt list items
	Enabled     bool     `json:"enabled"`
	Description string   `json:"description,omitempty"`
}

// ConfigNAT is a NAT rule: 1:1, source, destination, or NPTv6.
type ConfigNAT struct {
	Type        string `json:"type"`              // "1:1", "source", "dest", "nptv6"
	PublicIP    string `json:"public_ip,omitempty"`
	PrivateIP   string `json:"private_ip,omitempty"`
	PublicPort  string `json:"public_port,omitempty"`
	PrivatePort string `json:"private_port,omitempty"`
	Interface   string `json:"interface,omitempty"`
	Description string `json:"description,omitempty"`
}

// ConfigDNSForwarder is dnsmasq / DNS forwarder settings.
type ConfigDNSForwarder struct {
	Enabled  bool     `json:"enabled"`
	Listen   []string `json:"listen"`
	Upstream []string `json:"upstream"`
	Domains  []string `json:"domains,omitempty"`
}

// ConfigDNSResolver is Unbound DNS resolver configuration.
type ConfigDNSResolver struct {
	Enabled  bool     `json:"enabled"`
	DNSSEC   bool     `json:"dnssec"`
	Listen   []string `json:"listen"`
	Upstream []string `json:"upstream,omitempty"`
	RPZZones []string `json:"rpz_zones,omitempty"`
}

// ConfigDHCPScope is one DHCP scope (range) on an interface.
type ConfigDHCPScope struct {
	Interface  string   `json:"interface"`
	RangeStart string   `json:"range_start"`
	RangeEnd   string   `json:"range_end"`
	Gateway    string   `json:"gateway,omitempty"`
	DNS        []string `json:"dns,omitempty"`
	Domain     string   `json:"domain,omitempty"`
	LeaseTime  string   `json:"lease_time,omitempty"`
	Enabled    bool     `json:"enabled"`
}

// ConfigDHCPREServation is a DHCP static reservation (MAC → IP).
type ConfigDHCPREServation struct {
	MACAddress string `json:"mac"`
	IPAddress  string `json:"ip"`
	Hostname   string `json:"hostname,omitempty"`
	Description string `json:"description,omitempty"`
}

// ConfigVIP is a Virtual IP — OPNsense "Virtual IPs", VyOS "virtual-address".
type ConfigVIP struct {
	Address    string `json:"address"`
	Mode       string `json:"mode"`        // "carp", "ipalias", "proxyarp", "single"
	Interface  string `json:"interface"`
	VHID       uint8  `json:"vhid,omitempty"`
	Password   string `json:"password,omitempty"`
	AdvBase    int    `json:"adv_base,omitempty"`
	AdvSkew    int    `json:"adv_skew,omitempty"`
	Description string `json:"description,omitempty"`
}

// PortVLANInfo represents VLAN membership for a switch port
type PortVLANInfo struct {
	VID    string `json:"vid"`    // VLAN ID (e.g., "1", "60")
	Tagged bool   `json:"tagged"` // true if port is tagged on this VLAN
}

// SwitchPortInfo represents physical switch port information from board.json/swconfig
type SwitchPortInfo struct {
	PortNumber int            `json:"port_number"`      // Port index (0, 1, 2, etc.)
	PortName   string         `json:"port_name"`        // Canonical name — the kernel netdev ("eth1") when known, else a role+index label ("lan1")
	Device     string         `json:"device,omitempty"` // Kernel netdev name ("eth1") when the board exposes one
	LinkStatus string         `json:"link_status"`      // "up", "down"
	Role       string         `json:"role"`             // "lan", "wan"
	VLANs      []PortVLANInfo `json:"vlans"`            // VLAN membership per port (from swconfig)
}

// ConfigData contains the parsed configuration data from a network device
type ConfigData struct {
	OsType        string            `json:"os_type"`      // "opnsense", "openwrt", "fortinet", "freebsd"
	DeviceModel   string            `json:"device_model"` // Specific model/version info
	Hostname      string            `json:"hostname"`
	Domain        string            `json:"domain,omitempty"`
	ConfigVersion string            `json:"config_version"` // Configuration version/timestamp
	Source        ConfigSource      `json:"source"`         // How config was obtained
	Interfaces    []ConfigInterface `json:"interfaces"`
	VLANs         []ConfigVLAN      `json:"vlans"`
	Routes        []ConfigRoute     `json:"routes"`
	// RoutingProtocols records the control plane: which dynamic routing
	// protocols are configured, and how. Empty means the device routes
	// statically (or not at all) — see ControlPlane.
	RoutingProtocols []ConfigRoutingProtocol `json:"routing_protocols,omitempty"`
	FirewallRules    []ConfigFirewallRule     `json:"firewall_rules"`
	Aliases         []ConfigAlias           `json:"aliases,omitempty"`
	NATs            []ConfigNAT             `json:"nats,omitempty"`
	DNSForwarder    *ConfigDNSForwarder     `json:"dns_forwarder,omitempty"`
	DNSResolver     *ConfigDNSResolver      `json:"dns_resolver,omitempty"`
	DHCPScopes      []ConfigDHCPScope       `json:"dhcp_scopes,omitempty"`
	DHCPReservations []ConfigDHCPREServation `json:"dhcp_reservations,omitempty"`
	VIPs            []ConfigVIP             `json:"vips,omitempty"`
	NTP             *ConfigNTPConfig        `json:"ntp,omitempty"`
	Banner          *ConfigBanner          `json:"banner,omitempty"`
	LLDP           *ConfigLLDPSettings  `json:"lldp,omitempty"`
	Syslog          *ConfigSyslogConfig     `json:"syslog,omitempty"`
	SwitchPorts    []SwitchPortInfo        `json:"switch_ports,omitempty"`
	Raw            string                  `json:"raw,omitempty"`
	ParsedAt       time.Time               `json:"parsed_at"`
}

// ControlPlane summarises how this device learns routes, as a stable, sorted
// list of protocol identifiers.
//
// A device with no dynamic protocol configured but with static routes reports
// ["static"]; one with neither reports nil. A device running an OSPF underlay
// beneath iBGP — the shape of the lab's BGP ring — reports ["bgp", "ospf"],
// because both are genuinely part of its control plane and collapsing that to a
// single "primary" protocol would lose the underlay.
func (c *ConfigData) ControlPlane() []string {
	seen := make(map[string]bool, len(c.RoutingProtocols))
	for _, p := range c.RoutingProtocols {
		if p.Enabled && p.Type != "" {
			seen[p.Type] = true
		}
	}
	if len(seen) == 0 {
		if len(c.Routes) == 0 {
			return nil
		}
		return []string{RoutingProtoStatic}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// RoutingProtocolsOfType returns the configured instances of one protocol.
func (c *ConfigData) RoutingProtocolsOfType(t string) []ConfigRoutingProtocol {
	var out []ConfigRoutingProtocol
	for _, p := range c.RoutingProtocols {
		if p.Type == t {
			out = append(out, p)
		}
	}
	return out
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

// ConfigParser defines the interface that all device-specific configuration parsers must
// implement.
//
// It knows what to ASK a device and how to read the answer. It knows nothing about how
// the connection is made -- that is the Transport's job (transport.go). The two used to
// be fused: every parser opened its own SSH connection, so there were five copies of the
// same client code and not one of the fetch strategies could be tested without a device.
type ConfigParser interface {
	// GetOsType returns the OS type this parser handles (e.g., "opnsense", "fortinet")
	GetOsType() string

	// SupportsDevice returns true if this parser can handle the given device
	SupportsDevice(device s.SNMPDevice) bool

	// Fetch assembles the device's raw configuration over an already-open session.
	//
	// This is where each OS's real knowledge lives: which commands to run, in what
	// order, and what to fall back on when one is unavailable. OPNsense reads
	// /conf/config.xml and falls back to configctl; FreeBSD tries a CIDR-formatted
	// ifconfig and falls back to the plain one; OpenWrt runs seven uci commands and
	// tags each block so ParseConfig can re-split them.
	Fetch(sess Session) (string, error)

	// ParseConfig parses raw configuration data and returns structured ConfigData
	ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*ConfigData, error)

	// ValidateConfig performs basic validation on parsed configuration
	ValidateConfig(config *ConfigData) []error
}

// ConfigParserOptions contains options for configuration parsing
type ConfigParserOptions struct {
	Source            ConfigSource      `json:"source"`
	FilePath          string            `json:"file_path,omitempty"`
	OsType            string            `json:"os_type,omitempty"` // Manual OS override (opnsense, openwrt, fortinet)
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
	r.parsers[parser.GetOsType()] = parser
}

// osTypeAliases maps the name an operator is likely to type to the canonical OS
// type a parser registers under.
//
// The canonical name is the operating system, because that is what determines
// the grammar — but people reach for the vendor. Someone with a MikroTik in front
// of them types "mikrotik", not "routeros", and being told that is an unknown
// device type is a pointless obstacle. Aliases are resolved on lookup only; the
// registry itself still holds exactly one entry per parser, and ListParsers
// keeps reporting canonical names so error messages stay unambiguous.
var osTypeAliases = map[string]string{
	"mikrotik":  "routeros",
	"ros":       "routeros",
	"pfsense":   "opnsense", // same FreeBSD/ifconfig grammar
	"freebsd":   "opnsense",
	"vyatta":    "vyos",
	"fortios":   "fortinet",
	"fortigate": "fortinet",
}

// resolveOsType maps an alias to its canonical OS type, leaving unknown names
// untouched so the caller still reports them as unsupported.
func resolveOsType(osType string) string {
	key := strings.ToLower(strings.TrimSpace(osType))
	if canonical, ok := osTypeAliases[key]; ok {
		return canonical
	}
	return key
}

// GetParser returns the appropriate parser for a OS type, accepting the common
// vendor-name aliases as well as the canonical one.
func (r *ConfigParserRegistry) GetParser(osType string) (ConfigParser, bool) {
	parser, exists := r.parsers[resolveOsType(osType)]
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
func (r *ConfigParserRegistry) GetParserForDeviceWithType(device s.SNMPDevice, osType string) (ConfigParser, bool) {
	// Use manual type if specified and valid
	if osType != "" {
		if parser, exists := r.parsers[resolveOsType(osType)]; exists {
			return parser, true
		}
	}

	// Fallback to auto-detection
	return r.GetParserForDevice(device)
}

// ListParsers returns all registered parser OS types
func (r *ConfigParserRegistry) ListParsers() []string {
	var types []string
	for osType := range r.parsers {
		types = append(types, osType)
	}
	return types
}

// DefaultRegistry is the global parser registry. Parsers self-register via init().
var DefaultRegistry = NewConfigParserRegistry()
