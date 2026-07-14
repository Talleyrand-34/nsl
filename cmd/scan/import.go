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
	importHuman        bool
)

// isJSONArray reports whether the first non-whitespace byte of data is '[',
// i.e. the document is a JSON array (an import-plan list) rather than an object
// (a raw ScanResult).
func isJSONArray(data []byte) bool {
	for _, b := range data {
		switch b {
		case ' ', '\t', '\r', '\n':
			continue
		case '[':
			return true
		default:
			return false
		}
	}
	return false
}

var importScanCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import devices from a scan JSON file (raw scan result or an edited import plan)",
	Long: `Import devices from a JSON file produced by a scan. Two shapes are accepted and
auto-detected:

  • An import plan — a JSON array of plans, as emitted by "scan run" (and the web
    UI). This is the edit-and-import half of the round-trip: each plan is executed
    as you edited it. Use -H/--human to review (Approve/Edit/Skip/Quit) per device.
  • A raw scan result — a JSON object as written by "scan host/network --raw".
    It is discovered, analyzed, then imported.

Examples:
  nsl-graph scan run 10.0.2.0/24 > plan.json && $EDITOR plan.json
  nsl-graph scan import plan.json                 # execute the edited plan
  nsl-graph scan import plan.json -H              # review each device first
  nsl-graph scan import network_scan.json --auto-import   # raw scan result`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		if importFile != "" {
			filename = importFile
		}

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Loading scan from %s...\n", filename)

		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
			os.Exit(1)
		}

		// Auto-detect the artifact: a top-level JSON array is an import plan list
		// (the editable pipeline artifact); an object is a raw ScanResult.
		if isJSONArray(data) {
			var plans []s.DeviceImportPlan
			if err := json.Unmarshal(data, &plans); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse import plan: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Loaded import plan: %d device(s)\n", len(plans))
			if len(plans) == 0 {
				fmt.Fprintln(os.Stderr, "No devices in import plan.")
				return
			}
			options := s.ImportOptions{
				AutoImport:        true,
				MergeIPs:          importMergeIPs,
				CreateZones:       importCreateZones,
				DefaultZone:       importDefaultZone,
				SkipExisting:      importSkipExisting,
				InteractiveVLANs:  true,
				VLANAccuracyLevel: importVLANAccuracy,
			}
			if err := executePlans(service, plans, options, importHuman); err != nil {
				fmt.Fprintf(os.Stderr, "Import failed: %v\n", err)
				os.Exit(1)
			}
			return
		}

		var scanResult s.ScanResult
		if err := json.Unmarshal(data, &scanResult); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse scan results: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Loaded scan from %v: %d devices\n",
			scanResult.StartTime.Format("2006-01-02 15:04:05"), len(scanResult.Devices))

		if len(scanResult.Devices) == 0 {
			fmt.Fprintln(os.Stderr, "No devices in scan results.")
			return
		}

		devices, err := service.DiscoverDevices(&scanResult)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Device discovery failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "\nDiscovered devices:\n")
		for i, device := range devices {
			fmt.Fprintf(os.Stderr, "%d. %s (%s) - %s %s [%s]\n",
				i+1,
				device.SuggestedName,
				device.Device.IP,
				device.Brand,
				device.Model,
				device.ModelType)

			if len(device.Device.Interfaces) > 0 {
				fmt.Fprintf(os.Stderr, "   Interfaces: %d\n", len(device.Device.Interfaces))
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
				fmt.Fprintf(os.Stderr, "\nWould you like to import these %d devices? (y/n): ", len(devices))
				var confirm string
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
					fmt.Fprintln(os.Stderr, "Import cancelled.")
					return
				}

				fmt.Fprint(os.Stderr, "Import all devices? (y/n/s for selective): ")
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) == "s" || strings.ToLower(confirm) == "selective" {
					devices = selectDevicesInteractively(devices)
				}

				importOptions.AutoImport = true
			}

			if len(devices) > 0 {
				fmt.Fprintf(os.Stderr, "Analyzing %d devices for interactive VLAN mapping...\n", len(devices))
				if err := importDevicesWithInteractiveVLANMapping(service, devices, importOptions); err != nil {
					fmt.Fprintf(os.Stderr, "Import failed: %v\n", err)
					os.Exit(1)
				}
				fmt.Fprintln(os.Stderr, "Devices imported successfully!")
			} else {
				fmt.Fprintln(os.Stderr, "No devices selected for import.")
			}
		} else {
			fmt.Fprintln(os.Stderr, "\nUse --auto-import to automatically add devices to database")
			fmt.Fprintln(os.Stderr, "Use --review to review devices before importing")
		}
	},
}

func selectDevicesInteractively(devices []s.DiscoveredDevice) []s.DiscoveredDevice {
	var selected []s.DiscoveredDevice

	fmt.Fprintln(os.Stderr, "\nSelect devices to import (enter device numbers, separated by commas):")
	fmt.Fprintln(os.Stderr, "Examples: 1,3,5 or 1-3,5 or 'all' or 'none'")

	var input string
	fmt.Fprint(os.Stderr, "Selection: ")
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

	fmt.Fprintf(os.Stderr, "Selected %d devices for import.\n", len(selected))
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
	importScanCmd.Flags().BoolVarP(&importHuman, "human", "H", false, "For an import-plan file: review each device (Approve/Edit/Skip/Quit) instead of executing it directly")
}

// importDevicesWithInteractiveVLANMapping performs interactive VLAN mapping and import
func importDevicesWithInteractiveVLANMapping(service q.NetServiceInt, devices []s.DiscoveredDevice, options s.ImportOptions) error {
	for i, device := range devices {
		fmt.Fprintf(os.Stderr, "\n=== Device %d/%d: %s (%s) ===\n", i+1, len(devices), device.SuggestedName, device.Device.IP)

		// Analyze device for import plan
		plan, err := service.AnalyzeDeviceForImport(device)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to analyze device: %v\n", err)
			continue
		}

		// Show analysis summary
		fmt.Fprintf(os.Stderr, "Summary: %s\n", plan.Summary)

		// Display detailed interface analysis
		if err := displayImportPlan(plan); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to display plan: %v\n", err)
			continue
		}

		// Get user decision (respecting non-interactive flags)
		action, modifiedPlan, err := getUserImportDecision(plan)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting user input: %v\n", err)
			continue
		}

		switch action {
		case s.ActionApprove:
			fmt.Fprintln(os.Stderr, "Importing device with approved plan...")
			if err := service.ExecuteApprovedImportPlan(modifiedPlan, options); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to execute import plan: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "✓ Device %s imported successfully\n", device.SuggestedName)
			}
		case s.ActionSkip:
			fmt.Fprintf(os.Stderr, "⏭ Skipping device %s\n", device.SuggestedName)
		case s.ActionSkipDevice:
			fmt.Fprintf(os.Stderr, "⏭ Skipping device %s\n", device.SuggestedName)
		case s.ActionQuit:
			fmt.Fprintln(os.Stderr, "Import cancelled by user.")
			return nil
		}
	}

	return nil
}

// displayImportPlan shows the detailed import plan to the user
func displayImportPlan(plan s.DeviceImportPlan) error {
	for i, interfacePlan := range plan.InterfacePlans {
		iface := interfacePlan.Interface
		fmt.Fprintf(os.Stderr, "\n  Interface %d: %s (MAC: %s)\n", i+1, iface.Name, iface.MAC)

		if len(iface.IPAddresses) > 0 {
			fmt.Fprintf(os.Stderr, "    IP Addresses: %v\n", iface.IPAddresses)
		}

		if len(iface.VLANs) > 0 {
			fmt.Fprintf(os.Stderr, "    VLAN Memberships:\n")
			for _, vlan := range iface.VLANs {
				taggedStr := "untagged"
				if vlan.Tagged {
					taggedStr = "tagged"
				}
				fmt.Fprintf(os.Stderr, "      - VLAN %s (%s)\n", vlan.VLANNumber, taggedStr)
			}
		}

		if len(interfacePlan.IPMappings) > 0 {
			fmt.Fprintf(os.Stderr, "    Proposed IP-to-VLAN Mappings:\n")
			for j, mapping := range interfacePlan.IPMappings {
				confidence := getConfidenceSymbol(mapping.Confidence)
				fmt.Fprintf(os.Stderr, "      [%d] %s → VLAN %s %s\n", j+1, mapping.IP, mapping.VLANNumber, confidence)
				fmt.Fprintf(os.Stderr, "          Reason: %s\n", mapping.Reason)
			}
		}

		if len(interfacePlan.VLANsToCreate) > 0 {
			fmt.Fprintf(os.Stderr, "    VLANs to Create:\n")
			for _, vlanPlan := range interfacePlan.VLANsToCreate {
				fmt.Fprintf(os.Stderr, "      - VLAN %s: %s (IP segments: %v)\n",
					vlanPlan.VLANNumber, vlanPlan.VLANName, vlanPlan.IPSegmentIDs)
			}
		}

		if len(interfacePlan.VLANsToUpdate) > 0 {
			fmt.Fprintf(os.Stderr, "    VLANs to Update:\n")
			for _, vlanPlan := range interfacePlan.VLANsToUpdate {
				fmt.Fprintf(os.Stderr, "      - VLAN %s: Add IP segments %v\n",
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
		fmt.Fprintf(os.Stderr, "Auto-approving import (--approve-all flag)\n")
		return s.ActionApprove, plan, nil
	}

	if importSkipAll {
		fmt.Fprintf(os.Stderr, "Auto-skipping device (--skip-all flag)\n")
		return s.ActionSkip, plan, nil
	}

	if importQuitOnFirst {
		fmt.Fprintf(os.Stderr, "Auto-quitting on first device (--quit-on-first flag)\n")
		return s.ActionQuit, plan, nil
	}

	// Interactive mode
	for {
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fmt.Fprintf(os.Stderr, "  (A)pprove and import\n")
		fmt.Fprintf(os.Stderr, "  (E)dit mappings\n")
		fmt.Fprintf(os.Stderr, "  (S)kip this device\n")
		fmt.Fprintf(os.Stderr, "  (Q)uit import process\n")
		fmt.Fprintf(os.Stderr, "Choice (A/E/S/Q): ")

		var choice string
		fmt.Scanln(&choice)
		choice = strings.ToUpper(strings.TrimSpace(choice))

		switch choice {
		case "A", "APPROVE":
			return s.ActionApprove, plan, nil
		case "E", "EDIT":
			modifiedPlan, err := editImportPlan(plan)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error editing plan: %v\n", err)
				continue
			}
			return s.ActionApprove, modifiedPlan, nil
		case "S", "SKIP":
			return s.ActionSkip, plan, nil
		case "Q", "QUIT":
			return s.ActionQuit, plan, nil
		default:
			fmt.Fprintf(os.Stderr, "Invalid choice. Please enter A, E, S, or Q.\n")
		}
	}
}

// editImportPlan allows the user to edit IP-to-VLAN mappings
func editImportPlan(plan s.DeviceImportPlan) (s.DeviceImportPlan, error) {
	fmt.Fprintln(os.Stderr, "\n=== Edit IP-to-VLAN Mappings ===")

	for ifaceIndex, interfacePlan := range plan.InterfacePlans {
		if len(interfacePlan.IPMappings) == 0 {
			continue
		}

		fmt.Fprintf(os.Stderr, "\nInterface: %s\n", interfacePlan.Interface.Name)

		for ipIndex, mapping := range interfacePlan.IPMappings {
			fmt.Fprintf(os.Stderr, "  [%d] %s → VLAN %s (%s)\n",
				ipIndex+1, mapping.IP, mapping.VLANNumber, mapping.Confidence)
			fmt.Fprintf(os.Stderr, "      Enter new VLAN ID (or press Enter to keep current): ")

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

				fmt.Fprintf(os.Stderr, "      ✓ Updated to VLAN %s\n", newVLAN)
			}
		}
	}

	// Regenerate VLAN plans based on modified mappings
	// This is a simplified version - in production, you'd want to call the service method
	fmt.Fprintln(os.Stderr, "\n✓ Import plan updated with your changes.")

	return plan, nil
}
