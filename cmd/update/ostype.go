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

// osTypeUpdateCmd represents the OS type update command
var osTypeUpdateCmd = &cobra.Command{
	Use:   "ostype",
	Short: "Update an existing OS type",
	Long: `Update an existing OS type by its ID.
	
A OS type categorizes network equipment types (e.g., router, switch, firewall).

Example:
  nsl-graph update ostype --id 1 --name "Layer3Switch"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required flags
		osTypeId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if osTypeId == "" {
			fmt.Fprintf(os.Stderr, "OS type ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		newOsTypeName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'name': %v\n", err)
			os.Exit(1)
		}
		if newOsTypeName == "" {
			fmt.Fprintf(os.Stderr, "New OS type name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the OS type
		err = service.UpdateOsType(osTypeId, newOsTypeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating OS type: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated OS type with ID %s to '%s'\n", osTypeId, newOsTypeName)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(osTypeUpdateCmd)

	osTypeUpdateCmd.Flags().String("id", "", "ID of the OS type to update (required)")
	osTypeUpdateCmd.Flags().String("name", "", "New name for the OS type (required)")
	osTypeUpdateCmd.MarkFlagRequired("id")
	osTypeUpdateCmd.MarkFlagRequired("name")
}
