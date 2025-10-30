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
package cmd_delete

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// vlanDeleteCmd represents the vlan delete command
var vlanDeleteCmd = &cobra.Command{
	Use:   "vlan",
	Short: "Delete a VLAN",
	Long: `Delete a VLAN by its database ID.

Example:
  nsl-graph delete vlan --id 1`,
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

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Delete the VLAN
		err = service.DeleteVlan(vlanDbId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting VLAN: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully deleted VLAN with database ID %s\n", vlanDbId)
	},
}

func init() {
	cmd.DeleteCmd.AddCommand(vlanDeleteCmd)

	vlanDeleteCmd.Flags().String("id", "", "Database ID of the VLAN to delete (required)")
	vlanDeleteCmd.MarkFlagRequired("id")
}
