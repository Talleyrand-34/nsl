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
package cmd_delete

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

var devicePortDelCmd = &cobra.Command{
	Use:   "deviceport [device_id] [model_port_id]",
	Short: "Delete a device port",
	Long:  `Delete a device port from the database by device ID and model port ID. Use --cascade to also delete all dependent connections.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceId := args[0]
		modelPortId := args[1]
		cascade, _ := cmd.Flags().GetBool("cascade")

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			return fmt.Errorf("error connecting to database: %w", err)
		}

		if cascade {
			err = service.DeleteDevicePortCascade(deviceId, modelPortId)
			if err == nil {
				fmt.Printf("Device port (device='%s', port='%s') and all dependencies deleted successfully\n", deviceId, modelPortId)
			}
		} else {
			err = service.DeleteDevicePort(deviceId, modelPortId)
			if err == nil {
				fmt.Printf("Device port (device='%s', port='%s') deleted successfully\n", deviceId, modelPortId)
			}
		}

		if err != nil {
			return fmt.Errorf("error deleting device port: %w", err)
		}
		return nil
	},
}

func init() {
	cmd_root.DeleteCmd.AddCommand(devicePortDelCmd)
	devicePortDelCmd.Flags().Bool("cascade", false, "Delete device port and all dependent objects (connections)")
}
