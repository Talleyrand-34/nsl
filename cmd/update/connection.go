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
	
You can change which devices and ports are connected by specifying new device IDs and model port IDs.

Example:
  nsl-graph update connection --id 1 --from-device-id 2 --from-modelport-id 3 --to-device-id 4 --to-modelport-id 5`,
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
		fromDeviceId, _ := cmd.Flags().GetString("from-device-id")
		fromModelPortId, _ := cmd.Flags().GetString("from-modelport-id")
		toDeviceId, _ := cmd.Flags().GetString("to-device-id")
		toModelPortId, _ := cmd.Flags().GetString("to-modelport-id")
		allowVLANUnion, _ := cmd.Flags().GetBool("allow-vlan-union")

		// All connection parameters are required
		if fromDeviceId == "" || fromModelPortId == "" || toDeviceId == "" || toModelPortId == "" {
			fmt.Fprintf(os.Stderr, "All connection parameters are required: --from-device-id, --from-modelport-id, --to-device-id, --to-modelport-id.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the connection
		err = service.UpdateConnection(connectionId, fromDeviceId, fromModelPortId, toDeviceId, toModelPortId, allowVLANUnion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating connection: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated connection with ID %s\n", connectionId)
		fmt.Printf("  From device ID: %s, port ID: %s\n", fromDeviceId, fromModelPortId)
		fmt.Printf("  To device ID: %s, port ID: %s\n", toDeviceId, toModelPortId)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(ConnectionUpdateCmd)

	ConnectionUpdateCmd.Flags().String("id", "", "ID of the connection to update (required)")
	ConnectionUpdateCmd.Flags().String("from-device-id", "", "Source device ID (required)")
	ConnectionUpdateCmd.Flags().String("from-modelport-id", "", "Source model port ID (required)")
	ConnectionUpdateCmd.Flags().String("to-device-id", "", "Destination device ID (required)")
	ConnectionUpdateCmd.Flags().String("to-modelport-id", "", "Destination model port ID (required)")
	ConnectionUpdateCmd.Flags().Bool("allow-vlan-union", false, "Allow connection if VLANs have any overlap (default: strict matching)")
	// ConnectionUpdateCmd.MarkFlagRequired("id")
	// ConnectionUpdateCmd.MarkFlagRequired("from-device-id")
	// ConnectionUpdateCmd.MarkFlagRequired("from-modelport-id")
	// ConnectionUpdateCmd.MarkFlagRequired("to-device-id")
	// ConnectionUpdateCmd.MarkFlagRequired("to-modelport-id")
}