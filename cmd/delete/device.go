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
	"os"

	"github.com/spf13/cobra"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// deviceDelCmd represents the device delete command
var deviceDelCmd = &cobra.Command{
	Use:   "device [device_id]",
	Short: "Delete a device",
	Long:  `Delete a device from the database by ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceId := args[0]
		
		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		err = service.DeleteDevice(deviceId)
		if err != nil {
			fmt.Printf("Error deleting device: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Device with ID '%s' deleted successfully\n", deviceId)
	},
}

func init() {
	cmd_root.DeleteCmd.AddCommand(deviceDelCmd)
}