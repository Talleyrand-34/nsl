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
package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// zoneModCmd represents the zone creation command
var zoneModCmd = &cobra.Command{
	Use:   "zone",
	Short: "Add a network zone",
	Long: `Create a new network zone with specified hierarchy and ownership.
	
A zone represents a logical or physical grouping of network devices (like a datacenter, building, or rack).
Zones can have parent-child relationships to model hierarchical network structures.

Examples:
  nsl-graph add zone --name "DataCenter-1" --owner "IT Department" --zonetype "Datacenter"
  nsl-graph add zone --name "Rack-A1" --father-name "DataCenter-1" --owner "IT Department" --zonetype "Rack"
  nsl-graph add zone --name "Floor-2" --father-id 1 --owner "IT Department" --zonetype "Floor"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get required zone name
		zoneName, err := cmd.Flags().GetString("name")
		if err != nil || zoneName == "" {
			fmt.Fprintf(os.Stderr, "Zone name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get required owner
		ownerName, err := cmd.Flags().GetString("owner")
		if err != nil || ownerName == "" {
			fmt.Fprintf(os.Stderr, "Owner owner is required. Use --owner flag.\n")
			os.Exit(1)
		}

		// Get required zone type
		zoneTypeName, err := cmd.Flags().GetString("zonetype")
		if err != nil || zoneTypeName == "" {
			fmt.Fprintf(os.Stderr, "Zone type is required. Use --zonetype flag.\n")
			os.Exit(1)
		}

		// Get parent zone information (optional)
		fatherZoneId, _ := cmd.Flags().GetString("father-id")
		fatherZoneName, _ := cmd.Flags().GetString("father-name")

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Create the zone
		err = service.AddZone(zoneName, fatherZoneId, fatherZoneName, ownerName, zoneTypeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating zone: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created zone '%s' of type '%s' owned by '%s'",
			zoneName, zoneTypeName, ownerName)
		if fatherZoneId != "" {
			fmt.Printf(" under parent zone ID %s", fatherZoneId)
		} else if fatherZoneName != "" {
			fmt.Printf(" under parent zone '%s'", fatherZoneName)
		}
		fmt.Println()
	},
}

func init() {
	cmd.AddCmd.AddCommand(zoneModCmd)

	zoneModCmd.Flags().String("name", "", "Zone name/identifier (required)")
	zoneModCmd.Flags().String("father-name", "", "Parent zone name (optional)")
	zoneModCmd.Flags().String("father-id", "", "Parent zone ID (optional, takes precedence over father-name)")
	zoneModCmd.Flags().String("owner", "", "Owner owner name (required)")
	zoneModCmd.Flags().String("zonetype", "", "Zone type name (required)")

	zoneModCmd.MarkFlagRequired("name")
	zoneModCmd.MarkFlagRequired("owner")
	zoneModCmd.MarkFlagRequired("zonetype")
}
