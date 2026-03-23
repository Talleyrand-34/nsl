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

// deviceClassUpdateCmd represents the device class update command
var deviceClassUpdateCmd = &cobra.Command{
	Use:   "deviceclass",
	Short: "Update an existing device class",
	Long: `Update an existing device class by its ID.
	
A device class categorizes network equipment types (e.g., router, switch, firewall).

Example:
  nsl-graph update deviceclass --id 1 --name "Layer3Switch"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required flags
		deviceClassId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if deviceClassId == "" {
			fmt.Fprintf(os.Stderr, "Device class ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		newDeviceClassName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'name': %v\n", err)
			os.Exit(1)
		}
		if newDeviceClassName == "" {
			fmt.Fprintf(os.Stderr, "New device class name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the device class
		err = service.UpdateDeviceClass(deviceClassId, newDeviceClassName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating device class: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated device class with ID %s to '%s'\n", deviceClassId, newDeviceClassName)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(deviceClassUpdateCmd)

	deviceClassUpdateCmd.Flags().String("id", "", "ID of the device class to update (required)")
	deviceClassUpdateCmd.Flags().String("name", "", "New name for the device class (required)")
	deviceClassUpdateCmd.MarkFlagRequired("id")
	deviceClassUpdateCmd.MarkFlagRequired("name")
}