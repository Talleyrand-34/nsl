package configparser

import (
	"fmt"
	"os"
	"strings"

	s "nsl-graph/internal/scanner"
)

// ConfigurationMerger coordinates SNMP data with configuration data
type ConfigurationMerger struct {
	parserRegistry *ConfigParserRegistry
}

// NewConfigurationMerger creates a new configuration merger
func NewConfigurationMerger() *ConfigurationMerger {
	return &ConfigurationMerger{
		parserRegistry: DefaultRegistry,
	}
}

// EnhanceDeviceWithConfig enhances a discovered device with configuration data
func (m *ConfigurationMerger) EnhanceDeviceWithConfig(
	device s.DiscoveredDevice,
	options ConfigParserOptions,
) (*EnhancedDiscoveredDevice, error) {
	enhanced := &EnhancedDiscoveredDevice{
		DiscoveredDevice: device,
		EnhancedBy:       []string{"snmp"},
	}

	// Skip configuration parsing if source is none
	if options.Source == ConfigSourceManual {
		return enhanced, nil
	}

	// Find appropriate parser for the device (manual type or auto-detection)
	parser, exists := m.parserRegistry.GetParserForDeviceWithType(device.Device, options.OsType)
	if !exists {
		if options.OsType != "" {
			return enhanced, fmt.Errorf(
				"specified OS type '%s' not available or device SNMP data doesn't match auto-detection",
				options.OsType,
			)
		}
		return enhanced, fmt.Errorf(
			"no configuration parser available for OS type: %s",
			device.Device.SysDescr,
		)
	}

	// Get configuration data
	var rawConfig string
	var err error

	switch options.Source {
	case ConfigSourceSSH:
		if options.SSHCredentials == nil {
			return enhanced, fmt.Errorf("SSH credentials required for SSH config source")
		}
		rawConfig, err = FetchConfig(DefaultTransport, parser, device.Device.IP, *options.SSHCredentials)
		if err != nil {
			return enhanced, fmt.Errorf("failed to retrieve config via SSH: %w", err)
		}

	case ConfigSourceFile:
		if options.FilePath == "" {
			return enhanced, fmt.Errorf("config file path required for file config source")
		}
		// Read config from file
		content, err := readConfigFile(options.FilePath)
		if err != nil {
			return enhanced, fmt.Errorf("failed to read config file: %w", err)
		}
		rawConfig = content

	default:
		return enhanced, fmt.Errorf("unsupported config source: %v", options.Source)
	}

	// Parse configuration
	configData, err := parser.ParseConfig(rawConfig, device.Device)
	if err != nil {
		return enhanced, fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Set the source that was actually used
	configData.Source = options.Source

	// Validate configuration
	if validationErrors := parser.ValidateConfig(configData); len(validationErrors) > 0 {
		// Log validation errors but don't fail
		for _, validationError := range validationErrors {
			fmt.Printf("Config validation warning: %v\n", validationError)
		}
	}

	enhanced.ConfigData = configData
	enhanced.EnhancedBy = append(enhanced.EnhancedBy, "config")

	// Merge SNMP and config data if requested
	if options.MergeWithSNMP {
		discrepancies, err := m.detectDiscrepancies(device.Device, configData)
		if err != nil {
			return enhanced, fmt.Errorf("failed to detect discrepancies: %w", err)
		}

		enhanced.Discrepancies = discrepancies

		// Handle discrepancies based on action
		if len(discrepancies) > 0 {
			switch options.DiscrepancyAction {
			case DiscrepancyActionFail:
				return enhanced, fmt.Errorf(
					"configuration discrepancies detected: %d conflicts found",
					len(discrepancies),
				)
			case DiscrepancyActionPreferSNMP:
				// Keep SNMP data as primary, use config as supplemental
			case DiscrepancyActionPreferConfig:
				// Use config data as primary, supplement with SNMP
				enhanced.DiscoveredDevice = m.mergeConfigIntoDevice(
					enhanced.DiscoveredDevice,
					configData,
				)
			}
		}
	}

	return enhanced, nil
}

// detectDiscrepancies compares SNMP data with configuration data
func (m *ConfigurationMerger) detectDiscrepancies(
	snmpDevice s.SNMPDevice,
	configData *ConfigData,
) ([]ConfigDiscrepancy, error) {
	var discrepancies []ConfigDiscrepancy

	// Check hostname discrepancy
	if configData.Hostname != "" && snmpDevice.SysName != "" {
		if configData.Hostname != snmpDevice.SysName {
			discrepancies = append(discrepancies, ConfigDiscrepancy{
				Type:        "hostname",
				Object:      "system",
				SNMPValue:   snmpDevice.SysName,
				ConfigValue: configData.Hostname,
				Description: "Device hostname differs between SNMP and configuration",
				Severity:    "warning",
			})
		}
	}

	// Check interface discrepancies
	for _, snmpIface := range snmpDevice.Interfaces {
		configIface := findConfigInterface(configData.Interfaces, snmpIface.Name)
		if configIface == nil {
			// SNMP interface not found in config
			discrepancies = append(discrepancies, ConfigDiscrepancy{
				Type:        "interface",
				Object:      snmpIface.Name,
				SNMPValue:   "exists",
				ConfigValue: "missing",
				Description: fmt.Sprintf(
					"Interface %s exists in SNMP but not in configuration",
					snmpIface.Name,
				),
				Severity: "warning",
			})
			continue
		}

		// Check interface status discrepancy
		snmpEnabled := snmpIface.OperStatus == 1
		if snmpEnabled != configIface.Enabled {
			discrepancies = append(discrepancies, ConfigDiscrepancy{
				Type:        "interface_status",
				Object:      snmpIface.Name,
				SNMPValue:   boolToString(snmpEnabled),
				ConfigValue: boolToString(configIface.Enabled),
				Description: fmt.Sprintf(
					"Interface %s status differs: SNMP shows %s, config shows %s",
					snmpIface.Name,
					statusToString(snmpEnabled),
					statusToString(configIface.Enabled),
				),
				Severity: "error",
			})
		}

		// Check IP address discrepancies
		snmpIPs := normalizeIPList(snmpIface.IPAddresses)
		configIPs := normalizeIPList(configIface.IPAddresses)
		if !equalStringSlices(snmpIPs, configIPs) {
			discrepancies = append(discrepancies, ConfigDiscrepancy{
				Type:        "ip_address",
				Object:      snmpIface.Name,
				SNMPValue:   strings.Join(snmpIPs, ","),
				ConfigValue: strings.Join(configIPs, ","),
				Description: fmt.Sprintf(
					"Interface %s IP addresses differ between SNMP and configuration",
					snmpIface.Name,
				),
				Severity: "warning",
			})
		}

		// Check VLAN discrepancies
		snmpVLANs := extractVLANNumbers(snmpIface.VLANs)
		configVLANs := extractConfigVLANNumbers(configIface.VLANs)
		if !equalStringSlices(snmpVLANs, configVLANs) {
			discrepancies = append(discrepancies, ConfigDiscrepancy{
				Type:        "vlan",
				Object:      snmpIface.Name,
				SNMPValue:   strings.Join(snmpVLANs, ","),
				ConfigValue: strings.Join(configVLANs, ","),
				Description: fmt.Sprintf(
					"Interface %s VLAN configuration differs between SNMP and configuration",
					snmpIface.Name,
				),
				Severity: "warning",
			})
		}
	}

	// Check for config interfaces not found in SNMP
	for _, configIface := range configData.Interfaces {
		snmpIface := findSNMPInterface(snmpDevice.Interfaces, configIface.Name)
		if snmpIface == nil {
			discrepancies = append(discrepancies, ConfigDiscrepancy{
				Type:        "interface",
				Object:      configIface.Name,
				SNMPValue:   "missing",
				ConfigValue: "exists",
				Description: fmt.Sprintf(
					"Interface %s exists in configuration but not in SNMP",
					configIface.Name,
				),
				Severity: "warning",
			})
		}
	}

	return discrepancies, nil
}

// mergeConfigIntoDevice merges configuration data into the discovered device
func (m *ConfigurationMerger) mergeConfigIntoDevice(
	device s.DiscoveredDevice,
	configData *ConfigData,
) s.DiscoveredDevice {
	merged := device

	// Use config hostname if available and different
	if configData.Hostname != "" && configData.Hostname != device.Device.SysName {
		merged.SuggestedName = configData.Hostname
	}

	// Enhance device interfaces with config data
	for i, snmpIface := range merged.Device.Interfaces {
		configIface := findConfigInterface(configData.Interfaces, snmpIface.Name)
		if configIface != nil {
			// Add description from config if missing in SNMP
			if snmpIface.Name == configIface.Name && configIface.Description != "" {
				// Note: We can't directly modify the interface in the slice
				// This would require restructuring the data types to allow enhancement
			}

			// Merge VLAN information from config
			for _, configVLAN := range configIface.VLANs {
				// Check if this VLAN already exists in SNMP data
				exists := false
				for _, snmpVLAN := range snmpIface.VLANs {
					if snmpVLAN.VLANNumber == configVLAN.ID {
						exists = true
						break
					}
				}

				// Add VLAN from config if not found in SNMP
				if !exists {
					merged.Device.Interfaces[i].VLANs = append(
						merged.Device.Interfaces[i].VLANs,
						s.VLANMembership{
							VLANNumber: configVLAN.ID,
							Tagged:     configVLAN.Tagged,
						},
					)
				}
			}

			// Propagate parent interface name from config if not already set.
			if merged.Device.Interfaces[i].Parent == "" && configIface.Parent != "" {
				merged.Device.Interfaces[i].Parent = configIface.Parent
			}

			// Propagate IsBridge from config if the config says this is a bridge interface.
			if configIface.Type == "bridge" {
				merged.Device.Interfaces[i].IsBridge = true
			}
		}
	}

	return merged
}

// Helper functions

func readConfigFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file %q: %w", filePath, err)
	}
	return string(data), nil
}

func findConfigInterface(interfaces []ConfigInterface, name string) *ConfigInterface {
	for _, iface := range interfaces {
		if iface.Name == name {
			return &iface
		}
	}
	return nil
}

func findSNMPInterface(interfaces []s.DeviceInterface, name string) *s.DeviceInterface {
	for _, iface := range interfaces {
		if iface.Name == name {
			return &iface
		}
	}
	return nil
}

func normalizeIPList(ips []string) []string {
	var normalized []string
	for _, ip := range ips {
		// Remove CIDR notation for comparison
		if strings.Contains(ip, "/") {
			ip = strings.Split(ip, "/")[0]
		}
		normalized = append(normalized, ip)
	}
	return normalized
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func extractVLANNumbers(vlans []s.VLANMembership) []string {
	var numbers []string
	for _, vlan := range vlans {
		numbers = append(numbers, vlan.VLANNumber)
	}
	return numbers
}

func extractConfigVLANNumbers(vlans []ConfigVLAN) []string {
	var numbers []string
	for _, vlan := range vlans {
		numbers = append(numbers, vlan.ID)
	}
	return numbers
}

func boolToString(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

func statusToString(enabled bool) string {
	if enabled {
		return "up"
	}
	return "down"
}

// GetAvailableParsers returns a list of available configuration parsers
func (m *ConfigurationMerger) GetAvailableParsers() []string {
	return m.parserRegistry.ListParsers()
}

// GetParserForDevice returns the appropriate parser for a device
func (m *ConfigurationMerger) GetParserForDevice(device s.SNMPDevice) (ConfigParser, bool) {
	return m.parserRegistry.GetParserForDevice(device)
}

// TestConfigConnection tests configuration retrieval without full parsing
func (m *ConfigurationMerger) TestConfigConnection(
	device s.SNMPDevice,
	options ConfigParserOptions,
) error {
	_, exists := m.parserRegistry.GetParserForDevice(device)
	if !exists {
		return fmt.Errorf("no configuration parser available for device")
	}

	switch options.Source {
	case ConfigSourceSSH:
		if options.SSHCredentials == nil {
			return fmt.Errorf("SSH credentials required")
		}
		return ValidateHost(device.IP, *options.SSHCredentials)
	case ConfigSourceFile:
		if options.FilePath == "" {
			return fmt.Errorf("config file path required")
		}
		// Check if file exists and is readable
		_, err := readConfigFile(options.FilePath)
		return err
	default:
		return fmt.Errorf("unsupported config source: %v", options.Source)
	}
}

// Import functions from parsers package
// Note: These would typically be imported, but since we can't use relative imports in Go,
// we'll need to implement them as proper imports when integrating
