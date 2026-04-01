package parsers

import (
	"fmt"
	"strings"
	"time"

	"nsl-graph/internal/configparser"
	s "nsl-graph/internal/scanner"
)

// FreeBSDParser retrieves and parses OPNsense (and generic FreeBSD) device state
// using standard shell commands (hostname, ifconfig) rather than XML config files.
// It registers as "opnsense" and is the primary auto-detected parser for FreeBSD devices.
type FreeBSDParser struct{}

// NewFreeBSDParser creates a new FreeBSD/ifconfig-based parser.
func NewFreeBSDParser() *FreeBSDParser {
	return &FreeBSDParser{}
}

func init() { configparser.DefaultRegistry.RegisterParser(NewFreeBSDParser()) }

// GetDeviceType returns the device type key used for --device-type and registry lookup.
func (p *FreeBSDParser) GetDeviceType() string {
	return "opnsense"
}

// SupportsDevice returns true for FreeBSD-based devices (OPNsense, pfSense, generic FreeBSD).
func (p *FreeBSDParser) SupportsDevice(device s.SNMPDevice) bool {
	descr := strings.ToLower(device.SysDescr)
	return strings.Contains(descr, "freebsd") || strings.Contains(descr, "opnsense") ||
		strings.Contains(device.SysName, "opnsense") || strings.Contains(device.SysName, "OPNsense")
}

// GetConfigViaSSH retrieves device state by running hostname and ifconfig over SSH.
// The returned string has the form:
//
//	HOSTNAME:<value>\n<ifconfig -a output>
//
// Uses -f inet:cidr,inet6:cidr for CIDR notation (FreeBSD 12+); falls back to
// plain ifconfig -a for older releases.
func (p *FreeBSDParser) GetConfigViaSSH(ip string, creds configparser.SSHCredentials) (string, error) {
	client := configparser.NewSSHClient(creds)
	if err := client.Connect(ip); err != nil {
		return "", fmt.Errorf("failed to connect to FreeBSD device: %w", err)
	}
	defer client.Close()

	hostname, err := client.Execute("hostname")
	if err != nil {
		hostname = ""
	}
	hostname = strings.TrimSpace(hostname)

	// Try CIDR-format ifconfig first (FreeBSD 12+, all current OPNsense releases).
	ifcfgOut, err := client.Execute("ifconfig -a -f inet:cidr,inet6:cidr")
	if err != nil || strings.TrimSpace(ifcfgOut) == "" {
		ifcfgOut, err = client.Execute("ifconfig -a")
		if err != nil {
			return "", fmt.Errorf("failed to execute ifconfig: %w", err)
		}
	}

	return fmt.Sprintf("HOSTNAME:%s\n%s", hostname, ifcfgOut), nil
}

// ParseConfig parses the combined HOSTNAME+ifconfig output produced by GetConfigViaSSH,
// or a raw ifconfig dump (e.g. from --config-source file) without the HOSTNAME prefix.
func (p *FreeBSDParser) ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*configparser.ConfigData, error) {
	hostname, ifcfgBody := extractFreeBSDHostname(rawConfig)
	if hostname == "" {
		hostname = deviceInfo.SysName
	}

	interfaces, vlans, err := parseIfconfigOutput(ifcfgBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ifconfig output: %w", err)
	}

	return &configparser.ConfigData{
		DeviceType:  p.GetDeviceType(),
		DeviceModel: "OPNsense Firewall",
		Hostname:    hostname,
		Source:      configparser.ConfigSourceSSH,
		Interfaces:  interfaces,
		VLANs:       vlans,
		ParsedAt:    time.Now(),
		Raw:         rawConfig,
	}, nil
}

// ValidateConfig performs basic structural validation on parsed FreeBSD configuration.
func (p *FreeBSDParser) ValidateConfig(config *configparser.ConfigData) []error {
	var errs []error
	if config.Hostname == "" {
		errs = append(errs, fmt.Errorf("hostname is empty"))
	}
	seen := make(map[string]bool)
	for _, iface := range config.Interfaces {
		if iface.Name == "" {
			errs = append(errs, fmt.Errorf("interface with empty name"))
			continue
		}
		if seen[iface.Name] {
			errs = append(errs, fmt.Errorf("duplicate interface name: %s", iface.Name))
		}
		seen[iface.Name] = true
	}
	return errs
}

// ---------------------------------------------------------------------------
// Internal parsing helpers
// ---------------------------------------------------------------------------

// extractFreeBSDHostname splits the raw combined output into hostname and ifconfig body.
// The first line must start with "HOSTNAME:" when produced by GetConfigViaSSH.
// If no such prefix is found, the entire string is treated as ifconfig output.
func extractFreeBSDHostname(raw string) (hostname, body string) {
	const prefix = "HOSTNAME:"
	firstNewline := strings.IndexByte(raw, '\n')
	if firstNewline < 0 {
		return "", raw
	}
	firstLine := raw[:firstNewline]
	if strings.HasPrefix(firstLine, prefix) {
		return strings.TrimSpace(firstLine[len(prefix):]), raw[firstNewline+1:]
	}
	return "", raw
}

// parseIfconfigOutput parses BSD ifconfig output into ConfigInterface and ConfigVLAN slices.
//
// Interface blocks begin at column 0: "NAME: flags=...".
// Property lines are indented with a tab or spaces.
func parseIfconfigOutput(output string) ([]configparser.ConfigInterface, []configparser.ConfigVLAN, error) {
	var interfaces []configparser.ConfigInterface
	vlansByID := make(map[string]configparser.ConfigVLAN) // deduplicate top-level VLANs

	lines := strings.Split(output, "\n")

	var currentName string
	var currentLines []string

	flush := func() {
		if currentName == "" {
			return
		}
		iface, vlan := parseInterfaceBlock(currentName, currentLines)
		interfaces = append(interfaces, iface)
		if vlan != nil {
			if _, exists := vlansByID[vlan.ID]; !exists {
				vlansByID[vlan.ID] = *vlan
			}
		}
		currentName = ""
		currentLines = nil
	}

	for _, line := range lines {
		if line == "" {
			continue
		}
		// A new interface block starts when the first character is NOT whitespace.
		if line[0] != ' ' && line[0] != '\t' {
			flush()
			colonIdx := strings.IndexByte(line, ':')
			if colonIdx > 0 {
				currentName = line[:colonIdx]
				currentLines = []string{line}
			}
		} else if currentName != "" {
			currentLines = append(currentLines, line)
		}
	}
	flush()

	var vlans []configparser.ConfigVLAN
	for _, v := range vlansByID {
		vlans = append(vlans, v)
	}

	return interfaces, vlans, nil
}

// parseInterfaceBlock processes the lines belonging to a single interface block.
// Returns the ConfigInterface and an optional top-level ConfigVLAN when the
// interface is a VLAN subinterface.
func parseInterfaceBlock(name string, lines []string) (configparser.ConfigInterface, *configparser.ConfigVLAN) {
	iface := configparser.ConfigInterface{Name: name}

	if len(lines) > 0 {
		iface.Enabled = isIfaceUp(lines[0])
	}

	var vlanID, vlanParent string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "ether "):
			// MAC address: "ether XX:XX:XX:XX:XX:XX"
			iface.MACAddress = strings.TrimPrefix(trimmed, "ether ")

		case strings.HasPrefix(trimmed, "inet6 "):
			// IPv6: skip link-local scoped addresses (contain '%')
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 && !strings.Contains(fields[1], "%") {
				iface.IPAddresses = append(iface.IPAddresses, fields[1])
			}

		case strings.HasPrefix(trimmed, "inet "):
			// IPv4: "inet IP/PREFIX ..." — first field after "inet" is the address
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 && !strings.Contains(fields[1], "%") {
				iface.IPAddresses = append(iface.IPAddresses, fields[1])
			}

		case strings.HasPrefix(trimmed, "vlan: "):
			// VLAN membership: "vlan: N vlanpcp: P parent interface: IFACE"
			vlanID, vlanParent = parseVLANLine(trimmed)
		}
	}

	iface.Type = classifyFreeBSDIface(name, lines, vlanID != "")

	if vlanID != "" {
		iface.Parent = vlanParent
		iface.VLANs = []configparser.ConfigVLAN{
			{ID: vlanID, Tagged: true, Enabled: true},
		}
		return iface, &configparser.ConfigVLAN{
			ID:      vlanID,
			Name:    fmt.Sprintf("VLAN_%s", vlanID),
			Enabled: true,
			Tagged:  true,
		}
	}

	return iface, nil
}

// isIfaceUp returns true when the flags field on the header line contains the token "UP".
// Example: "igc0: flags=8843<UP,BROADCAST,RUNNING,SIMPLEX,MULTICAST> metric 0 mtu 1500"
func isIfaceUp(headerLine string) bool {
	lt := strings.IndexByte(headerLine, '<')
	gt := strings.IndexByte(headerLine, '>')
	if lt < 0 || gt <= lt {
		return false
	}
	for _, f := range strings.Split(headerLine[lt+1:gt], ",") {
		if f == "UP" {
			return true
		}
	}
	return false
}

// parseVLANLine extracts the VLAN ID and parent interface name from a vlan line.
// Input: "vlan: 10 vlanpcp: 0 parent interface: igc0"
func parseVLANLine(line string) (id, parent string) {
	fields := strings.Fields(line)
	for i, f := range fields {
		switch f {
		case "vlan:":
			if i+1 < len(fields) {
				id = fields[i+1]
			}
		case "interface:":
			if i+1 < len(fields) {
				parent = fields[i+1]
			}
		}
	}
	return id, parent
}

// classifyFreeBSDIface determines the interface type string.
func classifyFreeBSDIface(name string, lines []string, isVLAN bool) string {
	if isVLAN {
		return "vlan"
	}
	if strings.HasPrefix(name, "lo") {
		return "loopback"
	}
	if len(lines) > 0 && strings.Contains(lines[0], "POINTOPOINT") {
		return "tunnel"
	}

	// FreeBSD physical interface naming patterns
	// igc*, igb*, em*, ix*, bge*, re*, vtnet*, vmx*, hn*, ue*
	if isPhysicalInterfaceName(name) {
		return "physical"
	}

	return "logical"
}

// isPhysicalInterfaceName checks if name matches FreeBSD physical NIC patterns
func isPhysicalInterfaceName(name string) bool {
	physicalPrefixes := []string{
		"igc",   // Intel I225/I226 2.5G NICs
		"igb",   // Intel PRO/1000 gigabit
		"em",    // Intel PRO/100 (older)
		"ix",    // Intel 10GbE
		"bge",   // Broadcom BCM57xx
		"re",    // Realtek 8139/8169
		"vtnet", // VirtIO
		"vmx",   // VMware VMXNET3
		"hn",    // Hyper-V
		"ue",    // USB Ethernet adapters
	}
	for _, prefix := range physicalPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
