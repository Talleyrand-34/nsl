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

// VlanUpdateCmd represents the vlan update command
var VlanUpdateCmd = &cobra.Command{
	Use:   "vlan",
	Short: "Update an existing VLAN",
	Long: `Update a VLAN's ID or name by specifying the database ID.

You can change the VLAN ID, name, or both.

Example:
  nsl-graph update vlan --id 1 --vlan-id 200 --name "New VLAN Name"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required VLAN database ID
		vlanDbId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if vlanDbId == "" {
			fmt.Fprintf(os.Stderr, "VLAN database ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		// Get new VLAN parameters (at least one must be provided)
		newVlanID, _ := cmd.Flags().GetString("vlan-id")
		newVlanName, _ := cmd.Flags().GetString("name")

		if newVlanID == "" && newVlanName == "" {
			fmt.Fprintf(os.Stderr, "At least one update parameter is required: --vlan-id or --name.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the VLAN
		err = service.UpdateVlan(vlanDbId, newVlanID, newVlanName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating VLAN: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated VLAN with database ID %s\n", vlanDbId)
		if newVlanID != "" {
			fmt.Printf("  New VLAN ID: %s\n", newVlanID)
		}
		if newVlanName != "" {
			fmt.Printf("  New name: %s\n", newVlanName)
		}
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(VlanUpdateCmd)

	VlanUpdateCmd.Flags().String("id", "", "Database ID of the VLAN to update (required)")
	VlanUpdateCmd.Flags().String("vlan-id", "", "New VLAN ID (optional)")
	VlanUpdateCmd.Flags().String("name", "", "New VLAN name (optional)")
}
