// SPDX-License-Identifier: AGPL-3.0-or-later
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
package cmd_update

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// ZoneUpdateCmd represents the zone update command
var ZoneUpdateCmd = &cobra.Command{
	Use:   "zone",
	Short: "Update an existing zone",
	Long: `Update an existing zone by its ID.
	
You can update the zone name, parent zone, zone type, and owner assignment.
All foreign key references should be provided as IDs.

Examples:
  nsl-graph update zone --id 1 --name "New Zone Name"
  nsl-graph update zone --id 1 --name "DataCenter-02" --father-zone-id 2 --zonetype-id 1 --owner-id 1`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required zone ID
		zoneId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if zoneId == "" {
			fmt.Fprintf(os.Stderr, "Zone ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		// Get update fields
		newZoneName, _ := cmd.Flags().GetString("name")
		newFatherZoneId, _ := cmd.Flags().GetString("father-zone-id")
		newZoneTypeId, _ := cmd.Flags().GetString("zonetype-id")
		newOwnerId, _ := cmd.Flags().GetString("owner-id")

		// At least zone name is required
		if newZoneName == "" {
			fmt.Fprintf(os.Stderr, "Zone name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the zone
		err = service.UpdateZone(zoneId, newZoneName, newFatherZoneId, newZoneTypeId, newOwnerId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating zone: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated zone with ID %s\n", zoneId)
		fmt.Printf("  New name: %s\n", newZoneName)
		if newFatherZoneId != "" {
			fmt.Printf("  New father zone ID: %s\n", newFatherZoneId)
		}
		if newZoneTypeId != "" {
			fmt.Printf("  New zone type ID: %s\n", newZoneTypeId)
		}
		if newOwnerId != "" {
			fmt.Printf("  New owner ID: %s\n", newOwnerId)
		}
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(ZoneUpdateCmd)

	ZoneUpdateCmd.Flags().String("id", "", "ID of the zone to update (required)")
	ZoneUpdateCmd.Flags().String("name", "", "New name for the zone (required)")
	ZoneUpdateCmd.Flags().String("father-zone-id", "", "New father zone ID (optional)")
	ZoneUpdateCmd.Flags().String("zonetype-id", "", "New zone type ID (optional)")
	ZoneUpdateCmd.Flags().String("owner-id", "", "New owner ID (optional)")
	// ZoneUpdateCmd.MarkFlagRequired("id")
	// ZoneUpdateCmd.MarkFlagRequired("name")
}
