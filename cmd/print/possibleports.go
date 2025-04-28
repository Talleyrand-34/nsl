
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
package cmd_print

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// devicesCmd represents the devices command
var possiblePortPrintCmd = &cobra.Command{
	Use:   "possibleports",
	Short: "Especial: Print the possible ports which is the tuple deviceid+modelportid",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		deviceid, err := cmd.Flags().GetString("deviceid")
		if err != nil {

			fmt.Fprintf(os.Stderr, "Error reading flag '%v': %v\n", "all", err)
			os.Exit(1)
		}
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		if deviceid == "" {
			fmt.Println(string(service.GetAllPortsAll()))
		} else {
			fmt.Println(string(service.GetAllPortsDevice(deviceid)))
		}
	},
}

func init() {
	cmd.PrintCmd.AddCommand(possiblePortPrintCmd)
	possiblePortPrintCmd.Flags().
		String("deviceid", "", "Sets the name of the zone")
}
