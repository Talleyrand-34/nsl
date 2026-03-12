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

	"github.com/spf13/cobra"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
	s "nsl-graph/internal/scanner"
)

var (
	importFile         string
	importAutoImport   bool
	importMergeIPs     bool
	importReviewMode   bool
	importCreateZones  bool
	importDefaultZone  string
	importSkipExisting bool
)

var importScanCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import scan results from a JSON file",
	Long: `Import previously saved SNMP scan results from a JSON file and optionally add devices to the database.

Examples:
  nsl-graph scan import scan_results.json --review
  nsl-graph scan import network_scan.json --auto-import
  nsl-graph scan import results.json --auto-import --merge-ips`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		if importFile != "" {
			filename = importFile
		}

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Loading scan results from %s...\n", filename)

		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Failed to read file: %v\n", err)
			os.Exit(1)
		}

		var scanResult s.ScanResult
		if err := json.Unmarshal(data, &scanResult); err != nil {
			fmt.Printf("Failed to parse scan results: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Loaded scan from %v: %d devices\n",
			scanResult.StartTime.Format("2006-01-02 15:04:05"), len(scanResult.Devices))

		if len(scanResult.Devices) == 0 {
			fmt.Println("No devices in scan results.")
			return
		}

		devices, err := service.DiscoverDevices(&scanResult)
		if err != nil {
			fmt.Printf("Device discovery failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nDiscovered devices:\n")
		for i, device := range devices {
			fmt.Printf("%d. %s (%s) - %s %s [%s]\n",
				i+1,
				device.SuggestedName,
				device.Device.IP,
				device.Brand,
				device.Model,
				device.DeviceClass)

			if len(device.Device.Interfaces) > 0 {
				fmt.Printf("   Interfaces: %d\n", len(device.Device.Interfaces))
			}
		}

		if importAutoImport || importReviewMode {
			importOptions := s.ImportOptions{
				AutoImport:   importAutoImport,
				MergeIPs:     importMergeIPs,
				CreateZones:  importCreateZones,
				DefaultZone:  importDefaultZone,
				SkipExisting: importSkipExisting,
				ReviewMode:   importReviewMode,
			}

			if importReviewMode && !importAutoImport {
				fmt.Printf("\nWould you like to import these %d devices? (y/n): ", len(devices))
				var confirm string
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
					fmt.Println("Import cancelled.")
					return
				}

				fmt.Print("Import all devices? (y/n/s for selective): ")
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) == "s" || strings.ToLower(confirm) == "selective" {
					devices = selectDevicesInteractively(devices)
				}

				importOptions.AutoImport = true
			}

			if len(devices) > 0 {
				fmt.Printf("Importing %d devices...\n", len(devices))
				if err := service.ImportDiscoveredDevices(devices, importOptions); err != nil {
					fmt.Printf("Import failed: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("Devices imported successfully!")
			} else {
				fmt.Println("No devices selected for import.")
			}
		} else {
			fmt.Println("\nUse --auto-import to automatically add devices to database")
			fmt.Println("Use --review to review devices before importing")
		}
	},
}

func selectDevicesInteractively(devices []s.DiscoveredDevice) []s.DiscoveredDevice {
	var selected []s.DiscoveredDevice

	fmt.Println("\nSelect devices to import (enter device numbers, separated by commas):")
	fmt.Println("Examples: 1,3,5 or 1-3,5 or 'all' or 'none'")

	var input string
	fmt.Print("Selection: ")
	fmt.Scanln(&input)

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "all" {
		return devices
	}
	if input == "none" || input == "" {
		return selected
	}

	selectedIndexes := make(map[int]bool)
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			bounds := strings.Split(part, "-")
			if len(bounds) == 2 {
				start := parseIndex(bounds[0])
				end := parseIndex(bounds[1])
				for i := start; i <= end && i <= len(devices); i++ {
					if i > 0 {
						selectedIndexes[i-1] = true
					}
				}
			}
		} else {
			idx := parseIndex(part)
			if idx > 0 && idx <= len(devices) {
				selectedIndexes[idx-1] = true
			}
		}
	}

	for i := range devices {
		if selectedIndexes[i] {
			selected = append(selected, devices[i])
		}
	}

	fmt.Printf("Selected %d devices for import.\n", len(selected))
	return selected
}

func parseIndex(s string) int {
	var index int
	fmt.Sscanf(s, "%d", &index)
	return index
}

func init() {
	cmd_root.ScanCmd.AddCommand(importScanCmd)

	importScanCmd.Flags().StringVarP(&importFile, "file", "f", "", "File to import (alternative to positional arg)")
	importScanCmd.Flags().BoolVar(&importAutoImport, "auto-import", false, "Automatically import all discovered devices")
	importScanCmd.Flags().BoolVar(&importMergeIPs, "merge-ips", false, "Merge IP addresses with existing devices")
	importScanCmd.Flags().BoolVar(&importReviewMode, "review", false, "Review devices before importing")
	importScanCmd.Flags().BoolVar(&importCreateZones, "create-zones", true, "Create zones for discovered devices")
	importScanCmd.Flags().StringVar(&importDefaultZone, "default-zone", "Discovered", "Default zone for discovered devices")
	importScanCmd.Flags().BoolVar(&importSkipExisting, "skip-existing", true, "Skip devices with existing IP addresses")
}
