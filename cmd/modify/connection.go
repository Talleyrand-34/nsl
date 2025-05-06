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
var connectionModCmd = &cobra.Command{
	Use:   "connection",
	Short: "connection modifications subcommand",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"from-device", "from-model-port-id", "to-device", "to-model-port-id"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		fromDevice := vals[0]
		fromModelPortId := vals[1]
		toDevice := vals[2]
		toModelPortId := vals[3]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddConnection(fromDevice, fromModelPortId, toDevice, toModelPortId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing connection: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(connectionModCmd)

	connectionModCmd.Flags().
		String("from-device", "", "Sets the fatherzone by name if there is")
	connectionModCmd.Flags().
		String("to-device", "", "Sets the fatherzone by id if there is")
	connectionModCmd.Flags().
		String("from-model-port-id", "", "Sets the name of the zone")
	connectionModCmd.Flags().
		String("to-model-port-id", "", "Sets the proprietary of the zone")
}
