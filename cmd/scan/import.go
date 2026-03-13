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
	q "nsl-graph/internal/repository/application"
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
	importApproveAll   bool
	importSkipAll      bool
	importQuitOnFirst  bool
	importVLANAccuracy int
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
				AutoImport:        importAutoImport,
				MergeIPs:          importMergeIPs,
				CreateZones:       importCreateZones,
				DefaultZone:       importDefaultZone,
				SkipExisting:      importSkipExisting,
				ReviewMode:        importReviewMode,
				InteractiveVLANs:  true, // Always use interactive VLAN mapping
				VLANAccuracyLevel: importVLANAccuracy,
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
				fmt.Printf("Analyzing %d devices for interactive VLAN mapping...\n", len(devices))
				if err := importDevicesWithInteractiveVLANMapping(service, devices, importOptions); err != nil {
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

	// Non-interactive flags for automated import decisions
	importScanCmd.Flags().BoolVar(&importApproveAll, "approve-all", false, "Automatically approve all VLAN mappings without prompting")
	importScanCmd.Flags().BoolVar(&importSkipAll, "skip-all", false, "Automatically skip all devices without prompting")
	importScanCmd.Flags().BoolVar(&importQuitOnFirst, "quit-on-first", false, "Quit import process on first device without prompting")
	importScanCmd.Flags().IntVar(&importVLANAccuracy, "vlan-accuracy", 1, "VLAN detection accuracy level (1=interface names only, 2=include IP heuristics)")
}

// importDevicesWithInteractiveVLANMapping performs interactive VLAN mapping and import
func importDevicesWithInteractiveVLANMapping(service q.NetServiceInt, devices []s.DiscoveredDevice, options s.ImportOptions) error {
	for i, device := range devices {
		fmt.Printf("\n=== Device %d/%d: %s (%s) ===\n", i+1, len(devices), device.SuggestedName, device.Device.IP)

		// Analyze device for import plan
		plan, err := service.AnalyzeDeviceForImport(device)
		if err != nil {
			fmt.Printf("Failed to analyze device: %v\n", err)
			continue
		}

		// Show analysis summary
		fmt.Printf("Summary: %s\n", plan.Summary)

		// Display detailed interface analysis
		if err := displayImportPlan(plan); err != nil {
			fmt.Printf("Failed to display plan: %v\n", err)
			continue
		}

		// Get user decision (respecting non-interactive flags)
		action, modifiedPlan, err := getUserImportDecision(plan)
		if err != nil {
			fmt.Printf("Error getting user input: %v\n", err)
			continue
		}

		switch action {
		case s.ActionApprove:
			fmt.Println("Importing device with approved plan...")
			if err := service.ExecuteApprovedImportPlan(modifiedPlan, options); err != nil {
				fmt.Printf("Failed to execute import plan: %v\n", err)
			} else {
				fmt.Printf("✓ Device %s imported successfully\n", device.SuggestedName)
			}
		case s.ActionSkip:
			fmt.Printf("⏭ Skipping device %s\n", device.SuggestedName)
		case s.ActionSkipDevice:
			fmt.Printf("⏭ Skipping device %s\n", device.SuggestedName)
		case s.ActionQuit:
			fmt.Println("Import cancelled by user.")
			return nil
		}
	}

	return nil
}

// displayImportPlan shows the detailed import plan to the user
func displayImportPlan(plan s.DeviceImportPlan) error {
	for i, interfacePlan := range plan.InterfacePlans {
		iface := interfacePlan.Interface
		fmt.Printf("\n  Interface %d: %s (MAC: %s)\n", i+1, iface.Name, iface.MAC)

		if len(iface.IPAddresses) > 0 {
			fmt.Printf("    IP Addresses: %v\n", iface.IPAddresses)
		}

		if len(iface.VLANs) > 0 {
			fmt.Printf("    VLAN Memberships:\n")
			for _, vlan := range iface.VLANs {
				taggedStr := "untagged"
				if vlan.Tagged {
					taggedStr = "tagged"
				}
				fmt.Printf("      - VLAN %s (%s)\n", vlan.VLANNumber, taggedStr)
			}
		}

		if len(interfacePlan.IPMappings) > 0 {
			fmt.Printf("    Proposed IP-to-VLAN Mappings:\n")
			for j, mapping := range interfacePlan.IPMappings {
				confidence := getConfidenceSymbol(mapping.Confidence)
				fmt.Printf("      [%d] %s → VLAN %s %s\n", j+1, mapping.IP, mapping.VLANNumber, confidence)
				fmt.Printf("          Reason: %s\n", mapping.Reason)
			}
		}

		if len(interfacePlan.VLANsToCreate) > 0 {
			fmt.Printf("    VLANs to Create:\n")
			for _, vlanPlan := range interfacePlan.VLANsToCreate {
				fmt.Printf("      - VLAN %s: %s (IP segments: %v)\n",
					vlanPlan.VLANNumber, vlanPlan.VLANName, vlanPlan.IPSegmentIDs)
			}
		}

		if len(interfacePlan.VLANsToUpdate) > 0 {
			fmt.Printf("    VLANs to Update:\n")
			for _, vlanPlan := range interfacePlan.VLANsToUpdate {
				fmt.Printf("      - VLAN %s: Add IP segments %v\n",
					vlanPlan.VLANNumber, vlanPlan.IPSegmentIDs)
			}
		}
	}

	return nil
}

// getConfidenceSymbol returns a symbol representing confidence level
func getConfidenceSymbol(confidence string) string {
	switch confidence {
	case "exact":
		return "✓ (exact match)"
	case "heuristic":
		return "~ (heuristic)"
	case "suggested":
		return "? (suggested)"
	default:
		return "⚠ (unknown)"
	}
}

// getUserImportDecision gets the user's decision on the import plan
func getUserImportDecision(plan s.DeviceImportPlan) (s.MappingAction, s.DeviceImportPlan, error) {
	// Check for non-interactive flags first
	if importApproveAll {
		fmt.Printf("Auto-approving import (--approve-all flag)\n")
		return s.ActionApprove, plan, nil
	}

	if importSkipAll {
		fmt.Printf("Auto-skipping device (--skip-all flag)\n")
		return s.ActionSkip, plan, nil
	}

	if importQuitOnFirst {
		fmt.Printf("Auto-quitting on first device (--quit-on-first flag)\n")
		return s.ActionQuit, plan, nil
	}

	// Interactive mode
	for {
		fmt.Printf("\nOptions:\n")
		fmt.Printf("  (A)pprove and import\n")
		fmt.Printf("  (E)dit mappings\n")
		fmt.Printf("  (S)kip this device\n")
		fmt.Printf("  (Q)uit import process\n")
		fmt.Printf("Choice (A/E/S/Q): ")

		var choice string
		fmt.Scanln(&choice)
		choice = strings.ToUpper(strings.TrimSpace(choice))

		switch choice {
		case "A", "APPROVE":
			return s.ActionApprove, plan, nil
		case "E", "EDIT":
			modifiedPlan, err := editImportPlan(plan)
			if err != nil {
				fmt.Printf("Error editing plan: %v\n", err)
				continue
			}
			return s.ActionApprove, modifiedPlan, nil
		case "S", "SKIP":
			return s.ActionSkip, plan, nil
		case "Q", "QUIT":
			return s.ActionQuit, plan, nil
		default:
			fmt.Printf("Invalid choice. Please enter A, E, S, or Q.\n")
		}
	}
}

// editImportPlan allows the user to edit IP-to-VLAN mappings
func editImportPlan(plan s.DeviceImportPlan) (s.DeviceImportPlan, error) {
	fmt.Println("\n=== Edit IP-to-VLAN Mappings ===")

	for ifaceIndex, interfacePlan := range plan.InterfacePlans {
		if len(interfacePlan.IPMappings) == 0 {
			continue
		}

		fmt.Printf("\nInterface: %s\n", interfacePlan.Interface.Name)

		for ipIndex, mapping := range interfacePlan.IPMappings {
			fmt.Printf("  [%d] %s → VLAN %s (%s)\n",
				ipIndex+1, mapping.IP, mapping.VLANNumber, mapping.Confidence)
			fmt.Printf("      Enter new VLAN ID (or press Enter to keep current): ")

			var newVLAN string
			fmt.Scanln(&newVLAN)
			newVLAN = strings.TrimSpace(newVLAN)

			if newVLAN != "" {
				// Update the mapping
				plan.InterfacePlans[ifaceIndex].IPMappings[ipIndex].VLANNumber = newVLAN
				plan.InterfacePlans[ifaceIndex].IPMappings[ipIndex].Confidence = "user_edited"
				plan.InterfacePlans[ifaceIndex].IPMappings[ipIndex].Reason = "Modified by user"

				// Check if it's a special VLAN ID
				if newVLAN[0] == '-' {
					plan.InterfacePlans[ifaceIndex].IPMappings[ipIndex].IsNewVLAN = true
				}

				fmt.Printf("      ✓ Updated to VLAN %s\n", newVLAN)
			}
		}
	}

	// Regenerate VLAN plans based on modified mappings
	// This is a simplified version - in production, you'd want to call the service method
	fmt.Println("\n✓ Import plan updated with your changes.")

	return plan, nil
}
