// SPDX-License-Identifier: AGPL-3.0-or-later
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

// DevicePortUpdateCmd represents the "update deviceport" command.
var DevicePortUpdateCmd = &cobra.Command{
	Use:   "deviceport",
	Short: "Update an existing device port",
	Long: `Update an existing device port's MAC address and/or VLAN configs, identified
by its device ID and model-port ID. Only the fields you provide are changed.

VLAN configs accept '100:tagged,200:untagged' (or just '100,200' for tagged).

Example:
  nsl-graph update deviceport --deviceid <id> --modelportid <id> --vlan-configs 100:tagged`,
	Run: func(cmd *cobra.Command, args []string) {
		deviceid, _ := cmd.Flags().GetString("deviceid")
		modelportid, _ := cmd.Flags().GetString("modelportid")
		if deviceid == "" || modelportid == "" {
			fmt.Fprintf(os.Stderr, "Both --deviceid and --modelportid are required.\n")
			os.Exit(1)
		}
		macaddress, _ := cmd.Flags().GetString("macaddress")
		vlanConfigsStr, _ := cmd.Flags().GetStringSlice("vlan-configs")
		vlanConfigs := util.ParseVlanConfigs(vlanConfigsStr)

		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}
		if err := service.UpdateDevicePort(deviceid, modelportid, macaddress, vlanConfigs); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating device port: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated deviceport %s:%s\n", deviceid, modelportid)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(DevicePortUpdateCmd)
	DevicePortUpdateCmd.Flags().String("deviceid", "", "Device ID (required)")
	DevicePortUpdateCmd.Flags().String("modelportid", "", "Model port ID (required)")
	DevicePortUpdateCmd.Flags().String("macaddress", "", "New MAC address (optional)")
	DevicePortUpdateCmd.Flags().StringSlice("vlan-configs", []string{}, "VLAN configs (format: '100:tagged,200:untagged' or '100,200' for tagged)")
}
