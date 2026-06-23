package parsers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"nsl-graph/internal/configparser"
	s "nsl-graph/internal/scanner"
)

// OpenWrtParser handles parsing of OpenWrt UCI configuration files
type OpenWrtParser struct{}

// NewOpenWrtParser creates a new OpenWrt configuration parser
func NewOpenWrtParser() *OpenWrtParser {
	return &OpenWrtParser{}
}

func init() { configparser.DefaultRegistry.RegisterParser(NewOpenWrtParser()) }

// GetDeviceType returns the device type this parser handles
func (p *OpenWrtParser) GetDeviceType() string {
	return "openwrt"
}

// SupportsDevice returns true if this parser can handle the given device
func (p *OpenWrtParser) SupportsDevice(device s.SNMPDevice) bool {
	descr := strings.ToLower(device.SysDescr)
	return strings.Contains(descr, "linux") &&
		(strings.Contains(descr, "openwrt") || strings.Contains(device.SysName, "OpenWrt"))
}

// ParseConfig parses raw OpenWrt UCI configuration and returns structured ConfigData
func (p *OpenWrtParser) ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*configparser.ConfigData, error) {
	uciConfig, err := p.parseUCIConfig(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UCI config: %w", err)
	}

	configData := &configparser.ConfigData{
		DeviceType:    p.GetDeviceType(),
		DeviceModel:   extractOpenWrtModel(uciConfig, deviceInfo),
		Hostname:      extractHostname(uciConfig),
		ConfigVersion: extractConfigVersion(uciConfig),
		Source:        configparser.ConfigSourceSSH,
		ParsedAt:      time.Now(),
		Raw:           rawConfig,
	}

	// Extract board.json for physical port information
	boardPorts, _ := p.parseBoardJSON(rawConfig)
	if len(boardPorts) > 0 {
		// Parse swconfig VLAN data and populate into switch ports
		// Note: swconfig only works on older non-DSA switches. DSA switches
		// (like Aircube AC) don't have swconfig and VLAN info must come from
		// bridge vlan show or network config parsing instead.
		swconfigVLANs := p.parseSwconfigVLANs(rawConfig)
		for i := range boardPorts {
			if vlans, ok := swconfigVLANs[boardPorts[i].PortNumber]; ok {
				boardPorts[i].VLANs = vlans
			}
		}
		configData.SwitchPorts = boardPorts
	}

	// Parse interfaces
	interfaces, err := p.parseNetworkInterfaces(uciConfig, configData.SwitchPorts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interfaces: %w", err)
	}
	configData.Interfaces = interfaces

	// Parse VLANs
	vlans, err := p.parseNetworkVLANs(uciConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VLANs: %w", err)
	}
	configData.VLANs = vlans

	// Parse routes
	routes, err := p.parseNetworkRoutes(uciConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse routes: %w", err)
	}
	configData.Routes = routes

	// Parse firewall rules
	firewallRules, err := p.parseFirewallRules(uciConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse firewall rules: %w", err)
	}
	configData.FirewallRules = firewallRules

	// Parse wireless interfaces (radios + SSIDs)
	wifiIfaces, err := p.parseWirelessInterfaces(uciConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wireless interfaces: %w", err)
	}
	configData.Interfaces = append(configData.Interfaces, wifiIfaces...)

	return configData, nil
}

// GetConfigViaSSH retrieves OpenWrt configuration via SSH
func (p *OpenWrtParser) GetConfigViaSSH(ip string, creds configparser.SSHCredentials) (string, error) {
	client := configparser.NewSSHClient(creds)
	if err := client.Connect(ip); err != nil {
		return "", fmt.Errorf("failed to connect to OpenWrt device: %w", err)
	}
	defer client.Close()

	commands := []string{
		"uci show network",
		"uci show wireless",
		"uci show firewall",
		"uci show system",
		"uci show dhcp",
		"cat /etc/board.json",
		"swconfig dev switch0 show",
	}

	var allConfig strings.Builder
	for _, cmd := range commands {
		output, err := client.Execute(cmd)
		if err != nil {
			allConfig.WriteString(fmt.Sprintf("# Error executing %s: %v\n", cmd, err))
			continue
		}
		allConfig.WriteString(fmt.Sprintf("# %s\n", cmd))
		allConfig.WriteString(output)
		allConfig.WriteString("\n")
	}

	return allConfig.String(), nil
}

// parseBoardJSON parses /etc/board.json and returns physical switch ports.
// It extracts physical port information from the switch configuration.
func (p *OpenWrtParser) parseBoardJSON(rawConfig string) ([]configparser.SwitchPortInfo, error) {
	var ports []configparser.SwitchPortInfo

	lines := strings.Split(rawConfig, "\n")
	var jsonLines []string
	collecting := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "cat /etc/board.json") {
			collecting = true
			continue
		}
		if collecting {
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "network.") || strings.HasPrefix(trimmed, "wireless.") {
				if len(jsonLines) > 0 {
					break
				}
				continue
			}
			jsonLines = append(jsonLines, line)
		}
	}

	if len(jsonLines) == 0 {
		return ports, nil
	}

	jsonStr := strings.Join(jsonLines, "\n")

	var board struct {
		Switch map[string]struct {
			Ports []struct {
				Num       int    `json:"num"`
				Device    string `json:"device"`
				NeedTag   bool   `json:"need_tag"`
				WantUntag bool   `json:"want_untag"`
				Role      string `json:"role"`
				Index     int    `json:"index"`
			} `json:"ports"`
			Roles []struct {
				Role   string `json:"role"`
				Ports  string `json:"ports"`
				Device string `json:"device"`
			} `json:"roles"`
		} `json:"switch"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &board); err != nil {
		return ports, nil
	}

	portIndex := 0
	for switchName, sw := range board.Switch {
		_ = switchName
		for _, port := range sw.Ports {
			if port.Device == "eth0" {
				continue
			}
			if port.Role == "" {
				continue
			}

			portName := port.Role
			if port.Index > 0 {
				portName = fmt.Sprintf("%s%d", port.Role, port.Index)
			} else {
				portName = fmt.Sprintf("%s%d", port.Role, portIndex)
				portIndex++
			}

			linkStatus := "down"
			for _, role := range sw.Roles {
				if strings.Contains(role.Ports, strconv.Itoa(port.Num)) {
					linkStatus = "up"
					break
				}
			}

			ports = append(ports, configparser.SwitchPortInfo{
				PortNumber: port.Num,
				PortName:   portName,
				LinkStatus: linkStatus,
				Role:       port.Role,
			})
		}
	}

	return ports, nil
}

// parseSwconfigVLANs parses swconfig VLAN output and returns VLAN-port membership
// Output format from command: "VLAN 1: ports: 0t 2 4"
// Returns: map[portNumber][]PortVLANInfo where VID is the actual VLAN ID
func (p *OpenWrtParser) parseSwconfigVLANs(rawConfig string) map[int][]configparser.PortVLANInfo {
	portVLANs := make(map[int]map[int]bool)

	lines := strings.Split(rawConfig, "\n")
	currentVlan := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip lines starting with # (comments/errors) and Port lines
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "Port ") || strings.HasPrefix(trimmed, "switch0:") {
			continue
		}

		// Detect start of VLAN block: "VLAN 1:" or "VLAN 10:"
		if strings.HasPrefix(trimmed, "VLAN ") {
			parts := strings.Split(trimmed, " ")
			if len(parts) >= 2 {
				vlanNum, err := strconv.Atoi(parts[1][:len(parts[1])-1]) // Remove trailing :
				if err == nil {
					currentVlan = vlanNum
				}
			}
			continue
		}

		// Within VLAN block, look for "vid: N" line
		if strings.HasPrefix(trimmed, "vid:") {
			vidStr := strings.TrimPrefix(trimmed, "vid:")
			vidStr = strings.TrimSpace(vidStr)
			vlanNum, err := strconv.Atoi(vidStr)
			if err == nil {
				currentVlan = vlanNum
			}
			continue
		}

		// Within VLAN block, look for "ports: \"0t 2 4\"" line
		if strings.HasPrefix(trimmed, "ports:") {
			if currentVlan == 0 {
				continue
			}
			portsStr := strings.TrimPrefix(trimmed, "ports:")
			portsStr = strings.TrimSpace(portsStr)
			// Remove surrounding quotes if present: "\"0t 2 4\"" -> "0t 2 4"
			portsStr = strings.Trim(portsStr, "\"")

			portList := strings.Fields(portsStr)
			for _, portEntry := range portList {
				tagged := strings.HasSuffix(portEntry, "t")
				portNumStr := strings.TrimSuffix(portEntry, "t")
				portNum, err := strconv.Atoi(portNumStr)
				if err != nil {
					continue
				}

				// Skip port 0 (CPU port)
				if portNum == 0 {
					continue
				}

				if portVLANs[portNum] == nil {
					portVLANs[portNum] = make(map[int]bool)
				}
				portVLANs[portNum][currentVlan] = tagged
			}
			currentVlan = 0 // Reset after processing ports
		}
	}

	// Convert to map[portNumber][]PortVLANInfo
	result := make(map[int][]configparser.PortVLANInfo)
	for portNum, vlans := range portVLANs {
		for vid, tagged := range vlans {
			result[portNum] = append(result[portNum], configparser.PortVLANInfo{
				VID:    strconv.Itoa(vid),
				Tagged: tagged,
			})
		}
	}

	return result
}

// ValidateConfig performs basic validation on parsed OpenWrt configuration
func (p *OpenWrtParser) ValidateConfig(config *configparser.ConfigData) []error {
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

// UCIConfig represents parsed UCI configuration
type UCIConfig struct {
	Sections map[string]map[string]UCISection
}

type UCISection struct {
	Type    string
	Options map[string]string
	Lists   map[string][]string
}

// parseUCIConfig parses UCI format configuration
func (p *OpenWrtParser) parseUCIConfig(rawConfig string) (*UCIConfig, error) {
	config := &UCIConfig{
		Sections: make(map[string]map[string]UCISection),
	}

	lines := strings.Split(rawConfig, "\n")
	sectionRegex := regexp.MustCompile(`^([^.]+)\.([^=]+)=(.*)$`)
	optionRegex := regexp.MustCompile(`^([^.]+)\.([^.]+)\.([^=]+)=(.*)$`)
	listRegex := regexp.MustCompile(`^([^.]+)\.([^.]+)\.([^=]+)\[\]=(.*)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse list entries (package.section.option[]=value)
		if matches := listRegex.FindStringSubmatch(line); len(matches) == 5 {
			pkg, section, option, value := matches[1], matches[2], matches[3], stripUCIQuotes(matches[4])

			if config.Sections[pkg] == nil {
				config.Sections[pkg] = make(map[string]UCISection)
			}

			if sec, exists := config.Sections[pkg][section]; exists {
				if sec.Lists == nil {
					sec.Lists = make(map[string][]string)
				}
				sec.Lists[option] = append(sec.Lists[option], value)
				config.Sections[pkg][section] = sec
			}
			continue
		}

		// Parse options (package.section.option=value)
		if matches := optionRegex.FindStringSubmatch(line); len(matches) == 5 {
			pkg, section, option, value := matches[1], matches[2], matches[3], stripUCIQuotes(matches[4])

			if config.Sections[pkg] == nil {
				config.Sections[pkg] = make(map[string]UCISection)
			}

			sec := config.Sections[pkg][section]
			if sec.Options == nil {
				sec.Options = make(map[string]string)
			}
			sec.Options[option] = value
			config.Sections[pkg][section] = sec
			continue
		}

		// Parse section definitions (package.section=type)
		if matches := sectionRegex.FindStringSubmatch(line); len(matches) == 4 {
			pkg, section, sectionType := matches[1], matches[2], stripUCIQuotes(matches[3])

			if config.Sections[pkg] == nil {
				config.Sections[pkg] = make(map[string]UCISection)
			}

			sec := config.Sections[pkg][section]
			sec.Type = sectionType
			if sec.Options == nil {
				sec.Options = make(map[string]string)
			}
			config.Sections[pkg][section] = sec
		}
	}

	return config, nil
}

// parseNetworkInterfaces extracts interface information from UCI network config
func (p *OpenWrtParser) parseNetworkInterfaces(config *UCIConfig, switchPorts []configparser.SwitchPortInfo) ([]configparser.ConfigInterface, error) {
	var interfaces []configparser.ConfigInterface

	network, exists := config.Sections["network"]
	if !exists {
		return interfaces, nil
	}

	physicalDevs := make(map[string]bool)
	// eth0Referenced records that eth0 is used as an uplink/bridge member. On DSA
	// devices eth0 is the CPU/conduit port (excluded), but on single-port devices
	// (e.g. APs) eth0 is the only real port — emitted below when no other ethN exists.
	eth0Referenced := false
	// Track VLANs declared on bridge devices so they can be propagated to logical interfaces
	bridgeVLANs := make(map[string][]configparser.ConfigVLAN)

	for name, section := range network {
		if section.Type != "interface" && section.Type != "device" {
			continue
		}

		// For "device" sections, the real kernel name is in Options["name"]
		ifaceName := name
		if section.Type == "device" {
			if devName := section.Options["name"]; devName != "" {
				ifaceName = devName
			}
		}

		configIface := configparser.ConfigInterface{
			Name:        ifaceName,
			Description: section.Options["description"],
			Enabled:     section.Options["enabled"] != "0",
			Type:        determineInterfaceType(section),
		}

		// Parse IP configuration
		proto := section.Options["proto"]
		if proto == "static" {
			if ipaddr := section.Options["ipaddr"]; ipaddr != "" {
				if netmask := section.Options["netmask"]; netmask != "" {
					if prefix := netmaskToPrefix(netmask); prefix >= 0 {
						configIface.IPAddresses = append(configIface.IPAddresses, fmt.Sprintf("%s/%d", ipaddr, prefix))
					} else {
						configIface.IPAddresses = append(configIface.IPAddresses, fmt.Sprintf("%s/%s", ipaddr, netmask))
					}
				} else {
					configIface.IPAddresses = append(configIface.IPAddresses, ipaddr)
				}
			}
		}

		// Parse VLAN configuration
		if section.Options["type"] == "8021q" {
			if vid := section.Options["vid"]; vid != "" {
				vlan := configparser.ConfigVLAN{
					ID:     vid,
					Tagged: true,
				}
				configIface.VLANs = append(configIface.VLANs, vlan)
			}
		}

		// For bridge device sections, extract physical members from "ports" (e.g., "eth0.2")
		// eth0.2 → base=eth0 (physical parent), vlanID=2 (VLAN membership on bridge side)
		if section.Type == "device" && configIface.Type == "bridge" {
			for _, port := range unquotedFields(section.Options["ports"]) {
				parts := strings.SplitN(port, ".", 2)
				base := parts[0]

				// Skip loopback
				if base == "lo" {
					continue
				}

				// If base is eth0 and we have a VLAN suffix (e.g., eth0.4), emit the VLAN subinterface
				if base == "eth0" && len(parts) == 2 && parts[1] != "" {
					// Emit eth0 as a logical interface if not already emitted (it's the CPU port)
					eth0Exists := false
					for _, iface := range interfaces {
						if iface.Name == "eth0" {
							eth0Exists = true
							break
						}
					}
					if !eth0Exists {
						interfaces = append(interfaces, configparser.ConfigInterface{
							Name:   "eth0",
							Type:   "logical",
							Parent: "",
						})
					}

					vlanIfaceName := port // e.g., "eth0.4"
					// Only emit if not already emitted
					alreadyEmitted := false
					for _, iface := range interfaces {
						if iface.Name == vlanIfaceName {
							alreadyEmitted = true
							break
						}
					}
					if !alreadyEmitted {
						interfaces = append(interfaces, configparser.ConfigInterface{
							Name:   vlanIfaceName,
							Parent: "eth0",
							Type:   "vlan",
						})
					}
					// Set this VLAN subinterface as the bridge's parent
					configIface.Parent = vlanIfaceName
					// Extract VLAN ID
					configIface.VLANs = append(configIface.VLANs, configparser.ConfigVLAN{
						ID:     parts[1],
						Tagged: false,
					})
					continue
				}

				if base == "eth0" {
					eth0Referenced = true
				}
				if base == "" || base == "eth0" {
					continue
				}
				physicalDevs[base] = true
				if configIface.Parent == "" {
					configIface.Parent = base
				}
				if len(parts) == 2 && parts[1] != "" {
					configIface.VLANs = append(configIface.VLANs, configparser.ConfigVLAN{
						ID:     parts[1],
						Tagged: false,
					})
				}
			}
			// Record this bridge's VLANs for propagation to logical interfaces later
			if len(configIface.VLANs) > 0 {
				bridgeVLANs[ifaceName] = configIface.VLANs
			}
		}

		// Set parent interface reference
		// Newer OpenWrt uses 'device' instead of 'ifname'
		ifname := section.Options["ifname"]
		if ifname == "" {
			ifname = section.Options["device"]
		}
		if devs := unquotedFields(ifname); len(devs) > 0 {
			firstDev := devs[0]
			if strings.Contains(firstDev, ".") {
				// VLAN sub-interface: eth0.10
				parts := strings.Split(firstDev, ".")
				configIface.Parent = parts[0]
				configIface.Type = "vlan"
			} else {
				configIface.Parent = firstDev
			}
			// Track the physical device name for later
			physRoot := firstDev
			if strings.Contains(physRoot, ".") {
				physRoot = strings.Split(physRoot, ".")[0]
			}
			if physRoot == "eth0" {
				eth0Referenced = true
			}
			if physRoot != "" && physRoot != "lo" && physRoot != "eth0" {
				physicalDevs[physRoot] = true
			}
		}

		interfaces = append(interfaces, configIface)
	}

	// Propagate bridge VLANs to logical interfaces that sit on top of a bridge.
	// e.g., mgmt2 → parent br-2-mgmt2 → VLAN 2; copy VLAN 2 onto mgmt2 so
	// the import pipeline can map its IP to the correct VLAN.
	for i, iface := range interfaces {
		if iface.Type != "logical" || iface.Parent == "" || len(iface.VLANs) > 0 {
			continue
		}
		if vlans, ok := bridgeVLANs[iface.Parent]; ok {
			interfaces[i].VLANs = append(interfaces[i].VLANs, vlans...)
		}
	}

	// Emit physical/bridge devices referenced by interface sections
	// (ensures they get DevicePorts during import)
	for devName := range physicalDevs {
		if devName == "eth0" {
			continue
		}
		// Skip if already emitted as a "device" section
		alreadyEmitted := false
		for _, iface := range interfaces {
			if iface.Name == devName {
				alreadyEmitted = true
				break
			}
		}
		if alreadyEmitted {
			continue
		}
		interfaces = append(interfaces, configparser.ConfigInterface{
			Name:    devName,
			Enabled: true,
			Type:    determineDeviceType(devName),
		})
	}

	// On a single-port device eth0 is the real uplink, not a CPU port. When it is
	// referenced and there is no other ethN user port, make it a physical port
	// (converting the eth0 interface emitted as a VLAN parent, or adding one) so
	// it becomes a DevicePort that LLDP adjacencies can resolve against.
	if eth0Referenced {
		hasUserEth := false
		for _, iface := range interfaces {
			if len(iface.Name) > 3 && strings.HasPrefix(iface.Name, "eth") && iface.Name != "eth0" &&
				!strings.Contains(iface.Name, ".") {
				hasUserEth = true
			}
		}
		if !hasUserEth {
			converted := false
			for i := range interfaces {
				if interfaces[i].Name == "eth0" {
					interfaces[i].Type = "physical"
					converted = true
					break
				}
			}
			if !converted {
				interfaces = append(interfaces, configparser.ConfigInterface{
					Name:    "eth0",
					Enabled: true,
					Type:    "physical",
				})
			}
		}
	}

	// Add physical switch ports from board.json (lan1, lan2, lan3, wan0, etc.)
	// These are emitted as interfaces so they become DevicePorts, but marked
	// as type "switch-port" so they can be identified as physical switch ports
	// rather than regular logical interfaces.
	for _, port := range switchPorts {
		if port.Role == "" || port.Role == "cpu" {
			continue
		}

		portName := port.PortName
		if portName == "" {
			portName = fmt.Sprintf("%s%d", port.Role, port.PortNumber)
		}

		interfaces = append(interfaces, configparser.ConfigInterface{
			Name:    portName,
			Enabled: port.LinkStatus == "up",
			Type:    "switch-port",
		})
	}

	return interfaces, nil
}

// determineDeviceType returns "bridge" for br-* names, "physical" otherwise.
func determineDeviceType(name string) string {
	if strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "br_") {
		return "bridge"
	}
	return "physical"
}

// parseNetworkVLANs extracts VLAN information from UCI network config
func (p *OpenWrtParser) parseNetworkVLANs(config *UCIConfig) ([]configparser.ConfigVLAN, error) {
	var vlans []configparser.ConfigVLAN

	network, exists := config.Sections["network"]
	if !exists {
		return vlans, nil
	}

	for name, section := range network {
		if section.Type == "device" && section.Options["type"] == "8021q" {
			vid := section.Options["vid"]
			if vid == "" {
				continue
			}

			vlan := configparser.ConfigVLAN{
				ID:          vid,
				Name:        name,
				Description: section.Options["description"],
				Enabled:     section.Options["enabled"] != "0",
				Tagged:      true,
			}

			vlans = append(vlans, vlan)
		}

		// Also check for bridge VLANs
		if section.Type == "bridge-vlan" {
			vid := section.Options["vlan"]
			if vid == "" {
				continue
			}

			vlan := configparser.ConfigVLAN{
				ID:          vid,
				Name:        fmt.Sprintf("bridge_vlan_%s", vid),
				Description: section.Options["description"],
				Enabled:     true,
				Tagged:      true,
			}

			vlans = append(vlans, vlan)
		}
	}

	return vlans, nil
}

// parseNetworkRoutes extracts routing information from UCI network config
func (p *OpenWrtParser) parseNetworkRoutes(config *UCIConfig) ([]configparser.ConfigRoute, error) {
	var routes []configparser.ConfigRoute

	network, exists := config.Sections["network"]
	if !exists {
		return routes, nil
	}

	for _, section := range network {
		if section.Type == "route" || section.Type == "route6" {
			route := configparser.ConfigRoute{
				Network:     section.Options["target"],
				Gateway:     section.Options["gateway"],
				Interface:   section.Options["interface"],
				Description: section.Options["description"],
			}

			if metric := section.Options["metric"]; metric != "" {
				if m, err := strconv.Atoi(metric); err == nil {
					route.Metric = m
				}
			}

			routes = append(routes, route)
		}

		// Check for default gateway in interfaces
		if section.Type == "interface" {
			if gateway := section.Options["gateway"]; gateway != "" {
				route := configparser.ConfigRoute{
					Network:     "0.0.0.0/0",
					Gateway:     gateway,
					Description: "Default gateway",
				}
				routes = append(routes, route)
			}
		}
	}

	return routes, nil
}

// parseFirewallRules extracts firewall rules from UCI firewall config
func (p *OpenWrtParser) parseFirewallRules(config *UCIConfig) ([]configparser.ConfigFirewallRule, error) {
	var rules []configparser.ConfigFirewallRule

	firewall, exists := config.Sections["firewall"]
	if !exists {
		return rules, nil
	}

	for name, section := range firewall {
		if section.Type != "rule" {
			continue
		}

		rule := configparser.ConfigFirewallRule{
			ID:        name,
			Name:      section.Options["name"],
			Enabled:   section.Options["enabled"] != "0",
			Action:    mapOpenWrtTarget(section.Options["target"]),
			Direction: "forward", // OpenWrt rules are typically forward rules
			Protocol:  section.Options["proto"],
		}

		// Parse source
		if src := section.Options["src"]; src != "" {
			rule.SourceZone = src
		}
		if srcIP := section.Options["src_ip"]; srcIP != "" {
			rule.Source = append(rule.Source, srcIP)
		}

		// Parse destination
		if dest := section.Options["dest"]; dest != "" {
			rule.DestZone = dest
		}
		if destIP := section.Options["dest_ip"]; destIP != "" {
			rule.Destination = append(rule.Destination, destIP)
		}

		// Parse ports
		if destPort := section.Options["dest_port"]; destPort != "" {
			rule.Ports = append(rule.Ports, destPort)
		}
		if srcPort := section.Options["src_port"]; srcPort != "" {
			rule.Ports = append(rule.Ports, srcPort)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// Helper functions

func determineInterfaceType(section UCISection) string {
	if section.Options["type"] == "bridge" {
		return "bridge"
	}
	if section.Options["type"] == "8021q" {
		return "vlan"
	}
	if section.Options["proto"] == "none" {
		return "physical"
	}
	return "logical"
}

func mapOpenWrtTarget(target string) string {
	switch strings.ToLower(target) {
	case "accept":
		return "allow"
	case "drop", "reject":
		return "deny"
	default:
		return target
	}
}

func extractHostname(config *UCIConfig) string {
	if system, exists := config.Sections["system"]["system"]; exists {
		return system.Options["hostname"]
	}
	return ""
}

func extractConfigVersion(config *UCIConfig) string {
	// Try to get version from various sources
	if system, exists := config.Sections["system"]["system"]; exists {
		if version := system.Options["version"]; version != "" {
			return version
		}
	}
	return time.Now().Format("2006-01-02")
}

func extractOpenWrtModel(config *UCIConfig, deviceInfo s.SNMPDevice) string {
	// Try to extract model from system config
	if system, exists := config.Sections["system"]["system"]; exists {
		if model := system.Options["model"]; model != "" {
			return model
		}
	}

	// Fallback to SNMP description
	if deviceInfo.SysDescr != "" {
		return deviceInfo.SysDescr
	}

	return "OpenWrt Device"
}

// parseWirelessInterfaces extracts wifi-device (radios) and wifi-iface (SSIDs) from UCI wireless config.
func (p *OpenWrtParser) parseWirelessInterfaces(config *UCIConfig) ([]configparser.ConfigInterface, error) {
	var ifaces []configparser.ConfigInterface

	wireless, exists := config.Sections["wireless"]
	if !exists {
		return ifaces, nil
	}

	for name, section := range wireless {
		switch section.Type {
		case "wifi-device":
			band := normalizeWifiBand(section.Options["band"])
			ifaces = append(ifaces, configparser.ConfigInterface{
				Name:     name,
				Type:     "wifi-radio",
				Enabled:  section.Options["disabled"] != "1",
				WifiBand: band,
			})

		case "wifi-iface":
			ssid := section.Options["ssid"]
			iface := configparser.ConfigInterface{
				Type:         "wifi-iface",
				Enabled:      section.Options["disabled"] != "1",
				WifiSSID:     ssid,
				WifiSecurity: normalizeWifiSecurity(section.Options["encryption"]),
				WifiRadio:    section.Options["device"],
				Parent:       section.Options["device"],
			}
			// Use SSID as name if available, otherwise fall back to section key
			if ssid != "" {
				iface.Name = ssid
			} else {
				iface.Name = name
			}
			ifaces = append(ifaces, iface)
		}
	}

	return ifaces, nil
}

// normalizeWifiBand converts UCI band values to human-readable strings.
func normalizeWifiBand(band string) string {
	switch strings.ToLower(band) {
	case "2g", "2ghz", "2.4g", "2.4ghz", "bg", "bgn":
		return "2.4GHz"
	case "5g", "5ghz", "a", "an", "ac":
		return "5GHz"
	case "6g", "6ghz", "ax6":
		return "6GHz"
	default:
		return band
	}
}

// normalizeWifiSecurity maps UCI encryption values to canonical security modes.
func normalizeWifiSecurity(enc string) string {
	enc = strings.ToLower(enc)
	switch {
	case enc == "" || enc == "none":
		return "open"
	case strings.HasPrefix(enc, "sae") || enc == "psk-mixed+ccmp" || enc == "psk2+ccmp+sae":
		return "wpa3"
	case strings.HasPrefix(enc, "psk2") || strings.HasPrefix(enc, "psk+ccmp") || strings.HasPrefix(enc, "ccmp"):
		return "wpa2"
	case strings.HasPrefix(enc, "psk"):
		return "wpa2"
	default:
		return "open"
	}
}
