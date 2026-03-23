
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
package cmd_print

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
)

// devicesCmd represents the devices command
var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Print the devices",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// p, _ := db.DeviceRepository.FetchDevices()
		// fmt.Println(p)
	},
}

func init() {
	cmd.PrintCmd.AddCommand(devicesCmd)

	devicesCmd.Flags().
		BoolP("all-model-ports", "a", false, "Print the device with all the ports, not only those used")
	devicesCmd.Flags().
		BoolP("info-ports", "i", false, "Print the device with the ports used")
	devicesCmd.Flags().
		BoolP("possible-ports", "p", false, "Print the cartesian product of the possible ports of the model and the device ports")
}
