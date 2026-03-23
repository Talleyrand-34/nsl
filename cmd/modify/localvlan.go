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

// localvlanModCmd represents the localvlan command
var localvlanModCmd = &cobra.Command{
	Use:   "localvlan",
	Short: "Create a new local VLAN mapping",
	Long: `Create a local VLAN mapping by specifying a VLAN ID, device ID, and local VLAN name.

Local VLANs capture device-specific names for VLANs, allowing each device to have
its own local naming while maintaining global VLAN standardization.

Example:
  nsl-graph modify localvlan --vlan-id 100 --device-id abc123 --vlan-name "Local_Mgmt"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get required VLAN ID
		vlanID, err := cmd.Flags().GetString("vlan-id")
		if err != nil || vlanID == "" {
			fmt.Fprintf(os.Stderr, "VLAN ID is required. Use --vlan-id flag.\n")
			os.Exit(1)
		}

		// Get required device ID
		deviceID, err := cmd.Flags().GetString("device-id")
		if err != nil || deviceID == "" {
			fmt.Fprintf(os.Stderr, "Device ID is required. Use --device-id flag.\n")
			os.Exit(1)
		}

		// Get required local VLAN name
		vlanName, err := cmd.Flags().GetString("vlan-name")
		if err != nil || vlanName == "" {
			fmt.Fprintf(os.Stderr, "Local VLAN name is required. Use --vlan-name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Create the local VLAN mapping
		err = service.AddLocalVlan(vlanID, deviceID, vlanName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating local VLAN mapping: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created local VLAN mapping: VLAN %s on device %s (%s)\n", vlanID, deviceID, vlanName)
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(localvlanModCmd)

	localvlanModCmd.Flags().String("vlan-id", "", "VLAN ID (required, e.g., 100)")
	localvlanModCmd.Flags().String("device-id", "", "Device ID (required, database ID of the device)")
	localvlanModCmd.Flags().String("vlan-name", "", "Local VLAN name (required, e.g., 'Local_Mgmt')")

	localvlanModCmd.MarkFlagRequired("vlan-id")
	localvlanModCmd.MarkFlagRequired("device-id")
	localvlanModCmd.MarkFlagRequired("vlan-name")
}