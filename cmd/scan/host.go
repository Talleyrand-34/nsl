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
)

var hostScanCmd = &cobra.Command{
	Use:   "host [ip_address]",
	Short: "Query a specific host via SNMP to collect inventory",
	Long: `Query a specific device using SNMP to collect interfaces, MAC addresses, and VLAN assignments.

Examples:
  nsl-graph scan host 192.168.1.1
  nsl-graph scan host 10.0.0.1 --community private
  nsl-graph scan host 192.168.1.1 --output device_scan.json
  nsl-graph scan host 172.16.1.10 --auto-import
  nsl-graph scan host 192.168.1.254 --community public --snmp-version v2c`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ip := args[0]

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		options := s.ScanOptions{
			Timeout: time.Duration(hostTimeout) * time.Second,
			SNMP: s.SNMPOptions{
				Community: hostCommunity,
				Version:   hostSNMPVersion,
				Port:      hostSNMPPort,
			},
		}

		fmt.Printf("Querying %s via SNMP (community=%s)...\n", ip, hostCommunity)

		device, err := service.ScanDevice(ip, options)
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
			// Print header for columnar display
			fmt.Printf("  %-30s %-17s %-6s %-20s %-8s %-8s\n", "Interface", "MAC", "Status", "IPs", "VLAN", "Accuracy")
			fmt.Printf("  %-30s %-17s %-6s %-20s %-8s %-8s\n", strings.Repeat("-", 30), strings.Repeat("-", 17), strings.Repeat("-", 6), strings.Repeat("-", 20), strings.Repeat("-", 8), strings.Repeat("-", 8))

			for _, iface := range device.Interfaces {
				status := "down"
				if iface.OperStatus == 1 {
					status = "up"
				}

				// Infer VLAN information
				vlanInference := s.InferVLANFromInterface(iface.Name, iface.IPAddresses)

				// Format IP addresses for display
				ipsDisplay := "[]"
				if len(iface.IPAddresses) > 0 {
					ipsDisplay = fmt.Sprintf("%v", iface.IPAddresses)
					// Truncate if too long
					if len(ipsDisplay) > 18 {
						ipsDisplay = ipsDisplay[:15] + "..."
					}
				}

				// Display the interface row
				fmt.Printf("  %-30s %-17s %-6s %-20s %-8s %-8s\n",
					iface.Name,
					iface.MAC,
					status,
					ipsDisplay,
					vlanInference.FormatVLANDisplay(),
					vlanInference.FormatAccuracyDisplay())

				// Show additional VLAN details if explicitly detected
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

				// Show inference notes if any
				if vlanInference.Notes != "" {
					fmt.Printf("  %30s   Note: %s\n", "", vlanInference.Notes)
				}
			}
		}

		if len(device.Neighbors) > 0 {
			fmt.Printf("\nNeighbors (%d):\n", len(device.Neighbors))
			for _, n := range device.Neighbors {
				fmt.Printf("  [%s] local:%s → remote:%s (%s)", n.Protocol, n.LocalPort, n.RemoteName, n.RemotePort)
				if n.RemoteIP != "" {
					fmt.Printf(" IP:%s", n.RemoteIP)
				}
				fmt.Println()
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

			fmt.Printf("Classification: %s %s [%s]\n", discoveredDevice.Brand, discoveredDevice.Model, discoveredDevice.DeviceClass)

			importOptions := s.ImportOptions{
				AutoImport:        true,
				CreateZones:       true,
				DefaultZone:       hostDefaultZone,
				DefaultBrand:      hostDefaultBrand,
				SkipExisting:      true,
				InteractiveVLANs:  true, // Enable interactive VLAN mapping
				VLANAccuracyLevel: hostVLANAccuracy,
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

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// importSingleDeviceWithVLANMapping imports a device with interactive VLAN mapping
func importSingleDeviceWithVLANMapping(service q.NetServiceInt, device s.DiscoveredDevice, options s.ImportOptions) error {
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
				fmt.Printf("\n      %s → VLAN %s (%s)", mapping.IP, mapping.VLANNumber, mapping.Confidence)
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

	hostScanCmd.Flags().IntVarP(&hostTimeout, "timeout", "t", 10, "Timeout in seconds for the SNMP query")
	hostScanCmd.Flags().StringVar(&hostCommunity, "community", "public", "SNMP community string")
	hostScanCmd.Flags().StringVar(&hostSNMPVersion, "snmp-version", "v2c", "SNMP version (v1, v2c)")
	hostScanCmd.Flags().Uint16Var(&hostSNMPPort, "snmp-port", 161, "SNMP UDP port")

	hostScanCmd.Flags().StringVarP(&hostOutputFile, "output", "o", "", "Save scan results to JSON file")

	hostScanCmd.Flags().BoolVar(&hostAutoImport, "auto-import", false, "Automatically import discovered device")
	hostScanCmd.Flags().StringVar(&hostDefaultZone, "default-zone", "", "Default zone for the device (defaults to Generic)")
	hostScanCmd.Flags().StringVar(&hostDefaultBrand, "default-brand", "", "Default brand for unidentified device")
	hostScanCmd.Flags().BoolVar(&hostApproveAll, "approve-all", false, "Automatically approve all VLAN mappings without prompting (only with --auto-import)")
	hostScanCmd.Flags().IntVar(&hostVLANAccuracy, "vlan-accuracy", 1, "VLAN detection accuracy level (1=interface names only, 2=include IP heuristics)")
}
