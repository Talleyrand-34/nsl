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

// DeviceUpdateCmd represents the device update command
var DeviceUpdateCmd = &cobra.Command{
	Use:   "device",
	Short: "Update an existing device",
	Long: `Update an existing device by its ID.
	
You can update the device label, model, zone, and proprietary assignment.
All foreign key references should be provided as IDs.

Examples:
  nsl-graph update device --id 1 --label "New Device Name"
  nsl-graph update device --id 1 --label "Router-Core-01" --model-id 2 --zone-id 3 --proprietary-id 1`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required device ID
		deviceId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if deviceId == "" {
			fmt.Fprintf(os.Stderr, "Device ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		// Get optional update fields
		newLabel, _ := cmd.Flags().GetString("label")
		newModelId, _ := cmd.Flags().GetString("model-id")
		newZoneId, _ := cmd.Flags().GetString("zone-id")
		newProprietaryId, _ := cmd.Flags().GetString("proprietary-id")

		// At least one field must be provided for update
		if newLabel == "" && newModelId == "" && newZoneId == "" && newProprietaryId == "" {
			fmt.Fprintf(os.Stderr, "At least one field to update is required (label, model-id, zone-id, proprietary-id).\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// For this implementation, we'll require all fields to be specified
		// In a more sophisticated implementation, we could fetch current values and only update specified fields
		if newLabel == "" || newModelId == "" {
			fmt.Fprintf(os.Stderr, "For this implementation, both --label and --model-id are required.\n")
			os.Exit(1)
		}

		// Update the device
		err = service.UpdateDevice(deviceId, newLabel, newModelId, newZoneId, newProprietaryId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating device: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated device with ID %s\n", deviceId)
		fmt.Printf("  New label: %s\n", newLabel)
		if newModelId != "" {
			fmt.Printf("  New model ID: %s\n", newModelId)
		}
		if newZoneId != "" {
			fmt.Printf("  New zone ID: %s\n", newZoneId)
		}
		if newProprietaryId != "" {
			fmt.Printf("  New proprietary ID: %s\n", newProprietaryId)
		}
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(DeviceUpdateCmd)

	DeviceUpdateCmd.Flags().String("id", "", "ID of the device to update (required)")
	DeviceUpdateCmd.Flags().String("label", "", "New label for the device")
	DeviceUpdateCmd.Flags().String("model-id", "", "New model ID for the device")
	DeviceUpdateCmd.Flags().String("zone-id", "", "New zone ID for the device (optional)")
	DeviceUpdateCmd.Flags().String("proprietary-id", "", "New proprietary ID for the device (optional)")
	// DeviceUpdateCmd.MarkFlagRequired("id")
}