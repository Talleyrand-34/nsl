package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"nsl-graph/internal/configparser"
	s "nsl-graph/internal/scanner"
)

// CiscoParser handles parsing of Cisco IOS configuration files
// NOTE: This is a placeholder implementation for future development
type CiscoParser struct{}

// NewCiscoParser creates a new Cisco configuration parser
func NewCiscoParser() *CiscoParser {
	return &CiscoParser{}
}

func init() { configparser.DefaultRegistry.RegisterParser(NewCiscoParser()) }

// GetDeviceType returns the device type this parser handles
func (p *CiscoParser) GetDeviceType() string {
	return "cisco"
}

// SupportsDevice returns true if this parser can handle the given device
func (p *CiscoParser) SupportsDevice(device s.SNMPDevice) bool {
	descr := strings.ToLower(device.SysDescr)
	return strings.Contains(descr, "cisco") || strings.Contains(descr, "ios")
}

// ParseConfig parses raw Cisco IOS configuration and returns structured ConfigData
func (p *CiscoParser) ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*configparser.ConfigData, error) {
	// TODO: Implement full Cisco IOS configuration parsing
	// This is a placeholder implementation that provides basic parsing

	configData := &configparser.ConfigData{
		DeviceType:    p.GetDeviceType(),
		DeviceModel:   extractCiscoModel(rawConfig, deviceInfo),
		Hostname:      extractCiscoHostname(rawConfig),
		ConfigVersion: extractCiscoVersion(rawConfig),
		Source:        configparser.ConfigSourceSSH, // Will be set by caller
		ParsedAt:      time.Now(),
		Raw:           rawConfig,
	}

	// Basic interface parsing
	interfaces, err := p.parseBasicInterfaces(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interfaces: %w", err)
	}
	configData.Interfaces = interfaces

	// Basic VLAN parsing
	vlans, err := p.parseBasicVLANs(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VLANs: %w", err)
	}
	configData.VLANs = vlans

	// TODO: Implement complete parsing for:
	// - Static routes
	// - Access control lists (ACLs)
	// - VRF configurations
	// - BGP/OSPF configurations
	// - QoS policies

	return configData, nil
}

// GetConfigViaSSH retrieves Cisco configuration via SSH
func (p *CiscoParser) GetConfigViaSSH(ip string, creds configparser.SSHCredentials) (string, error) {
	client := configparser.NewSSHClient(creds)
	if err := client.Connect(ip); err != nil {
		return "", fmt.Errorf("failed to connect to Cisco device: %w", err)
	}
	defer client.Close()

	// Cisco IOS commands to show configuration
	// Note: Some devices might require enable mode
	output, err := client.Execute("show running-config")
	if err != nil {
		// Try alternative command
		output, err = client.Execute("show config")
		if err != nil {
			return "", fmt.Errorf("failed to retrieve Cisco configuration: %w", err)
		}
	}

	return output, nil
}

// ValidateConfig performs basic validation on parsed Cisco configuration
func (p *CiscoParser) ValidateConfig(config *configparser.ConfigData) []error {
	var errors []error

	if config.Hostname == "" {
		errors = append(errors, fmt.Errorf("hostname is required"))
	}

	// Basic interface validation
	interfaceNames := make(map[string]bool)
	for _, iface := range config.Interfaces {
		if interfaceNames[iface.Name] {
			errors = append(errors, fmt.Errorf("duplicate interface name: %s", iface.Name))
		}
		interfaceNames[iface.Name] = true
	}

	// Basic VLAN validation
	vlanIDs := make(map[string]bool)
	for _, vlan := range config.VLANs {
		if vlanIDs[vlan.ID] {
			errors = append(errors, fmt.Errorf("duplicate VLAN ID: %s", vlan.ID))
		}
		vlanIDs[vlan.ID] = true

		if vlanID, err := strconv.Atoi(vlan.ID); err == nil {
			if vlanID < 1 || vlanID > 4094 {
				errors = append(errors, fmt.Errorf("VLAN ID %s is out of valid range (1-4094)", vlan.ID))
			}
		}
	}

	return errors
}

// parseBasicInterfaces provides basic interface parsing for Cisco IOS
// TODO: Extend this to handle all Cisco interface types and configurations
func (p *CiscoParser) parseBasicInterfaces(rawConfig string) ([]configparser.ConfigInterface, error) {
	var interfaces []configparser.ConfigInterface

	// Split config into sections
	sections := strings.Split(rawConfig, "\n!\n")

	interfaceRegex := regexp.MustCompile(`^interface\s+(.+)`)
	ipRegex := regexp.MustCompile(`ip address\s+(\S+)\s+(\S+)`)
	descriptionRegex := regexp.MustCompile(`description\s+(.+)`)
	shutdownRegex := regexp.MustCompile(`shutdown`)
	encapsulationRegex := regexp.MustCompile(`encapsulation dot1Q\s+(\d+)`)

	for _, section := range sections {
		lines := strings.Split(strings.TrimSpace(section), "\n")
		if len(lines) == 0 {
			continue
		}

		// Check if this is an interface section
		interfaceMatch := interfaceRegex.FindStringSubmatch(lines[0])
		if interfaceMatch == nil {
			continue
		}

		interfaceName := interfaceMatch[1]
		configIface := configparser.ConfigInterface{
			Name:    interfaceName,
			Enabled: true, // Default to enabled
			Type:    determineCiscoInterfaceType(interfaceName),
		}

		// Parse interface configuration lines
		for _, line := range lines[1:] {
			line = strings.TrimSpace(line)

			// Parse IP address
			if ipMatch := ipRegex.FindStringSubmatch(line); ipMatch != nil {
				ip, netmask := ipMatch[1], ipMatch[2]
				// Convert netmask to CIDR notation (simplified)
				cidr := convertNetmaskToCIDR(netmask)
				configIface.IPAddresses = append(configIface.IPAddresses, fmt.Sprintf("%s/%s", ip, cidr))
			}

			// Parse description
			if descMatch := descriptionRegex.FindStringSubmatch(line); descMatch != nil {
				configIface.Description = descMatch[1]
			}

			// Check if interface is shutdown
			if shutdownRegex.MatchString(line) {
				configIface.Enabled = false
			}

			// Parse VLAN encapsulation
			if encapMatch := encapsulationRegex.FindStringSubmatch(line); encapMatch != nil {
				vlan := configparser.ConfigVLAN{
					ID:     encapMatch[1],
					Tagged: true,
				}
				configIface.VLANs = append(configIface.VLANs, vlan)
				configIface.Type = "vlan"

				// Extract parent interface for subinterfaces
				if strings.Contains(interfaceName, ".") {
					parts := strings.Split(interfaceName, ".")
					if len(parts) > 1 {
						configIface.Parent = parts[0]
					}
				}
			}
		}

		interfaces = append(interfaces, configIface)
	}

	return interfaces, nil
}

// parseBasicVLANs provides basic VLAN parsing for Cisco IOS
// TODO: Extend this to handle VLAN database, VTP, and other VLAN features
func (p *CiscoParser) parseBasicVLANs(rawConfig string) ([]configparser.ConfigVLAN, error) {
	var vlans []configparser.ConfigVLAN

	// Look for VLAN definitions
	vlanRegex := regexp.MustCompile(`(?m)^vlan\s+(\d+)$`)
	nameRegex := regexp.MustCompile(`(?m)^\s+name\s+(.+)$`)

	lines := strings.Split(rawConfig, "\n")
	var currentVLAN *configparser.ConfigVLAN

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Start of VLAN definition
		if vlanMatch := vlanRegex.FindStringSubmatch(line); vlanMatch != nil {
			// Save previous VLAN if exists
			if currentVLAN != nil {
				vlans = append(vlans, *currentVLAN)
			}

			// Start new VLAN
			currentVLAN = &configparser.ConfigVLAN{
				ID:      vlanMatch[1],
				Name:    fmt.Sprintf("VLAN%s", vlanMatch[1]),
				Enabled: true,
				Tagged:  true,
			}
			continue
		}

		// VLAN name
		if currentVLAN != nil {
			if nameMatch := nameRegex.FindStringSubmatch(line); nameMatch != nil {
				currentVLAN.Name = nameMatch[1]
				currentVLAN.Description = nameMatch[1]
			}
		}

		// End of VLAN section (when we hit a non-VLAN line that doesn't start with space)
		if currentVLAN != nil && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "vlan ") && line != "" {
			vlans = append(vlans, *currentVLAN)
			currentVLAN = nil
		}
	}

	// Add last VLAN if exists
	if currentVLAN != nil {
		vlans = append(vlans, *currentVLAN)
	}

	return vlans, nil
}

// Helper functions for Cisco configuration parsing

func determineCiscoInterfaceType(interfaceName string) string {
	name := strings.ToLower(interfaceName)

	switch {
	case strings.Contains(name, "loopback"):
		return "loopback"
	case strings.Contains(name, "vlan"):
		return "vlan"
	case strings.Contains(name, "tunnel"):
		return "tunnel"
	case strings.Contains(name, "."):
		return "vlan" // Subinterface
	case strings.HasPrefix(name, "gi") || strings.HasPrefix(name, "gigabitethernet"):
		return "physical"
	case strings.HasPrefix(name, "fa") || strings.HasPrefix(name, "fastethernet"):
		return "physical"
	case strings.HasPrefix(name, "te") || strings.HasPrefix(name, "tengigabitethernet"):
		return "physical"
	case strings.HasPrefix(name, "eth") || strings.HasPrefix(name, "ethernet"):
		return "physical"
	case strings.HasPrefix(name, "ser") || strings.HasPrefix(name, "serial"):
		return "serial"
	default:
		return "unknown"
	}
}

func convertNetmaskToCIDR(netmask string) string {
	// Simple conversion for common netmasks
	// TODO: Implement full netmask to CIDR conversion
	switch netmask {
	case "255.255.255.255":
		return "32"
	case "255.255.255.254":
		return "31"
	case "255.255.255.252":
		return "30"
	case "255.255.255.248":
		return "29"
	case "255.255.255.240":
		return "28"
	case "255.255.255.224":
		return "27"
	case "255.255.255.192":
		return "26"
	case "255.255.255.128":
		return "25"
	case "255.255.255.0":
		return "24"
	case "255.255.254.0":
		return "23"
	case "255.255.252.0":
		return "22"
	case "255.255.248.0":
		return "21"
	case "255.255.240.0":
		return "20"
	case "255.255.224.0":
		return "19"
	case "255.255.192.0":
		return "18"
	case "255.255.128.0":
		return "17"
	case "255.255.0.0":
		return "16"
	case "255.254.0.0":
		return "15"
	case "255.252.0.0":
		return "14"
	case "255.248.0.0":
		return "13"
	case "255.240.0.0":
		return "12"
	case "255.224.0.0":
		return "11"
	case "255.192.0.0":
		return "10"
	case "255.128.0.0":
		return "9"
	case "255.0.0.0":
		return "8"
	default:
		return "24" // Default fallback
	}
}

func extractCiscoHostname(rawConfig string) string {
	re := regexp.MustCompile(`(?m)^hostname\s+(.+)$`)
	matches := re.FindStringSubmatch(rawConfig)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func extractCiscoVersion(rawConfig string) string {
	// Look for version information in various places
	patterns := []string{
		`(?m)^!\s*Last configuration change at (.+) by`,
		`(?m)^version\s+(.+)$`,
		`(?m)^!\s*NVGEN-\d+\s+(.+)$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(rawConfig)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return time.Now().Format("2006-01-02")
}

func extractCiscoModel(rawConfig string, deviceInfo s.SNMPDevice) string {
	// Try to extract model from show version output if present
	re := regexp.MustCompile(`(?m)^Cisco\s+([^\s,]+)`)
	matches := re.FindStringSubmatch(rawConfig)
	if len(matches) > 1 {
		return fmt.Sprintf("Cisco %s", matches[1])
	}

	// Fallback to SNMP description
	if strings.Contains(deviceInfo.SysDescr, "Cisco") {
		return deviceInfo.SysDescr
	}

	return "Cisco Device"
}

// TODO: Future enhancements for complete Cisco IOS support:
//
// 1. Interface Configuration:
//    - Switchport modes (access, trunk)
//    - Port security
//    - EtherChannel configuration
//    - Voice VLANs
//    - Port mirroring (SPAN)
//
// 2. VLAN Configuration:
//    - VLAN Trunking Protocol (VTP)
//    - VLAN database mode vs config mode
//    - Private VLANs
//    - Voice VLANs
//
// 3. Routing Configuration:
//    - Static routes
//    - Dynamic routing protocols (OSPF, EIGRP, BGP, RIP)
//    - Route maps and access lists
//    - VRF configuration
//    - Policy-based routing
//
// 4. Security Configuration:
//    - Access Control Lists (ACLs)
//    - Authentication (AAA)
//    - Port security
//    - 802.1X configuration
//
// 5. Quality of Service (QoS):
//    - Class maps and policy maps
//    - Traffic shaping and policing
//    - DSCP marking
//
// 6. High Availability:
//    - HSRP/VRRP configuration
//    - Stack configuration
//    - VSS/VPC configuration
//
// 7. Management Configuration:
//    - SNMP configuration
//    - Logging configuration
//    - NTP configuration
//    - SSH/Telnet configuration