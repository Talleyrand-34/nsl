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
package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// DeviceModCmd represents the device creation command
var DeviceModCmd = &cobra.Command{
	Use:   "device",
	Short: "Create a new network device",
	Long: `Create a new network device with a specified model, zone, and ownership.
	
A device is an instance of a model (like a specific router or switch) deployed in a particular zone.
You can specify the zone either by ID (--zone-id) or by name (--zone-name).

Examples:
  nsl-graph modify device --label "Router-01" --model "ISR4431" --zone-name "DataCenter" --proprietary "IT Department"
  nsl-graph modify device --label "Switch-Core-01" --model "Catalyst2960" --zone-id 1 --proprietary "Network Team"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get required device label
		deviceLabel, err := cmd.Flags().GetString("label")
		if err != nil || deviceLabel == "" {
			fmt.Fprintf(os.Stderr, "Device label is required. Use --label flag.\n")
			os.Exit(1)
		}

		// Get required model name
		modelName, err := cmd.Flags().GetString("model")
		if err != nil || modelName == "" {
			fmt.Fprintf(os.Stderr, "Model name is required. Use --model flag.\n")
			os.Exit(1)
		}

		// Get zone information (either by ID or name)
		zoneId, _ := cmd.Flags().GetString("zone-id")
		zoneName, _ := cmd.Flags().GetString("zone-name")

		if zoneId == "" && zoneName == "" {
			fmt.Fprintf(os.Stderr, "Zone is required. Use either --zone-id or --zone-name flag.\n")
			os.Exit(1)
		}

		// Get required proprietary
		proprietaryName, err := cmd.Flags().GetString("proprietary")
		if err != nil || proprietaryName == "" {
			fmt.Fprintf(os.Stderr, "Proprietary owner is required. Use --proprietary flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Create the device
		isUnmanaged, _ := cmd.Flags().GetBool("is-unmanaged")
		isInvisible, _ := cmd.Flags().GetBool("is-invisible")
		err = service.AddDevice(deviceLabel, modelName, zoneId, zoneName, proprietaryName, isUnmanaged, isInvisible)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating device: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created device '%s' with model '%s'", deviceLabel, modelName)
		if zoneId != "" {
			fmt.Printf(" in zone ID %s", zoneId)
		} else {
			fmt.Printf(" in zone '%s'", zoneName)
		}
		fmt.Printf(" owned by '%s'\n", proprietaryName)
		if isUnmanaged {
			fmt.Printf("  Device is marked as unmanaged\n")
		}
		if isInvisible {
			fmt.Printf("  Device is marked as invisible\n")
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(DeviceModCmd)

	DeviceModCmd.Flags().String("label", "", "Device identifier/name (required)")
	DeviceModCmd.Flags().String("model", "", "Model name for the device (required)")
	DeviceModCmd.Flags().String("zone-id", "", "Zone ID where device will be deployed")
	DeviceModCmd.Flags().String("zone-name", "", "Zone name where device will be deployed")
	DeviceModCmd.Flags().String("proprietary", "", "Proprietary owner of the device (required)")
	DeviceModCmd.Flags().Bool("is-unmanaged", false, "Device is unmanaged (replicates all VLANs through all ports)")
	DeviceModCmd.Flags().Bool("is-invisible", false, "Device is invisible in network scope (no IP)")

	DeviceModCmd.MarkFlagRequired("label")
	DeviceModCmd.MarkFlagRequired("model")
	DeviceModCmd.MarkFlagRequired("proprietary")
}
