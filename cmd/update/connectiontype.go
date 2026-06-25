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

// ConnectionTypeUpdateCmd represents the connection type update command
var ConnectionTypeUpdateCmd = &cobra.Command{
	Use:   "connectiontype",
	Short: "Update an existing connection type",
	Long: `Update an existing connection type by its ID.
	
A connection type describes the medium or protocol used for network connections (e.g., ethernet, fiber, wireless).

Example:
  nsl-graph update connectiontype --id 1 --name "Single-Mode Fiber"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required flags
		connectionTypeId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if connectionTypeId == "" {
			fmt.Fprintf(os.Stderr, "Connection type ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		newConnectionTypeName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'name': %v\n", err)
			os.Exit(1)
		}
		if newConnectionTypeName == "" {
			fmt.Fprintf(os.Stderr, "New connection type name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the connection type
		err = service.UpdateConnectionType(connectionTypeId, newConnectionTypeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating connection type: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated connection type with ID %s to '%s'\n", connectionTypeId, newConnectionTypeName)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(ConnectionTypeUpdateCmd)

	ConnectionTypeUpdateCmd.Flags().String("id", "", "ID of the connection type to update (required)")
	ConnectionTypeUpdateCmd.Flags().String("name", "", "New name for the connection type (required)")
	// ConnectionTypeUpdateCmd.MarkFlagRequired("id")
	// ConnectionTypeUpdateCmd.MarkFlagRequired("name")
}
