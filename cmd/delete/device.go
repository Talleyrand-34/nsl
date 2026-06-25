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

// deviceDelCmd represents the device delete command
var deviceDelCmd = &cobra.Command{
	Use:   "device [device_id]",
	Short: "Delete a device",
	Long:  `Delete a device from the database by ID. Use --cascade to also delete all dependent device ports and connections.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceId := args[0]
		cascade, _ := cmd.Flags().GetBool("cascade")

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			return fmt.Errorf("error connecting to database: %w", err)
		}

		if cascade {
			err = service.DeleteDeviceCascade(deviceId)
			if err == nil {
				fmt.Printf("Device with ID '%s' and all dependencies deleted successfully\n", deviceId)
			}
		} else {
			err = service.DeleteDevice(deviceId)
			if err == nil {
				fmt.Printf("Device with ID '%s' deleted successfully\n", deviceId)
			}
		}

		if err != nil {
			return fmt.Errorf("error deleting device: %w", err)
		}
		return nil
	},
}

func init() {
	cmd_root.DeleteCmd.AddCommand(deviceDelCmd)
	deviceDelCmd.Flags().Bool("cascade", false, "Delete device and all dependent objects (device ports, connections)")
}
