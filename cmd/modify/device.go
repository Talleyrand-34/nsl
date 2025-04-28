
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
package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// zoneModCmd represents the port command
var DeviceModCmd = &cobra.Command{
	Use:   "device",
	Short: "device modifications subcommand",
	Long:  `Specify a device which consists on a name`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"name", "model", "zoneid", "zonename", "proprietary"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		label := vals[0]
		model := vals[1]
		zoneid := vals[2]
		zonename := vals[3]
		proprietary := vals[4]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddDevice(label, model, zoneid, zonename, proprietary)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing zone: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(DeviceModCmd)

	DeviceModCmd.Flags().
		String("name", "", "Sets the name of the zone")
	DeviceModCmd.Flags().
		String("model", "", "Sets the model of the device by name")
	DeviceModCmd.Flags().
		String("zoneid", "", "Sets the zone by id if there is (prioritized over name)")
	DeviceModCmd.Flags().
		String("zonename", "", "Sets the zone by name of the zone")
	DeviceModCmd.Flags().
		String("proprietary", "", "Sets the proprietary of the zone")
}
