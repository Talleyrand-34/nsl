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

// ConnectionUpdateCmd represents the connection update command
var ConnectionUpdateCmd = &cobra.Command{
	Use:   "connection",
	Short: "Update an existing connection",
	Long: `Update an existing connection by its ID.
 	
 You can change which device ports are connected by specifying new deviceport IDs.

 Example:
   nsl-graph update connection --id 1 --from-deviceport-id 2 --to-deviceport-id 3`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required connection ID
		connectionId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if connectionId == "" {
			fmt.Fprintf(os.Stderr, "Connection ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		// Get connection parameters
		fromDeviceportId, _ := cmd.Flags().GetString("from-deviceport-id")
		toDeviceportId, _ := cmd.Flags().GetString("to-deviceport-id")

		// All connection parameters are required
		if fromDeviceportId == "" || toDeviceportId == "" {
			fmt.Fprintf(os.Stderr, "All connection parameters are required: --from-deviceport-id, --to-deviceport-id.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the connection
		err = service.UpdateConnection(connectionId, fromDeviceportId, toDeviceportId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating connection: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated connection with ID %s\n", connectionId)
		fmt.Printf("  From deviceport ID: %s\n", fromDeviceportId)
		fmt.Printf("  To deviceport ID: %s\n", toDeviceportId)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(ConnectionUpdateCmd)

	ConnectionUpdateCmd.Flags().String("id", "", "ID of the connection to update (required)")
	ConnectionUpdateCmd.Flags().String("from-deviceport-id", "", "Source deviceport ID (required)")
	ConnectionUpdateCmd.Flags().String("to-deviceport-id", "", "Destination deviceport ID (required)")
}
