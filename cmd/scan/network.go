/*
  Copyright © 2025 Talleyrand-34 (t34@t34.dev)

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
	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
)

var (
	scanSubnet      string
	scanTimeout     int
	scanCommunity   string
	scanSNMPVersion string
	scanSNMPPort    uint16
	scanOutputFile  string
	scanAutoImport  bool
	scanMergeIPs    bool
	scanReviewMode  bool
	scanCreateZones bool
	scanDefaultZone string
	scanDefaultBrand string
	scanSkipExisting bool

	scanProfile     string
	scanSaveProfile string
)

var networkScanCmd = &cobra.Command{
	Use:   "network [subnet]",
	Short: "Scan a network subnet to discover devices via SNMP",
	Long: `Scan a network subnet using SNMP to collect device inventory (interfaces, MACs, VLANs).

Examples:
  nsl-graph scan network 192.168.1.0/24
  nsl-graph scan network 192.168.1.0/24 --community private
  nsl-graph scan network 10.0.0.0/24 --community public --auto-import
  nsl-graph scan network 172.16.1.0/24 --output scan_results.json --review`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		subnet := args[0]
		if scanSubnet != "" {
			subnet = scanSubnet
		}

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		// Apply a saved scan profile (explicit --profile, else auto-matched by
		// subnet). SNMP-only fields; explicit flags override.
		if p, ok := service.ResolveScanProfile(subnet, scanProfile); ok {
			fl := cmd.Flags()
			if !fl.Changed("community") && p.SNMPCommunity != "" {
				scanCommunity = p.SNMPCommunity
			}
			if !fl.Changed("snmp-version") && p.SNMPVersion != "" {
				scanSNMPVersion = p.SNMPVersion
			}
			if !fl.Changed("snmp-port") && p.SNMPPort != 0 {
				scanSNMPPort = uint16(p.SNMPPort)
			}
			if !fl.Changed("timeout") && p.TimeoutSec != 0 {
				scanTimeout = p.TimeoutSec
			}
			fmt.Printf("Applied scan profile %q\n", p.Name)
		} else if scanProfile != "" {
			fmt.Printf("Error: no scan profile named %q\n", scanProfile)
			os.Exit(1)
		}

		if scanSaveProfile != "" {
			p := e.ScanProfile{
				Name:          scanSaveProfile,
				Host:          subnet,
				SNMPCommunity: scanCommunity,
				SNMPVersion:   scanSNMPVersion,
				SNMPPort:      int(scanSNMPPort),
				TimeoutSec:    scanTimeout,
				ScanSource:    "snmp",
			}
			if err := service.AddScanProfile(p); err != nil {
				fmt.Printf("Error saving profile: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Saved scan profile %q (host %s).\n", scanSaveProfile, subnet)
		}

		options := s.ScanOptions{
			Subnet:  subnet,
			Timeout: time.Duration(scanTimeout) * time.Second,
			SNMP: s.SNMPOptions{
				Community: scanCommunity,
				Version:   scanSNMPVersion,
				Port:      scanSNMPPort,
			},
		}

		fmt.Printf("Starting SNMP scan of %s (community=%s, version=%s)...\n",
			subnet, scanCommunity, scanSNMPVersion)

		scanResult, err := service.ScanNetwork(subnet, options)
		if err != nil {
			fmt.Printf("Network scan failed: %v\n", err)
			os.Exit(1)
		}

		duration := scanResult.EndTime.Sub(scanResult.StartTime)
		fmt.Printf("Scan completed in %v\n", duration)
		fmt.Printf("Found %d reachable devices\n", len(scanResult.Devices))

		if scanOutputFile != "" {
			if err := saveScanResults(scanResult, scanOutputFile); err != nil {
				fmt.Printf("Failed to save results: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Results saved to %s\n", scanOutputFile)
		}

		if len(scanResult.Devices) == 0 {
			fmt.Println("No SNMP-reachable devices found.")
			return
		}

		devices, err := service.DiscoverDevices(scanResult)
		if err != nil {
			fmt.Printf("Device discovery failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nDiscovered %d devices:\n", len(devices))
		for _, device := range devices {
			fmt.Printf("  - %s (%s) - %s %s [%s]\n",
				device.SuggestedName,
				device.Device.IP,
				device.Brand,
				device.Model,
				device.DeviceClass)
			if len(device.Device.Interfaces) > 0 {
				fmt.Printf("    Interfaces: %d", len(device.Device.Interfaces))
				for _, iface := range device.Device.Interfaces {
					if iface.MAC != "" {
						fmt.Printf("  %s/%s", iface.Name, iface.MAC)
					}
				}
				fmt.Println()
			}
		}

		if scanAutoImport || scanReviewMode {
			importOptions := s.ImportOptions{
				AutoImport:       scanAutoImport,
				MergeIPs:         scanMergeIPs,
				CreateZones:      scanCreateZones,
				DefaultZone:      scanDefaultZone,
				DefaultBrand:     scanDefaultBrand,
				SkipExisting:     scanSkipExisting,
				ReviewMode:       scanReviewMode,
				InteractiveVLANs: true, // Enable interactive VLAN mapping
			}

			if scanReviewMode && !scanAutoImport {
				fmt.Println("\nReview mode enabled. Would you like to import these devices? (y/n)")
				var confirm string
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
					fmt.Println("Import cancelled.")
					return
				}
				importOptions.AutoImport = true
			}

			fmt.Printf("Importing %d discovered devices with interactive VLAN mapping...\n", len(devices))
			if err := importDevicesWithVLANMapping(service, devices, importOptions); err != nil {
				fmt.Printf("Import failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Devices imported successfully!")
		} else {
			fmt.Println("\nUse --auto-import to automatically add devices to database")
			fmt.Println("Use --review to review devices before importing")
		}
	},
}

func saveScanResults(result *s.ScanResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// importDevicesWithVLANMapping imports devices using interactive VLAN mapping
func importDevicesWithVLANMapping(service q.NetServiceInt, devices []s.DiscoveredDevice, options s.ImportOptions) error {
	for i, device := range devices {
		fmt.Printf("\n=== Device %d/%d: %s (%s) ===\n", i+1, len(devices), device.SuggestedName, device.Device.IP)

		// Use the same logic as scan host for consistency
		if err := importSingleDeviceWithVLANMapping(service, device, options); err != nil {
			fmt.Printf("Failed to import device %s: %v\n", device.SuggestedName, err)
			if !options.ReviewMode {
				return err
			}
		} else {
			fmt.Printf("✓ Device %s imported successfully\n", device.SuggestedName)
		}
	}
	return nil
}


func init() {
	cmd_root.ScanCmd.AddCommand(networkScanCmd)

	networkScanCmd.Flags().StringVar(&scanSubnet, "subnet", "", "Network subnet to scan (alternative to positional arg)")
	networkScanCmd.Flags().IntVarP(&scanTimeout, "timeout", "t", 30, "Timeout in seconds for the scan")
	networkScanCmd.Flags().StringVar(&scanCommunity, "community", "public", "SNMP community string")
	networkScanCmd.Flags().StringVar(&scanSNMPVersion, "snmp-version", "v2c", "SNMP version (v1, v2c)")
	networkScanCmd.Flags().Uint16Var(&scanSNMPPort, "snmp-port", 161, "SNMP UDP port")

	networkScanCmd.Flags().StringVarP(&scanOutputFile, "output", "o", "", "Save scan results to JSON file")

	networkScanCmd.Flags().BoolVar(&scanAutoImport, "auto-import", false, "Automatically import discovered devices")
	networkScanCmd.Flags().BoolVar(&scanMergeIPs, "merge-ips", false, "Merge IP addresses with existing devices")
	networkScanCmd.Flags().BoolVar(&scanReviewMode, "review", false, "Review devices before importing")
	networkScanCmd.Flags().BoolVar(&scanCreateZones, "create-zones", true, "Create zones for discovered devices")
	networkScanCmd.Flags().StringVar(&scanDefaultZone, "default-zone", "Discovered", "Default zone for discovered devices")
	networkScanCmd.Flags().StringVar(&scanDefaultBrand, "default-brand", "", "Default brand for unidentified devices")
	networkScanCmd.Flags().BoolVar(&scanSkipExisting, "skip-existing", true, "Skip devices with existing IP addresses")

	networkScanCmd.Flags().StringVar(&scanProfile, "profile", "", "Use a saved scan profile by name (else auto-matched by subnet)")
	networkScanCmd.Flags().StringVar(&scanSaveProfile, "save-profile", "", "Save the effective SNMP parameters as a profile with this name")
}
