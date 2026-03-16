/*
Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package cmd_scan

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
	q "nsl-graph/internal/repository/application"
	s "nsl-graph/internal/scanner"
)

var (
	hostTimeout      int
	hostCommunity    string
	hostSNMPVersion  string
	hostSNMPPort     uint16
	hostAutoImport   bool
	hostDefaultZone  string
	hostDefaultBrand string
	hostOutputFile   string
	hostApproveAll   bool
	hostVLANAccuracy int
	hostScanSource   string // "snmp" or "ssh"

	// Configuration parsing options
	hostConfigSource      string
	hostConfigFile        string
	hostDeviceType        string
	hostSSHUsername       string
	hostSSHPassword       string
	hostSSHKeyFile        string
	hostSSHPort           int
	hostDiscrepancyAction string
	hostMergeConfig       bool
	hostConfigTimeout     int
)

var hostScanCmd = &cobra.Command{
	Use:   "host [ip_or_ssh_alias]",
	Short: "Query a specific host via SNMP to collect inventory",
	Long: `Query a specific device to collect interfaces, MAC addresses, and VLAN assignments.

The argument can be an IP address or an SSH config alias from ~/.ssh/config.
When an alias is given, HostName/User/IdentityFile/Port are resolved automatically;
explicit flags (--ssh-user, --ssh-key, --ssh-port) always take precedence.

Use --scan-source to choose the primary protocol:
  snmp (default): query via SNMP; optionally augment with SSH/file config
  ssh:            skip SNMP entirely; use SSH configuration as the primary source

Examples:
  # SNMP scanning (default)
  nsl-graph scan host 192.168.1.1
  nsl-graph scan host 10.0.0.1 --snmp-community private
  nsl-graph scan host 192.168.1.1 --output device_scan.json
  nsl-graph scan host 172.16.1.10 --auto-import

  # SSH-only scanning via IP
  nsl-graph scan host 10.0.0.245 --scan-source ssh --ssh-user root --auto-import
  nsl-graph scan host 10.0.0.1   --scan-source ssh --ssh-user admin --device-type opnsense --auto-import

  # SSH-only scanning via ~/.ssh/config alias (resolves host/user/key automatically)
  nsl-graph scan host opnsense --scan-source ssh --auto-import
  nsl-graph scan host opnsense --auto-import          # scan-source defaults to ssh for aliases
  nsl-graph scan host opnsense --ssh-user admin       # override a single field

  # SNMP + SSH configuration augmentation
  nsl-graph scan host 10.0.0.1 --config-source ssh --ssh-user admin --merge-configs
  nsl-graph scan host 10.0.0.132 --config-source ssh --ssh-user admin --discrepancy-action prefer-config
  nsl-graph scan host 10.0.0.245 --config-source file --config-file openwrt.conf --auto-import`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ip := args[0]

		// If the argument is not a valid IP address, treat it as an ~/.ssh/config alias.
		if net.ParseIP(ip) == nil {
			entry, err := parseSSHConfigAlias(ip)
			if err != nil {
				fmt.Printf("Error: %q is not a valid IP address and could not be resolved from ~/.ssh/config: %v\n", ip, err)
				os.Exit(1)
			}
			if entry.HostName != "" {
				fmt.Printf("Resolved SSH alias %q → %s\n", ip, entry.HostName)
				ip = entry.HostName
			}
			// Apply resolved values only when the flag was not set explicitly
			if !cmd.Flags().Changed("ssh-user") && entry.User != "" {
				hostSSHUsername = entry.User
			}
			if !cmd.Flags().Changed("ssh-key") && entry.IdentityFile != "" {
				hostSSHKeyFile = entry.IdentityFile
			}
			if !cmd.Flags().Changed("ssh-port") && entry.Port != 0 {
				hostSSHPort = entry.Port
			}
			// Default scan-source to ssh when an alias is used
			if !cmd.Flags().Changed("scan-source") {
				hostScanSource = "ssh"
			}
		}

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		var device *s.SNMPDevice

		switch hostScanSource {
		case "ssh":
			// SSH-only mode: skip SNMP, use SSH config as primary source
			if hostSSHUsername == "" {
				fmt.Println("Error: --ssh-user is required when using --scan-source ssh")
				os.Exit(1)
			}
			// Implicitly enable SSH config source if not already set
			if hostConfigSource == "none" {
				hostConfigSource = "ssh"
			}
			fmt.Printf("Scanning %s via SSH (user=%s)...\n", ip, hostSSHUsername)
			// Create a minimal device record; SSH config will populate port/interface
			// details during import via ExecuteApprovedImportPlan
			device = &s.SNMPDevice{
				IP:        ip,
				Reachable: true,
				SysName:   ip,
			}

		default: // "snmp"
			options := s.ScanOptions{
				Timeout: time.Duration(hostTimeout) * time.Second,
				SNMP: s.SNMPOptions{
					Community: hostCommunity,
					Version:   hostSNMPVersion,
					Port:      hostSNMPPort,
				},
			}

			fmt.Printf("Querying %s via SNMP (community=%s)...\n", ip, hostCommunity)

			var err error
			device, err = service.ScanDevice(ip, options)
			if err != nil {
				fmt.Printf("Device scan failed: %v\n", err)
				os.Exit(1)
			}

			if !device.Reachable {
				fmt.Printf("Device %s did not respond to SNMP\n", ip)
				return
			}

			fmt.Printf("\nSNMP Results for %s:\n", ip)
			if device.SysName != "" {
				fmt.Printf("Name:     %s\n", device.SysName)
			}
			if device.SysDescr != "" {
				descr := device.SysDescr
				if len(descr) > 80 {
					descr = descr[:80] + "..."
				}
				fmt.Printf("Descr:    %s\n", descr)
			}
			if device.SysLocation != "" {
				fmt.Printf("Location: %s\n", device.SysLocation)
			}
			if device.SysContact != "" {
				fmt.Printf("Contact:  %s\n", device.SysContact)
			}

			if len(device.Interfaces) > 0 {
				fmt.Printf("\nInterfaces (%d):\n", len(device.Interfaces))
				fmt.Printf(
					"  %-30s %-17s %-6s %-20s %-8s %-8s\n",
					"Interface", "MAC", "Status", "IPs", "VLAN", "Accuracy",
				)
				fmt.Printf(
					"  %-30s %-17s %-6s %-20s %-8s %-8s\n",
					strings.Repeat("-", 30),
					strings.Repeat("-", 17),
					strings.Repeat("-", 6),
					strings.Repeat("-", 20),
					strings.Repeat("-", 8),
					strings.Repeat("-", 8),
				)

				for _, iface := range device.Interfaces {
					status := "down"
					if iface.OperStatus == 1 {
						status = "up"
					}

					vlanInference := s.InferVLANFromInterface(iface.Name, iface.IPAddresses)

					ipsDisplay := "[]"
					if len(iface.IPAddresses) > 0 {
						ipsDisplay = fmt.Sprintf("%v", iface.IPAddresses)
						if len(ipsDisplay) > 18 {
							ipsDisplay = ipsDisplay[:15] + "..."
						}
					}

					fmt.Printf("  %-30s %-17s %-6s %-20s %-8s %-8s\n",
						iface.Name, iface.MAC, status, ipsDisplay,
						vlanInference.FormatVLANDisplay(),
						vlanInference.FormatAccuracyDisplay())

					if len(iface.VLANs) > 0 {
						fmt.Printf("  %30s   SNMP VLANs:", "")
						for _, v := range iface.VLANs {
							t := "untagged"
							if v.Tagged {
								t = "tagged"
							}
							fmt.Printf(" %s(%s)", v.VLANNumber, t)
						}
						fmt.Println()
					}

					if vlanInference.Notes != "" {
						fmt.Printf("  %30s   Note: %s\n", "", vlanInference.Notes)
					}
				}
			}

			if len(device.Neighbors) > 0 {
				fmt.Printf("\nNeighbors (%d):\n", len(device.Neighbors))
				for _, n := range device.Neighbors {
					fmt.Printf(
						"  [%s] local:%s → remote:%s (%s)",
						n.Protocol, n.LocalPort, n.RemoteName, n.RemotePort,
					)
					if n.RemoteIP != "" {
						fmt.Printf(" IP:%s", n.RemoteIP)
					}
					fmt.Println()
				}
			}
		}

		// Save to JSON file if requested
		if hostOutputFile != "" {
			scanResult := &s.ScanResult{
				ID:        fmt.Sprintf("host_%s_%d", ip, time.Now().Unix()),
				Subnet:    ip + "/32", // Single host as /32
				StartTime: time.Now(),
				EndTime:   time.Now(),
				Devices:   []s.SNMPDevice{*device},
			}

			if err := saveHostScanResults(scanResult, hostOutputFile); err != nil {
				fmt.Printf("Failed to save results: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Results saved to %s\n", hostOutputFile)
		}

		if hostAutoImport {
			fmt.Println("\nImporting device...")

			// Create a ScanResult to match the file-based import workflow
			scanResult := &s.ScanResult{
				ID:        fmt.Sprintf("host_%s_%d", ip, time.Now().Unix()),
				Subnet:    ip + "/32", // Single host as /32
				StartTime: time.Now(),
				EndTime:   time.Now(),
				Devices:   []s.SNMPDevice{*device},
			}

			// Use the same discovery pipeline as file-based import
			devices, err := service.DiscoverDevices(scanResult)
			if err != nil {
				fmt.Printf("Device discovery failed: %v\n", err)
				os.Exit(1)
			}

			if len(devices) == 0 {
				fmt.Println("No devices discovered from scan result.")
				return
			}

			// Get the first (and only) discovered device
			discoveredDevice := devices[0]

			// Override zone if specified
			if hostDefaultZone != "" {
				discoveredDevice.SuggestedZone = hostDefaultZone
			}

			fmt.Printf(
				"Classification: %s %s [%s]\n",
				discoveredDevice.Brand,
				discoveredDevice.Model,
				discoveredDevice.DeviceClass,
			)

			importOptions := s.ImportOptions{
				AutoImport:        true,
				CreateZones:       true,
				DefaultZone:       hostDefaultZone,
				DefaultBrand:      hostDefaultBrand,
				SkipExisting:      true,
				InteractiveVLANs:  true, // Enable interactive VLAN mapping
				VLANAccuracyLevel: hostVLANAccuracy,

				// Configuration parsing options
				ConfigSource:       hostConfigSource,
				ConfigFile:         hostConfigFile,
				DeviceType:         hostDeviceType,
				SSHUsername:        hostSSHUsername,
				SSHPassword:        hostSSHPassword,
				SSHKeyFile:         hostSSHKeyFile,
				SSHPort:            hostSSHPort,
				DiscrepancyAction:  hostDiscrepancyAction,
				MergeWithConfig:    hostMergeConfig,
				ParseConfigTimeout: hostConfigTimeout,
			}

			if err := importSingleDeviceWithVLANMapping(service, discoveredDevice, importOptions); err != nil {
				fmt.Printf("Failed to import device: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Device '%s' imported successfully!\n", discoveredDevice.SuggestedName)
		} else {
			fmt.Println("\nUse --auto-import to add this device to the database")
		}
	},
}

func saveHostScanResults(result *s.ScanResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// importSingleDeviceWithVLANMapping imports a device with interactive VLAN mapping
func importSingleDeviceWithVLANMapping(
	service q.NetServiceInt,
	device s.DiscoveredDevice,
	options s.ImportOptions,
) error {
	// Analyze device for import plan
	plan, err := service.AnalyzeDeviceForImport(device)
	if err != nil {
		return fmt.Errorf("failed to analyze device: %w", err)
	}

	fmt.Printf("Analysis: %s\n", plan.Summary)

	// Show interface details with IP-VLAN mappings if any
	for i, interfacePlan := range plan.InterfacePlans {
		iface := interfacePlan.Interface
		fmt.Printf("  Interface %d: %s", i+1, iface.Name)

		if len(iface.IPAddresses) > 0 {
			fmt.Printf(" (IPs: %v)", iface.IPAddresses)
		}

		if len(interfacePlan.IPMappings) > 0 {
			fmt.Printf("\n    Proposed VLAN mappings:")
			for _, mapping := range interfacePlan.IPMappings {
				fmt.Printf(
					"\n      %s → VLAN %s (%s)",
					mapping.IP,
					mapping.VLANNumber,
					mapping.Confidence,
				)
			}
		}
		fmt.Println()
	}

	if plan.RequiresInput {
		if hostApproveAll {
			fmt.Printf("Auto-approving VLAN mapping (--approve-all flag)\n")
		} else {
			fmt.Printf("\nApprove this VLAN mapping? (y/n): ")
			var response string
			fmt.Scanln(&response)

			if strings.ToLower(strings.TrimSpace(response)) != "y" && strings.ToLower(strings.TrimSpace(response)) != "yes" {
				fmt.Println("Import cancelled by user.")
				return nil
			}
		}
	}

	// Execute the approved plan
	if err := service.ExecuteApprovedImportPlan(plan, options); err != nil {
		return fmt.Errorf("failed to execute import plan: %w", err)
	}

	return nil
}

func init() {
	cmd_root.ScanCmd.AddCommand(hostScanCmd)

	hostScanCmd.Flags().
		StringVar(&hostScanSource, "scan-source", "snmp", "Primary scan protocol: snmp (default) or ssh (skip SNMP, use SSH config as primary source)")
	hostScanCmd.Flags().
		IntVarP(&hostTimeout, "timeout", "t", 10, "Timeout in seconds for the SNMP query")
	hostScanCmd.Flags().
		StringVar(&hostCommunity, "snmp-community", "public", "SNMP community string")
	hostScanCmd.Flags().StringVar(&hostSNMPVersion, "snmp-version", "v2c", "SNMP version (v1, v2c)")
	hostScanCmd.Flags().Uint16Var(&hostSNMPPort, "snmp-port", 161, "SNMP UDP port")

	hostScanCmd.Flags().
		StringVarP(&hostOutputFile, "output", "o", "", "Save scan results to JSON file")

	hostScanCmd.Flags().
		BoolVar(&hostAutoImport, "auto-import", false, "Automatically import discovered device")
	hostScanCmd.Flags().
		StringVar(&hostDefaultZone, "default-zone", "", "Default zone for the device (defaults to Generic)")
	hostScanCmd.Flags().
		StringVar(&hostDefaultBrand, "default-brand", "", "Default brand for unidentified device")
	hostScanCmd.Flags().
		BoolVar(&hostApproveAll, "approve-all", false, "Automatically approve all VLAN mappings without prompting (only with --auto-import)")
	hostScanCmd.Flags().
		IntVar(&hostVLANAccuracy, "vlan-accuracy", 1, "VLAN detection accuracy level (1=interface names only, 2=include IP heuristics)")

	// Configuration parsing flags
	hostScanCmd.Flags().
		StringVar(&hostConfigSource, "config-source", "none", "Configuration source (none, ssh, file, manual)")
	hostScanCmd.Flags().
		StringVar(&hostConfigFile, "config-file", "", "Path to device configuration file (when using file source)")
	hostScanCmd.Flags().
		StringVar(&hostDeviceType, "device-type", "", "Device OS type (opnsense, openwrt, fortinet, cisco) - auto-detected if not specified")
	hostScanCmd.Flags().
		StringVar(&hostSSHUsername, "ssh-user", "", "SSH username for configuration retrieval")
	hostScanCmd.Flags().
		StringVar(&hostSSHPassword, "ssh-password", "", "SSH password for configuration retrieval")
	hostScanCmd.Flags().StringVar(&hostSSHKeyFile, "ssh-key", "", "Path to SSH private key file")
	hostScanCmd.Flags().IntVar(&hostSSHPort, "ssh-port", 22, "SSH port for configuration retrieval")
	hostScanCmd.Flags().
		StringVar(&hostDiscrepancyAction, "discrepancy-action", "prefer-snmp", "Action when SNMP and config data conflict (fail, prefer-snmp, prefer-config)")
	hostScanCmd.Flags().
		BoolVar(&hostMergeConfig, "merge-configs", false, "Merge SNMP data with configuration data")
	hostScanCmd.Flags().
		IntVar(&hostConfigTimeout, "config-timeout", 60, "Configuration parsing timeout in seconds")
}
