package cmd_compare

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd"
	util "nsl-graph/cmd/utils"
	e "nsl-graph/internal/repository/entities"
)

type ComparisonOptions struct {
	Host       string
	SNMPDBPath string
	SSHDbPath  string
	OutputText string
	OutputJSON string
}

type Discrepancy struct {
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Object      string `json:"object"`
	Property    string `json:"property,omitempty"`
	SNMPValue   string `json:"snmp_value,omitempty"`
	SSHValue    string `json:"ssh_value,omitempty"`
	Description string `json:"description"`
}

type ComparisonReport struct {
	Host          string        `json:"host"`
	MethodSummary MethodSummary `json:"method_summary"`
	Discrepancies []Discrepancy `json:"discrepancies"`
	Summary       Summary       `json:"summary"`
}

type MethodSummary struct {
	SNMP MethodInfo `json:"snmp"`
	SSH  MethodInfo `json:"ssh"`
}

type MethodInfo struct {
	InterfaceCount int      `json:"interface_count"`
	PhysicalPorts  int      `json:"physical_ports"`
	LogicalPorts   int      `json:"logical_ports"`
	VLANCount      int      `json:"vlan_count"`
	IPCount        int      `json:"ip_count"`
	InterfaceNames []string `json:"interface_names"`
}

type Summary struct {
	SSHAdvantages      int `json:"ssh_advantages"`
	PotentialBugSSH    int `json:"potential_bug_ssh"`
	PotentialBugSNMP   int `json:"potential_bug_snmp"`
	ValueMismatches    int `json:"value_mismatches"`
	SNMPLimitations    int `json:"snmp_limitations"`
	TotalDiscrepancies int `json:"total_discrepancies"`
}

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare SNMP vs SSH scan results",
	Long: `Compare device scan results from two databases - one scanned via SNMP and one via SSH.
Outputs a text report and JSON report detailing discrepancies between the two methods.`,
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		snmpDB, _ := cmd.Flags().GetString("snmp-db")
		sshDB, _ := cmd.Flags().GetString("ssh-db")
		outputText, _ := cmd.Flags().GetString("output-text")
		outputJSON, _ := cmd.Flags().GetString("output-json")

		if host == "" || snmpDB == "" || sshDB == "" {
			fmt.Println("Error: --host, --snmp-db, and --ssh-db are required")
			os.Exit(1)
		}

		opts := ComparisonOptions{
			Host:       host,
			SNMPDBPath: snmpDB,
			SSHDbPath:  sshDB,
			OutputText: outputText,
			OutputJSON: outputJSON,
		}

		if err := runComparison(opts); err != nil {
			fmt.Printf("Error running comparison: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(compareCmd)
	compareCmd.Flags().String("host", "", "Host IP to compare")
	compareCmd.Flags().String("snmp-db", "", "Path to SNMP database")
	compareCmd.Flags().String("ssh-db", "", "Path to SSH database")
	compareCmd.Flags().String("output-text", "", "Output text report to file (optional)")
	compareCmd.Flags().String("output-json", "", "Output JSON report to file (optional)")
}

func runComparison(opts ComparisonOptions) error {
	snmpData, err := extractData(opts.SNMPDBPath, opts.Host)
	if err != nil {
		return fmt.Errorf("failed to extract SNMP data: %w", err)
	}

	sshData, err := extractData(opts.SSHDbPath, opts.Host)
	if err != nil {
		return fmt.Errorf("failed to extract SSH data: %w", err)
	}

	report := compareData(opts.Host, snmpData, sshData)

	textOutput := formatTextReport(report)
	fmt.Println(textOutput)

	if opts.OutputText != "" {
		if err := os.WriteFile(opts.OutputText, []byte(textOutput), 0644); err != nil {
			return fmt.Errorf("failed to write text report: %w", err)
		}
		fmt.Printf("\nText report saved to: %s\n", opts.OutputText)
	}

	jsonOutput, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if opts.OutputJSON != "" {
		if err := os.WriteFile(opts.OutputJSON, jsonOutput, 0644); err != nil {
			return fmt.Errorf("failed to write JSON report: %w", err)
		}
		fmt.Printf("JSON report saved to: %s\n", opts.OutputJSON)
	}

	return nil
}

type ExtractedData struct {
	DeviceName  string
	DeviceModel string
	DeviceBrand string
	Interfaces  []InterfaceData
	DevicePorts []DevicePortData
	VLANs       []VLANData
	LocalVlans  []LocalVlanData
}

type InterfaceData struct {
	Name        string
	Description string
	Type        string
	MAC         string
	Status      string
	IPs         []string
	VLANs       []VLANInfo
	Parent      string
	WifiSSID    string
}

type VLANInfo struct {
	ID     string
	Tagged bool
}

type DevicePortData struct {
	PortName    string
	MAC         string
	VlanConfigs []VLANInfo
}

type VLANData struct {
	VlanID string
	Name   string
}

type LocalVlanData struct {
	VlanID   string
	VlanName string
	DeviceID string
}

func extractData(dbPath, hostIP string) (*ExtractedData, error) {
	originalPath := cmd.Srcdbpath
	cmd.Srcdbpath = dbPath
	defer func() { cmd.Srcdbpath = originalPath }()

	service, err := util.ServiceConnection()
	if err != nil {
		return nil, err
	}

	data := &ExtractedData{}

	devices, err := service.GetDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	var targetDevice e.Device
	for _, d := range devices {
		for _, ip := range d.IPs {
			if ip == hostIP {
				targetDevice = d
				break
			}
		}
		if targetDevice.ID != "" {
			break
		}
	}

	if targetDevice.ID == "" {
		return nil, fmt.Errorf("device with IP %s not found in database", hostIP)
	}

	data.DeviceName = targetDevice.Name
	data.DeviceModel = targetDevice.Model
	data.DeviceBrand = targetDevice.Brand

	deviceID := targetDevice.ID

	interfaces, err := service.GetAllDeviceInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get device interfaces: %w", err)
	}

	for _, iface := range interfaces {
		if iface.DeviceID != deviceID {
			continue
		}

		ifaceData := InterfaceData{
			Name:        iface.Name,
			Description: iface.Description,
			WifiSSID:    iface.WifiSSID,
			Status:      "unknown",
		}

		for _, v := range iface.VlanConfigs {
			ifaceData.VLANs = append(ifaceData.VLANs, VLANInfo{
				ID:     v.VlanNumber,
				Tagged: v.Tagged,
			})
		}

		for _, ip := range iface.IPAddresses {
			ifaceData.IPs = append(ifaceData.IPs, ip)
		}

		data.Interfaces = append(data.Interfaces, ifaceData)
	}

	ports, err := service.GetAllPortsDevice(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device ports: %w", err)
	}

	for _, p := range ports {
		portData := DevicePortData{
			PortName: p.PortName,
			MAC:      p.MacAddress,
		}
		for _, v := range p.VlanConfigs {
			portData.VlanConfigs = append(portData.VlanConfigs, VLANInfo{
				ID:     v.VlanNumber,
				Tagged: v.Tagged,
			})
		}
		data.DevicePorts = append(data.DevicePorts, portData)
	}

	return data, nil
}

func compareData(host string, snmp, ssh *ExtractedData) *ComparisonReport {
	report := &ComparisonReport{
		Host: host,
		MethodSummary: MethodSummary{
			SNMP: extractMethodInfo(snmp),
			SSH:  extractMethodInfo(ssh),
		},
		Discrepancies: []Discrepancy{},
	}

	snmpIfaces := make(map[string]*InterfaceData)
	for i := range snmp.Interfaces {
		snmpIfaces[snmp.Interfaces[i].Name] = &snmp.Interfaces[i]
	}

	sshIfaces := make(map[string]*InterfaceData)
	for i := range ssh.Interfaces {
		sshIfaces[ssh.Interfaces[i].Name] = &ssh.Interfaces[i]
	}

	for name, snmpIface := range snmpIfaces {
		_, exists := sshIfaces[name]
		if !exists {
			report.Discrepancies = append(report.Discrepancies, Discrepancy{
				Category:    "SSH_PARSER_BUG",
				Severity:    "warning",
				Object:      name,
				Description: fmt.Sprintf("Interface '%s' exists in SNMP but NOT in SSH config (SSH parser may not be extracting it)", name),
			})
			continue
		}

		compareInterfaces(report, name, snmpIface, sshIfaces[name])
	}

	for name := range sshIfaces {
		_, exists := snmpIfaces[name]
		if !exists {
			report.Discrepancies = append(report.Discrepancies, Discrepancy{
				Category:    "SNMP_LIMITATION",
				Severity:    "info",
				Object:      name,
				Description: fmt.Sprintf("Interface '%s' exists in SSH but NOT in SNMP (SNMP cannot detect this type of interface)", name),
			})
		}
	}

	sort.Slice(report.Discrepancies, func(i, j int) bool {
		if report.Discrepancies[i].Category != report.Discrepancies[j].Category {
			return report.Discrepancies[i].Category < report.Discrepancies[j].Category
		}
		return report.Discrepancies[i].Object < report.Discrepancies[j].Object
	})

	report.Summary = calculateSummary(report.Discrepancies)

	return report
}

func extractMethodInfo(data *ExtractedData) MethodInfo {
	info := MethodInfo{
		InterfaceNames: []string{},
	}

	physicalPorts := make(map[string]bool)
	logicalPorts := make(map[string]bool)

	for _, iface := range data.Interfaces {
		info.InterfaceNames = append(info.InterfaceNames, iface.Name)

		if iface.Type == "ethernet" || iface.Type == "wifi" {
			physicalPorts[iface.Name] = true
		} else {
			logicalPorts[iface.Name] = true
		}

		info.IPCount += len(iface.IPs)
		info.VLANCount += len(iface.VLANs)
	}

	info.InterfaceCount = len(data.Interfaces)
	info.PhysicalPorts = len(physicalPorts)
	info.LogicalPorts = len(logicalPorts)

	sort.Strings(info.InterfaceNames)
	return info
}

func compareInterfaces(report *ComparisonReport, name string, snmpIface, sshIface *InterfaceData) {
	if len(snmpIface.IPs) != len(sshIface.IPs) {
		report.Discrepancies = append(report.Discrepancies, Discrepancy{
			Category:    "VALUE_MISMATCH",
			Severity:    "warning",
			Object:      name,
			Property:    "ip_count",
			SNMPValue:   fmt.Sprintf("%d", len(snmpIface.IPs)),
			SSHValue:    fmt.Sprintf("%d", len(sshIface.IPs)),
			Description: fmt.Sprintf("Interface '%s' IP count differs: SNMP=%d, SSH=%d", name, len(snmpIface.IPs), len(sshIface.IPs)),
		})
	}

	snmpIPs := make(map[string]bool)
	for _, ip := range snmpIface.IPs {
		snmpIPs[normalizeIP(ip)] = true
	}
	for _, ip := range sshIface.IPs {
		normalized := normalizeIP(ip)
		if !snmpIPs[normalized] {
			report.Discrepancies = append(report.Discrepancies, Discrepancy{
				Category:    "SNMP_LIMITATION",
				Severity:    "info",
				Object:      name,
				Property:    "ip",
				SSHValue:    ip,
				Description: fmt.Sprintf("Interface '%s' has IP '%s' in SSH but not in SNMP (SNMP limitation)", name, ip),
			})
		}
	}
	for _, ip := range snmpIface.IPs {
		normalized := normalizeIP(ip)
		sshIPs := make(map[string]bool)
		for _, sip := range sshIface.IPs {
			sshIPs[normalizeIP(sip)] = true
		}
		if !sshIPs[normalized] {
			report.Discrepancies = append(report.Discrepancies, Discrepancy{
				Category:    "SSH_PARSER_BUG",
				Severity:    "info",
				Object:      name,
				Property:    "ip",
				SNMPValue:   ip,
				Description: fmt.Sprintf("Interface '%s' has IP '%s' in SNMP but not in SSH (SSH parser may not be extracting it)", name, ip),
			})
		}
	}

	snmpVLANs := make(map[string]bool)
	for _, v := range snmpIface.VLANs {
		snmpVLANs[v.ID] = true
	}
	sshVLANs := make(map[string]bool)
	for _, v := range sshIface.VLANs {
		sshVLANs[v.ID] = true
	}

	for vlanID := range snmpVLANs {
		if _, exists := sshVLANs[vlanID]; !exists {
			report.Discrepancies = append(report.Discrepancies, Discrepancy{
				Category:    "SSH_PARSER_BUG",
				Severity:    "info",
				Object:      name,
				Property:    "vlan",
				SNMPValue:   vlanID,
				Description: fmt.Sprintf("Interface '%s' has VLAN '%s' in SNMP but not in SSH config (SSH parser may not be extracting it)", name, vlanID),
			})
		}
	}

	for vlanID := range sshVLANs {
		if _, exists := snmpVLANs[vlanID]; !exists {
			report.Discrepancies = append(report.Discrepancies, Discrepancy{
				Category:    "SNMP_LIMITATION",
				Severity:    "info",
				Object:      name,
				Property:    "vlan",
				SSHValue:    vlanID,
				Description: fmt.Sprintf("Interface '%s' has VLAN '%s' in SSH config but not in SNMP (SNMP cannot detect this VLAN)", name, vlanID),
			})
		}
	}

	if snmpIface.WifiSSID != "" && sshIface.WifiSSID == "" {
		report.Discrepancies = append(report.Discrepancies, Discrepancy{
			Category:    "SNMP_LIMITATION",
			Severity:    "info",
			Object:      name,
			Property:    "wifi_ssid",
			SNMPValue:   snmpIface.WifiSSID,
			Description: fmt.Sprintf("Interface '%s' has WiFi SSID '%s' in SNMP but SSH parser may not extract it", name, snmpIface.WifiSSID),
		})
	} else if sshIface.WifiSSID != "" && snmpIface.WifiSSID == "" {
		report.Discrepancies = append(report.Discrepancies, Discrepancy{
			Category:    "SSH_ADVANTAGE",
			Severity:    "info",
			Object:      name,
			Property:    "wifi_ssid",
			SSHValue:    sshIface.WifiSSID,
			Description: fmt.Sprintf("Interface '%s' has WiFi SSID '%s' in SSH but not in SNMP (SNMP limitation)", name, sshIface.WifiSSID),
		})
	}
}

func normalizeIP(ip string) string {
	if strings.Contains(ip, "/") {
		return strings.Split(ip, "/")[0]
	}
	return ip
}

func calculateSummary(discrepancies []Discrepancy) Summary {
	summary := Summary{TotalDiscrepancies: len(discrepancies)}

	for _, d := range discrepancies {
		switch d.Category {
		case "SSH_ADVANTAGE":
			summary.SSHAdvantages++
		case "SSH_PARSER_BUG":
			if d.Severity == "warning" {
				summary.PotentialBugSSH++
			} else {
				summary.PotentialBugSSH++
			}
		case "SNMP_LIMITATION":
			summary.SNMPLimitations++
		case "VALUE_MISMATCH":
			summary.ValueMismatches++
		}
	}

	return summary
}

func formatTextReport(report *ComparisonReport) string {
	var sb strings.Builder

	sb.WriteString("=== SNMP vs SSH Comparison Report ===\n")
	sb.WriteString(fmt.Sprintf("Host: %s\n\n", report.Host))

	sb.WriteString("--- Method Summary ---\n")
	sb.WriteString(fmt.Sprintf("SNMP: %d interfaces (%d physical, %d logical), %d IPs, %d VLANs\n",
		report.MethodSummary.SNMP.InterfaceCount,
		report.MethodSummary.SNMP.PhysicalPorts,
		report.MethodSummary.SNMP.LogicalPorts,
		report.MethodSummary.SNMP.IPCount,
		report.MethodSummary.SNMP.VLANCount))
	sb.WriteString(fmt.Sprintf("SSH:  %d interfaces (%d physical, %d logical), %d IPs, %d VLANs\n\n",
		report.MethodSummary.SSH.InterfaceCount,
		report.MethodSummary.SSH.PhysicalPorts,
		report.MethodSummary.SSH.LogicalPorts,
		report.MethodSummary.SSH.IPCount,
		report.MethodSummary.SSH.VLANCount))

	sb.WriteString("--- Discrepancies ---\n")
	if len(report.Discrepancies) == 0 {
		sb.WriteString("No discrepancies found.\n")
	} else {
		for _, d := range report.Discrepancies {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", d.Category, d.Object, d.Description))
			if d.SNMPValue != "" || d.SSHValue != "" {
				sb.WriteString(fmt.Sprintf("    SNMP: %s | SSH: %s\n", d.SNMPValue, d.SSHValue))
			}
		}
	}

	sb.WriteString("\n--- Summary ---\n")
	sb.WriteString(fmt.Sprintf("SSH Advantages (data SSH can get that SNMP can't): %d\n", report.Summary.SSHAdvantages))
	sb.WriteString(fmt.Sprintf("SNMP Limitations (inherent SNMP limitations): %d\n", report.Summary.SNMPLimitations))
	sb.WriteString(fmt.Sprintf("Potential SSH Parser Bugs (SSH should have but doesn't): %d\n", report.Summary.PotentialBugSSH))
	sb.WriteString(fmt.Sprintf("Value Mismatches (same data, different values): %d\n", report.Summary.ValueMismatches))
	sb.WriteString(fmt.Sprintf("Total Discrepancies: %d\n", report.Summary.TotalDiscrepancies))

	return sb.String()
}
