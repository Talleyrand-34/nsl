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
package cmd_delete

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

var deviceInterfaceDelCmd = &cobra.Command{
	Use:   "deviceinterface [interface_id]",
	Short: "Delete a device interface",
	Long:  `Delete a logical device interface by ID. Also removes all associated interface-port links.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			return fmt.Errorf("error connecting to database: %w", err)
		}

		err = service.DeleteDeviceInterface(id)
		if err != nil {
			return fmt.Errorf("error deleting device interface: %w", err)
		}

		fmt.Printf("Device interface '%s' deleted successfully\n", id)
		return nil
	},
}

var interfacePortDelCmd = &cobra.Command{
	Use:   "interfaceport [interface_id] [device_id] [model_port_id]",
	Short: "Remove an interface-port link",
	Long:  `Remove the association between a logical interface and a physical port.`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		interfaceID := args[0]
		deviceID := args[1]
		modelPortID := args[2]

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			return fmt.Errorf("error connecting to database: %w", err)
		}

		err = service.DeleteInterfacePort(interfaceID, deviceID, modelPortID)
		if err != nil {
			return fmt.Errorf("error deleting interface-port link: %w", err)
		}

		fmt.Printf("Interface-port link (interface='%s', device='%s', port='%s') deleted successfully\n", interfaceID, deviceID, modelPortID)
		return nil
	},
}

func init() {
	cmd_root.DeleteCmd.AddCommand(deviceInterfaceDelCmd)
	cmd_root.DeleteCmd.AddCommand(interfacePortDelCmd)
}
