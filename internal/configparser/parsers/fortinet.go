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

// FortinetParser handles parsing of Fortinet FortiGate CLI configuration
type FortinetParser struct{}

// NewFortinetParser creates a new Fortinet configuration parser
func NewFortinetParser() *FortinetParser {
	return &FortinetParser{}
}

func init() { configparser.DefaultRegistry.RegisterParser(NewFortinetParser()) }

// GetOsType returns the OS type this parser handles
func (p *FortinetParser) GetOsType() string {
	return "fortinet"
}

// SupportsDevice returns true if this parser can handle the given device
func (p *FortinetParser) SupportsDevice(device s.SNMPDevice) bool {
	descr := strings.ToLower(device.SysDescr)
	return strings.Contains(descr, "fortinet") || strings.Contains(descr, "fortigate") ||
		strings.Contains(device.SysName, "FortiGate") || strings.Contains(device.SysName, "FG")
}

// cleanFortinetCLIOutput removes CLI prompt prefixes and pager markers from
// FortiGate SSH output. FortiGate prepends the prompt ("HOSTNAME # ") to the
// first line of each command's output, and may inject "--More--" pager markers
// in the middle of long output.
func cleanFortinetCLIOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		// Strip CLI prompt: anything up to and including " # "
		// e.g. "FGT30D3X15012871 # config system global" → "config system global"
		if idx := strings.Index(line, " # "); idx >= 0 {
			line = line[idx+3:]
		}
		// Handle --More-- pager: recover any content that follows the marker
		if idx := strings.Index(line, "--More--"); idx >= 0 {
			after := strings.TrimLeft(line[idx+len("--More--"):], " ")
			line = after // may be empty — filtered by TrimSpace in parseFortiConfig
		}
		cleaned = append(cleaned, line)
	}
	return strings.Join(cleaned, "\n")
}

// ParseConfig parses raw Fortinet CLI configuration and returns structured ConfigData
func (p *FortinetParser) ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*configparser.ConfigData, error) {
	rawConfig = cleanFortinetCLIOutput(rawConfig)
	fortiConfig, err := p.parseFortiConfig(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse FortiGate config: %w", err)
	}

	configData := &configparser.ConfigData{
		OsType:        p.GetOsType(),
		DeviceModel:   extractFortinetModel(fortiConfig, deviceInfo),
		Hostname:      extractFortinetHostname(fortiConfig),
		ConfigVersion: extractFortinetVersion(fortiConfig),
		Source:        configparser.ConfigSourceSSH, // Will be set by caller
		ParsedAt:      time.Now(),
		Raw:           rawConfig,
	}

	// Parse interfaces
	interfaces, err := p.parseFortinetInterfaces(fortiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interfaces: %w", err)
	}
	configData.Interfaces = interfaces

	// Parse VLANs
	vlans, err := p.parseFortinetVLANs(fortiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VLANs: %w", err)
	}
	configData.VLANs = vlans

	// Parse routes
	routes, err := p.parseFortinetRoutes(fortiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse routes: %w", err)
	}
	configData.Routes = routes

	// Parse firewall rules (policies)
	firewallRules, err := p.parseFortinetPolicies(fortiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse firewall policies: %w", err)
	}
	configData.FirewallRules = firewallRules

	return configData, nil
}

// GetConfigViaSSH retrieves Fortinet configuration via SSH
func (p *FortinetParser) GetConfigViaSSH(ip string, creds configparser.SSHCredentials) (string, error) {
	client := configparser.NewSSHClient(creds)
	if err := client.Connect(ip); err != nil {
		return "", fmt.Errorf("failed to connect to FortiGate device: %w", err)
	}
	defer client.Close()

	// Use targeted commands to avoid nested config blocks in show full-configuration.
	// show system interface output is clean (no nested config...end blocks).
	sysInterface, ifErr := client.Execute("show system interface")
	if ifErr != nil {
		// Fallback: full configuration (handled by depth-tracking parser)
		output, err := client.Execute("show full-configuration")
		if err != nil {
			return "", fmt.Errorf("failed to retrieve FortiGate configuration: %w", ifErr)
		}
		return output, nil
	}

	// Also fetch system global for hostname extraction
	var parts []string
	if sysGlobal, err := client.Execute("show system global"); err == nil {
		parts = append(parts, sysGlobal)
	}
	parts = append(parts, sysInterface)
	return strings.Join(parts, "\n"), nil
}

// ValidateConfig performs basic validation on parsed Fortinet configuration
func (p *FortinetParser) ValidateConfig(config *configparser.ConfigData) []error {
	var errors []error

	if config.Hostname == "" {
		errors = append(errors, fmt.Errorf("hostname is required"))
	}

	// Check for interface consistency
	interfaceNames := make(map[string]bool)
	for _, iface := range config.Interfaces {
		if interfaceNames[iface.Name] {
			errors = append(errors, fmt.Errorf("duplicate interface name: %s", iface.Name))
		}
		interfaceNames[iface.Name] = true
	}

	// Check for VLAN ID consistency
	vlanIDs := make(map[string]bool)
	for _, vlan := range config.VLANs {
		if vlanIDs[vlan.ID] {
			errors = append(errors, fmt.Errorf("duplicate VLAN ID: %s", vlan.ID))
		}
		vlanIDs[vlan.ID] = true

		// Validate VLAN ID range
		if vlanID, err := strconv.Atoi(vlan.ID); err == nil {
			if vlanID < 1 || vlanID > 4094 {
				errors = append(errors, fmt.Errorf("VLAN ID %s is out of valid range (1-4094)", vlan.ID))
			}
		}
	}

	return errors
}

// FortinetConfig represents parsed FortiGate configuration
type FortinetConfig struct {
	Sections map[string]interface{}
	Raw      string
}

// parseFortiConfig parses FortiGate CLI configuration format.
// It correctly handles nested config...end blocks by tracking nesting depth.
func (p *FortinetParser) parseFortiConfig(rawConfig string) (*FortinetConfig, error) {
	config := &FortinetConfig{
		Sections: make(map[string]interface{}),
		Raw:      rawConfig,
	}

	lines := strings.Split(rawConfig, "\n")
	var sectionKey string
	var sectionContent strings.Builder
	depth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "config ") {
			if depth == 0 {
				// Top-level section start
				sectionKey = strings.Join(strings.Fields(trimmed)[1:], " ")
				sectionContent.Reset()
			} else {
				// Nested config block: capture as content
				sectionContent.WriteString(trimmed + "\n")
			}
			depth++
		} else if trimmed == "end" {
			depth--
			if depth == 0 {
				// Close the top-level section
				if sectionKey != "" {
					config.Sections[sectionKey] = sectionContent.String()
					sectionKey = ""
					sectionContent.Reset()
				}
			} else {
				// Closing a nested block: capture as content
				sectionContent.WriteString(trimmed + "\n")
			}
		} else if depth > 0 {
			sectionContent.WriteString(trimmed + "\n")
		}
	}

	return config, nil
}

// parseFortinetInterfaces extracts interface configuration from FortiGate config
func (p *FortinetParser) parseFortinetInterfaces(config *FortinetConfig) ([]configparser.ConfigInterface, error) {
	var interfaces []configparser.ConfigInterface

	// Look for system interface section
	interfaceSection, exists := config.Sections["system interface"]
	if !exists {
		return interfaces, nil
	}

	interfaceBlocks := p.parseConfigBlocks(interfaceSection.(string))

	for _, block := range interfaceBlocks {
		name := p.extractQuotedValue(block, "edit")
		if name == "" {
			continue
		}

		configIface := configparser.ConfigInterface{
			Name:        name,
			Description: p.extractQuotedValue(block, "set description"),
			Enabled:     !strings.Contains(block, "set status down"),
			Type:        p.determineFortinetInterfaceType(block),
		}

		// Parse IP configuration
		if ip := p.extractIPWithMask(block, "set ip"); ip != "" {
			configIface.IPAddresses = append(configIface.IPAddresses, ip)
		}

		// Parse VLAN configuration
		if vlanID := p.extractValue(block, "set vlanid"); vlanID != "" {
			vlan := configparser.ConfigVLAN{
				ID:     vlanID,
				Tagged: true,
			}
			configIface.VLANs = append(configIface.VLANs, vlan)

			// Set parent interface for VLANs
			if parent := p.extractQuotedValue(block, "set interface"); parent != "" {
				configIface.Parent = parent
				configIface.Type = "vlan"
			}
		}

		interfaces = append(interfaces, configIface)
	}

	return interfaces, nil
}

// parseFortinetVLANs extracts VLAN information from FortiGate config
func (p *FortinetParser) parseFortinetVLANs(config *FortinetConfig) ([]configparser.ConfigVLAN, error) {
	var vlans []configparser.ConfigVLAN

	// VLANs are typically defined as VLAN interfaces in FortiGate
	interfaceSection, exists := config.Sections["system interface"]
	if !exists {
		return vlans, nil
	}

	interfaceBlocks := p.parseConfigBlocks(interfaceSection.(string))

	for _, block := range interfaceBlocks {
		vlanID := p.extractValue(block, "set vlanid")
		if vlanID == "" {
			continue
		}

		name := p.extractQuotedValue(block, "edit")
		description := p.extractQuotedValue(block, "set description")

		vlan := configparser.ConfigVLAN{
			ID:          vlanID,
			Name:        name,
			Description: description,
			Enabled:     !strings.Contains(block, "set status down"),
			Tagged:      true,
		}

		vlans = append(vlans, vlan)
	}

	return vlans, nil
}

// parseFortinetRoutes extracts routing information from FortiGate config
func (p *FortinetParser) parseFortinetRoutes(config *FortinetConfig) ([]configparser.ConfigRoute, error) {
	var routes []configparser.ConfigRoute

	// Look for static routes
	routeSection, exists := config.Sections["router static"]
	if !exists {
		return routes, nil
	}

	routeBlocks := p.parseConfigBlocks(routeSection.(string))

	for _, block := range routeBlocks {
		dst := p.extractValue(block, "set dst")
		gateway := p.extractValue(block, "set gateway")
		device := p.extractQuotedValue(block, "set device")

		if dst == "" && gateway == "" {
			continue
		}

		route := configparser.ConfigRoute{
			Network:   dst,
			Gateway:   gateway,
			Interface: device,
		}

		// Parse metric
		if metric := p.extractValue(block, "set distance"); metric != "" {
			if m, err := strconv.Atoi(metric); err == nil {
				route.Metric = m
			}
		}

		routes = append(routes, route)
	}

	return routes, nil
}

// parseFortinetPolicies extracts firewall policies from FortiGate config
func (p *FortinetParser) parseFortinetPolicies(config *FortinetConfig) ([]configparser.ConfigFirewallRule, error) {
	var rules []configparser.ConfigFirewallRule

	// Look for firewall policies
	policySection, exists := config.Sections["firewall policy"]
	if !exists {
		return rules, nil
	}

	policyBlocks := p.parseConfigBlocks(policySection.(string))

	for _, block := range policyBlocks {
		policyID := p.extractValue(block, "edit")
		name := p.extractQuotedValue(block, "set name")

		rule := configparser.ConfigFirewallRule{
			ID:        policyID,
			Name:      name,
			Enabled:   !strings.Contains(block, "set status disable"),
			Action:    p.mapFortinetAction(p.extractValue(block, "set action")),
			Direction: "forward",
		}

		// Parse source zone
		if srcZone := p.extractQuotedValue(block, "set srcintf"); srcZone != "" {
			rule.SourceZone = srcZone
		}

		// Parse destination zone
		if dstZone := p.extractQuotedValue(block, "set dstintf"); dstZone != "" {
			rule.DestZone = dstZone
		}

		// Parse source addresses
		if srcAddr := p.extractMultiValue(block, "set srcaddr"); len(srcAddr) > 0 {
			rule.Source = srcAddr
		}

		// Parse destination addresses
		if dstAddr := p.extractMultiValue(block, "set dstaddr"); len(dstAddr) > 0 {
			rule.Destination = dstAddr
		}

		// Parse services (ports)
		if services := p.extractMultiValue(block, "set service"); len(services) > 0 {
			rule.Ports = services
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// Helper functions for parsing FortiGate configuration

func (p *FortinetParser) parseConfigBlocks(content string) []string {
	var blocks []string
	var currentBlock strings.Builder
	inBlock := false

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "edit ") {
			if inBlock {
				blocks = append(blocks, currentBlock.String())
			}
			currentBlock.Reset()
			currentBlock.WriteString(line + "\n")
			inBlock = true
		} else if line == "next" {
			if inBlock {
				blocks = append(blocks, currentBlock.String())
				currentBlock.Reset()
				inBlock = false
			}
		} else if inBlock {
			currentBlock.WriteString(line + "\n")
		}
	}

	// Handle case where last block doesn't end with "next"
	if inBlock && currentBlock.Len() > 0 {
		blocks = append(blocks, currentBlock.String())
	}

	return blocks
}

func (p *FortinetParser) extractValue(block, keyword string) string {
	re := regexp.MustCompile(keyword + `\s+(\S+)`)
	matches := re.FindStringSubmatch(block)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractIPWithMask extracts an IP address with optional subnet mask from a config block,
// returning CIDR notation (e.g. "192.168.1.1/24") when a mask is present.
func (p *FortinetParser) extractIPWithMask(block, keyword string) string {
	re := regexp.MustCompile(keyword + `\s+(\d+\.\d+\.\d+\.\d+)\s+(\d+\.\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(block)
	if len(matches) == 3 {
		if prefix := netmaskToPrefix(matches[2]); prefix >= 0 {
			return fmt.Sprintf("%s/%d", matches[1], prefix)
		}
		return fmt.Sprintf("%s/%s", matches[1], matches[2])
	}
	// Fall back to bare IP (e.g. management interface with no mask)
	re2 := regexp.MustCompile(keyword + `\s+(\d+\.\d+\.\d+\.\d+)`)
	if m := re2.FindStringSubmatch(block); len(m) > 1 {
		return m[1]
	}
	return ""
}

func (p *FortinetParser) extractQuotedValue(block, keyword string) string {
	re := regexp.MustCompile(keyword + `\s+"([^"]+)"`)
	matches := re.FindStringSubmatch(block)
	if len(matches) > 1 {
		return matches[1]
	}

	// Try without quotes
	re = regexp.MustCompile(keyword + `\s+(\S+)`)
	matches = re.FindStringSubmatch(block)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (p *FortinetParser) extractMultiValue(block, keyword string) []string {
	re := regexp.MustCompile(keyword + `\s+"([^"]+)"`)
	matches := re.FindAllStringSubmatch(block, -1)

	var values []string
	for _, match := range matches {
		if len(match) > 1 {
			// Split multiple values in quotes
			parts := strings.Fields(match[1])
			values = append(values, parts...)
		}
	}

	return values
}

func (p *FortinetParser) determineFortinetInterfaceType(block string) string {
	if strings.Contains(block, "set vlanid") {
		return "vlan"
	}
	if strings.Contains(block, "set type aggregate") {
		return "aggregate"
	}
	if strings.Contains(block, "set type tunnel") {
		return "tunnel"
	}
	return "physical"
}

func (p *FortinetParser) mapFortinetAction(action string) string {
	switch strings.ToLower(action) {
	case "accept", "allow":
		return "allow"
	case "deny", "drop":
		return "deny"
	default:
		return action
	}
}

func extractFortinetHostname(config *FortinetConfig) string {
	globalSection, exists := config.Sections["system global"]
	if !exists {
		return ""
	}

	content := globalSection.(string)
	re := regexp.MustCompile(`set hostname\s+"?([^"\s]+)"?`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.Trim(matches[1], "\"")
	}

	return ""
}

func extractFortinetVersion(config *FortinetConfig) string {
	// Try to extract version from config
	lines := strings.Split(config.Raw, "\n")
	for _, line := range lines {
		if strings.Contains(line, "#config-version") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return time.Now().Format("2006-01-02")
}

func extractFortinetModel(config *FortinetConfig, deviceInfo s.SNMPDevice) string {
	// Try to extract model from system global
	globalSection, exists := config.Sections["system global"]
	if exists {
		content := globalSection.(string)
		if strings.Contains(content, "FortiGate") {
			return "FortiGate"
		}
	}

	// Fallback to SNMP description
	if strings.Contains(deviceInfo.SysDescr, "FortiGate") {
		return deviceInfo.SysDescr
	}

	return "FortiGate Firewall"
}
