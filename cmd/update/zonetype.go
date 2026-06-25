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

// ZoneTypeUpdateCmd represents the zone type update command
var ZoneTypeUpdateCmd = &cobra.Command{
	Use:   "zonetype",
	Short: "Update an existing zone type",
	Long: `Update an existing zone type by its ID.
	
A zone type categorizes zones (e.g., physical, logical, security, management).

Example:
  nsl-graph update zonetype --id 1 --name "Physical-Secure"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required flags
		zoneTypeId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if zoneTypeId == "" {
			fmt.Fprintf(os.Stderr, "Zone type ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		newZoneTypeName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'name': %v\n", err)
			os.Exit(1)
		}
		if newZoneTypeName == "" {
			fmt.Fprintf(os.Stderr, "New zone type name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the zone type
		err = service.UpdateZoneType(zoneTypeId, newZoneTypeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating zone type: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated zone type with ID %s to '%s'\n", zoneTypeId, newZoneTypeName)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(ZoneTypeUpdateCmd)

	ZoneTypeUpdateCmd.Flags().String("id", "", "ID of the zone type to update (required)")
	ZoneTypeUpdateCmd.Flags().String("name", "", "New name for the zone type (required)")
	// ZoneTypeUpdateCmd.MarkFlagRequired("id")
	// ZoneTypeUpdateCmd.MarkFlagRequired("name")
}
