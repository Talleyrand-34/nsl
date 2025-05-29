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
package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// connectionModCmd represents the connection creation command
var connectionModCmd = &cobra.Command{
	Use:   "connection",
	Short: "Create a new network connection between two device ports",
	Long: `Create a connection between two device ports by specifying device IDs and model port IDs.
	
A connection represents a physical link between two network devices through their specific ports.
All device and port IDs must exist in the database before creating the connection.

Example:
  nsl-graph modify connection --from-device-id 1 --from-modelport-id 2 --to-device-id 3 --to-modelport-id 4`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get required parameters with meaningful names
		fromDeviceId, err := cmd.Flags().GetString("from-device-id")
		if err != nil || fromDeviceId == "" {
			fmt.Fprintf(os.Stderr, "Source device ID is required. Use --from-device-id flag.\n")
			os.Exit(1)
		}

		fromModelPortId, err := cmd.Flags().GetString("from-modelport-id")
		if err != nil || fromModelPortId == "" {
			fmt.Fprintf(os.Stderr, "Source model port ID is required. Use --from-modelport-id flag.\n")
			os.Exit(1)
		}

		toDeviceId, err := cmd.Flags().GetString("to-device-id")
		if err != nil || toDeviceId == "" {
			fmt.Fprintf(os.Stderr, "Destination device ID is required. Use --to-device-id flag.\n")
			os.Exit(1)
		}

		toModelPortId, err := cmd.Flags().GetString("to-modelport-id")
		if err != nil || toModelPortId == "" {
			fmt.Fprintf(os.Stderr, "Destination model port ID is required. Use --to-modelport-id flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Create the connection
		err = service.AddConnection(fromDeviceId, fromModelPortId, toDeviceId, toModelPortId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating connection: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created connection from device %s (port %s) to device %s (port %s)\n", 
			fromDeviceId, fromModelPortId, toDeviceId, toModelPortId)
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(connectionModCmd)

	connectionModCmd.Flags().String("from-device-id", "", "ID of the source device (required)")
	connectionModCmd.Flags().String("from-modelport-id", "", "ID of the source device's model port (required)")
	connectionModCmd.Flags().String("to-device-id", "", "ID of the destination device (required)")
	connectionModCmd.Flags().String("to-modelport-id", "", "ID of the destination device's model port (required)")
	
	connectionModCmd.MarkFlagRequired("from-device-id")
	connectionModCmd.MarkFlagRequired("from-modelport-id")
	connectionModCmd.MarkFlagRequired("to-device-id")
	connectionModCmd.MarkFlagRequired("to-modelport-id")
}
