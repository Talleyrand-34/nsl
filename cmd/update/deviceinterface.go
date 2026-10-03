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

// DeviceInterfaceUpdateCmd represents the "update deviceinterface" command.
var DeviceInterfaceUpdateCmd = &cobra.Command{
	Use:   "deviceinterface",
	Short: "Update an existing device interface",
	Long: `Update a device interface's VLAN configurations, identified by its ID.

VLAN configs accept '10:untagged,20:tagged' (or just '10,20' for tagged).

Example:
  nsl-graph update deviceinterface --id <id> --vlan-configs 10:untagged,20:tagged`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		if id == "" {
			fmt.Fprintf(os.Stderr, "Interface ID is required. Use --id.\n")
			os.Exit(1)
		}
		vlanConfigsStr, _ := cmd.Flags().GetStringSlice("vlan-configs")
		vlanConfigs := util.ParseVlanConfigs(vlanConfigsStr)

		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}
		if err := service.UpdateDeviceInterface(id, vlanConfigs); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating device interface: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated device interface %s\n", id)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(DeviceInterfaceUpdateCmd)
	DeviceInterfaceUpdateCmd.Flags().String("id", "", "ID of the device interface to update (required)")
	DeviceInterfaceUpdateCmd.Flags().StringSlice("vlan-configs", []string{}, "VLAN configs (format: '10:untagged,20:tagged' or '10,20' for tagged)")
}
