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

// DeviceUpdateCmd represents the device update command
var DeviceUpdateCmd = &cobra.Command{
	Use:   "device",
	Short: "Update an existing device",
	Long: `Update an existing device by its ID.
	
You can update the device label, model, zone, and owner assignment.
All foreign key references should be provided as IDs.

Examples:
  nsl-graph update device --id 1 --label "New Device Name"
  nsl-graph update device --id 1 --label "Router-Core-01" --model-id 2 --zone-id 3 --owner-id 1`,
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
		newOwnerId, _ := cmd.Flags().GetString("owner-id")

		// is-unmanaged is a tri-state: only applied when the flag was set.
		var isUnmanaged *bool
		if cmd.Flags().Changed("is-unmanaged") {
			v, _ := cmd.Flags().GetBool("is-unmanaged")
			isUnmanaged = &v
		}

		// At least one field must be provided for update
		if newLabel == "" && newModelId == "" && newZoneId == "" && newOwnerId == "" && isUnmanaged == nil {
			fmt.Fprintf(os.Stderr, "At least one field to update is required (label, model-id, zone-id, owner-id, is-unmanaged).\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the device. Empty fields are left unchanged by the repository.
		err = service.UpdateDevice(deviceId, newLabel, newModelId, newZoneId, newOwnerId, isUnmanaged)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating device: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated device with ID %s\n", deviceId)
		if newLabel != "" {
			fmt.Printf("  New label: %s\n", newLabel)
		}
		if newModelId != "" {
			fmt.Printf("  New model ID: %s\n", newModelId)
		}
		if newZoneId != "" {
			fmt.Printf("  New zone ID: %s\n", newZoneId)
		}
		if newOwnerId != "" {
			fmt.Printf("  New owner ID: %s\n", newOwnerId)
		}
		if isUnmanaged != nil {
			fmt.Printf("  Unmanaged VLANs: %t\n", *isUnmanaged)
		}
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(DeviceUpdateCmd)

	DeviceUpdateCmd.Flags().String("id", "", "ID of the device to update (required)")
	DeviceUpdateCmd.Flags().String("label", "", "New label for the device")
	DeviceUpdateCmd.Flags().String("model-id", "", "New model ID for the device")
	DeviceUpdateCmd.Flags().String("zone-id", "", "New zone ID for the device (optional)")
	DeviceUpdateCmd.Flags().String("owner-id", "", "New owner ID for the device (optional)")
	DeviceUpdateCmd.Flags().Bool("is-unmanaged", false, "Mark device as unmanaged (replicates all VLANs through all ports); only applied when the flag is set")
	// DeviceUpdateCmd.MarkFlagRequired("id")
}
