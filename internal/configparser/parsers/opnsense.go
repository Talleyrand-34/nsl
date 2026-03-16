// Legacy XML-based OPNsense parser.
// Use the FreeBSD ifconfig parser (--device-type opnsense) instead.
// Still accessible via --device-type opnsense-xml for backwards compatibility.
package parsers

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"nsl-graph/internal/configparser"
	s "nsl-graph/internal/scanner"
)

// OPNsenseParser handles parsing of OPNsense XML configuration files
type OPNsenseParser struct{}

// NewOPNsenseParser creates a new OPNsense configuration parser
func NewOPNsenseParser() *OPNsenseParser {
	return &OPNsenseParser{}
}

func init() { configparser.DefaultRegistry.RegisterParser(NewOPNsenseParser()) }

// GetDeviceType returns the device type key for this legacy parser.
func (p *OPNsenseParser) GetDeviceType() string {
	return "opnsense-xml"
}

// SupportsDevice always returns false — auto-detection is handled by FreeBSDParser.
// Use --device-type opnsense-xml to select this parser explicitly.
func (p *OPNsenseParser) SupportsDevice(_ s.SNMPDevice) bool {
	return false
}

// ParseConfig parses raw OPNsense XML configuration and returns structured ConfigData
func (p *OPNsenseParser) ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*configparser.ConfigData, error) {
	// Strip any SSH banner / prompt text that precedes the XML document.
	xmlStart := strings.Index(rawConfig, "<?xml")
	if xmlStart < 0 {
		xmlStart = strings.Index(rawConfig, "<opnsense")
	}
	if xmlStart > 0 {
		rawConfig = rawConfig[xmlStart:]
	}

	var config OPNsenseConfig
	if err := xml.Unmarshal([]byte(rawConfig), &config); err != nil {
		return nil, fmt.Errorf("failed to parse OPNsense XML config: %w", err)
	}

	configData := &configparser.ConfigData{
		DeviceType:    p.GetDeviceType(),
		DeviceModel:   extractDeviceModel(config),
		Hostname:      config.System.Hostname,
		Domain:        config.System.Domain,
		ConfigVersion: config.Version,
		Source:        configparser.ConfigSourceSSH, // Will be set by caller
		ParsedAt:      time.Now(),
		Raw:           rawConfig,
	}

	// Parse interfaces
	interfaces, err := p.parseInterfaces(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interfaces: %w", err)
	}
	configData.Interfaces = interfaces

	// Parse VLANs
	vlans, err := p.parseVLANs(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VLANs: %w", err)
	}
	configData.VLANs = vlans

	// Parse routes
	routes, err := p.parseRoutes(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse routes: %w", err)
	}
	configData.Routes = routes

	// Parse firewall rules
	firewallRules, err := p.parseFirewallRules(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse firewall rules: %w", err)
	}
	configData.FirewallRules = firewallRules

	return configData, nil
}

// isXMLContent returns true if the string looks like OPNsense XML config.
func isXMLContent(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "<?xml") || strings.HasPrefix(s, "<opnsense")
}

// GetConfigViaSSH retrieves OPNsense configuration via SSH.
// Tries cat /conf/config.xml first (most reliable), then configctl as fallback.
func (p *OPNsenseParser) GetConfigViaSSH(ip string, creds configparser.SSHCredentials) (string, error) {
	client := configparser.NewSSHClient(creds)
	if err := client.Connect(ip); err != nil {
		return "", fmt.Errorf("failed to connect to OPNsense device: %w", err)
	}
	defer client.Close()

	// Read config file directly — most reliable across OPNsense versions.
	output, err := client.Execute("cat /conf/config.xml")
	if err == nil && isXMLContent(output) {
		return output, nil
	}

	// Fallback: configctl (may not exist or may be restricted on some setups).
	output2, err2 := client.Execute("configctl system config show")
	if err2 == nil && isXMLContent(output2) {
		return output2, nil
	}

	// If cat returned an error but somehow got valid XML, use it.
	if isXMLContent(output) {
		return output, nil
	}

	if err != nil {
		return "", fmt.Errorf("failed to retrieve OPNsense configuration: %w", err)
	}
	return "", fmt.Errorf("failed to retrieve OPNsense XML config: unexpected output: %s", strings.TrimSpace(output))
}

// ValidateConfig performs basic validation on parsed OPNsense configuration
func (p *OPNsenseParser) ValidateConfig(config *configparser.ConfigData) []error {
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

		if iface.Name == "" {
			errors = append(errors, fmt.Errorf("interface name cannot be empty"))
		}
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

// parseInterfaces extracts interface information from OPNsense config
func (p *OPNsenseParser) parseInterfaces(config OPNsenseConfig) ([]configparser.ConfigInterface, error) {
	var interfaces []configparser.ConfigInterface

	// Parse regular interfaces
	for name, iface := range config.Interfaces {
		if name == "" || iface.Disabled {
			continue
		}

		configIface := configparser.ConfigInterface{
			Name:        name,
			Description: iface.Descr,
			Enabled:     iface.Enable != "" && iface.Enable != "0",
			Type:        "physical",
		}

		// Parse IP addresses
		if iface.IP != "" {
			if iface.Subnet != "" {
				configIface.IPAddresses = append(configIface.IPAddresses, fmt.Sprintf("%s/%s", iface.IP, iface.Subnet))
			} else {
				configIface.IPAddresses = append(configIface.IPAddresses, iface.IP)
			}
		}

		// Parse VLAN configuration
		if iface.VlanTag != "" {
			vlan := configparser.ConfigVLAN{
				ID:      iface.VlanTag,
				Tagged:  true,
			}
			configIface.VLANs = append(configIface.VLANs, vlan)
		}

		interfaces = append(interfaces, configIface)
	}

	// Parse VLAN interfaces
	for _, vlan := range config.VLANs {
		configIface := configparser.ConfigInterface{
			Name:        fmt.Sprintf("%s.%s", vlan.If, vlan.Tag),
			Description: vlan.Descr,
			Enabled:     true,
			Type:        "vlan",
			Parent:      vlan.If,
		}

		// Add VLAN configuration
		vlanConfig := configparser.ConfigVLAN{
			ID:     vlan.Tag,
			Tagged: true,
		}
		configIface.VLANs = append(configIface.VLANs, vlanConfig)

		interfaces = append(interfaces, configIface)
	}

	return interfaces, nil
}

// parseVLANs extracts VLAN information from OPNsense config
func (p *OPNsenseParser) parseVLANs(config OPNsenseConfig) ([]configparser.ConfigVLAN, error) {
	var vlans []configparser.ConfigVLAN

	for _, vlan := range config.VLANs {
		configVLAN := configparser.ConfigVLAN{
			ID:          vlan.Tag,
			Name:        fmt.Sprintf("VLAN_%s", vlan.Tag),
			Description: vlan.Descr,
			Enabled:     true,
			Tagged:      true,
		}

		vlans = append(vlans, configVLAN)
	}

	return vlans, nil
}

// parseRoutes extracts routing information from OPNsense config
func (p *OPNsenseParser) parseRoutes(config OPNsenseConfig) ([]configparser.ConfigRoute, error) {
	var routes []configparser.ConfigRoute

	// Parse static routes
	for _, route := range config.StaticRoutes {
		configRoute := configparser.ConfigRoute{
			Network:     route.Network,
			Gateway:     route.Gateway,
			Description: route.Descr,
		}

		routes = append(routes, configRoute)
	}

	// Add default gateway if configured
	for _, iface := range config.Interfaces {
		if iface.Gateway != "" {
			configRoute := configparser.ConfigRoute{
				Network:     "0.0.0.0/0",
				Gateway:     iface.Gateway,
				Interface:   iface.If,
				Description: "Default gateway",
			}
			routes = append(routes, configRoute)
			break // Only add one default route
		}
	}

	return routes, nil
}

// parseFirewallRules extracts firewall rules from OPNsense config
func (p *OPNsenseParser) parseFirewallRules(config OPNsenseConfig) ([]configparser.ConfigFirewallRule, error) {
	var rules []configparser.ConfigFirewallRule

	for i, rule := range config.Filter.Rules {
		if rule.Disabled {
			continue
		}

		configRule := configparser.ConfigFirewallRule{
			ID:          fmt.Sprintf("rule_%d", i),
			Name:        rule.Descr,
			Enabled:     !rule.Disabled,
			Action:      mapOPNsenseAction(rule.Type),
			Direction:   mapOPNsenseDirection(rule.Direction),
			Protocol:    rule.Protocol,
		}

		// Parse source
		if rule.Source.Any {
			configRule.Source = append(configRule.Source, "any")
		} else if rule.Source.Address != "" {
			configRule.Source = append(configRule.Source, rule.Source.Address)
		}

		// Parse destination
		if rule.Destination.Any {
			configRule.Destination = append(configRule.Destination, "any")
		} else if rule.Destination.Address != "" {
			configRule.Destination = append(configRule.Destination, rule.Destination.Address)
		}

		// Parse ports
		if rule.Destination.Port != "" {
			configRule.Ports = append(configRule.Ports, rule.Destination.Port)
		}

		rules = append(rules, configRule)
	}

	return rules, nil
}

// Helper functions for mapping OPNsense-specific values

func mapOPNsenseAction(actionType string) string {
	switch strings.ToLower(actionType) {
	case "pass":
		return "allow"
	case "block", "reject":
		return "deny"
	default:
		return actionType
	}
}

func mapOPNsenseDirection(direction string) string {
	switch strings.ToLower(direction) {
	case "in":
		return "in"
	case "out":
		return "out"
	default:
		return "forward"
	}
}

func extractDeviceModel(config OPNsenseConfig) string {
	// Try to extract model information from various sources
	if config.System.Product != "" {
		return config.System.Product
	}
	if config.System.Platform != "" {
		return config.System.Platform
	}
	return "OPNsense Firewall"
}

// OPNsense XML configuration structures

// OPNsenseInterfaceMap decodes named children of <interfaces> into a Go map.
// encoding/xml cannot decode map[string]T directly, so we implement xml.Unmarshaler.
type OPNsenseInterfaceMap map[string]OPNsenseInterface

func (m *OPNsenseInterfaceMap) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*m = make(OPNsenseInterfaceMap)
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			var iface OPNsenseInterface
			if err := d.DecodeElement(&iface, &t); err != nil {
				return err
			}
			(*m)[t.Name.Local] = iface
		case xml.EndElement:
			return nil
		}
	}
}

type OPNsenseConfig struct {
	XMLName      xml.Name             `xml:"opnsense"`
	Version      string               `xml:"version"`
	System       OPNsenseSystem       `xml:"system"`
	Interfaces   OPNsenseInterfaceMap `xml:"interfaces"`
	VLANs        []OPNsenseVLAN       `xml:"vlans>vlan"`
	StaticRoutes []OPNsenseStaticRoute `xml:"staticroutes>route"`
	Filter       OPNsenseFilter       `xml:"filter"`
}

type OPNsenseSystem struct {
	Hostname string `xml:"hostname"`
	Domain   string `xml:"domain"`
	Product  string `xml:"product"`
	Platform string `xml:"platform"`
}

type OPNsenseInterface struct {
	If       string `xml:"if"`
	Descr    string `xml:"descr"`
	Enable   string `xml:"enable"`
	Disabled bool   `xml:"disabled"`
	IP       string `xml:"ipaddr"`
	Subnet   string `xml:"subnet"`
	Gateway  string `xml:"gateway"`
	VlanTag  string `xml:"tag"`
}

type OPNsenseVLAN struct {
	If    string `xml:"if"`
	Tag   string `xml:"tag"`
	Descr string `xml:"descr"`
}

type OPNsenseStaticRoute struct {
	Network string `xml:"network"`
	Gateway string `xml:"gateway"`
	Descr   string `xml:"descr"`
}

type OPNsenseFilter struct {
	Rules []OPNsenseFirewallRule `xml:"rule"`
}

type OPNsenseFirewallRule struct {
	Type        string                      `xml:"type"`
	Interface   string                      `xml:"interface"`
	Direction   string                      `xml:"direction"`
	Protocol    string                      `xml:"protocol"`
	Source      OPNsenseFirewallAddress     `xml:"source"`
	Destination OPNsenseFirewallAddress     `xml:"destination"`
	Descr       string                      `xml:"descr"`
	Disabled    bool                        `xml:"disabled"`
}

type OPNsenseFirewallAddress struct {
	Any     bool   `xml:"any"`
	Address string `xml:"address"`
	Port    string `xml:"port"`
}