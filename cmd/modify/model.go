
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

// modelDeviceModCmd represents the port command
var modelDeviceModCmd = &cobra.Command{
	Use:   "model",
	Short: "model modifications subcommand",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"name", "brand", "class"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		name := vals[0]
		brand := vals[1]
		class := vals[2]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddModel(name, brand, class)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing modelDevice: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(modelDeviceModCmd)

	modelDeviceModCmd.Flags().
		String("name", "", "Sets the name of the modelDevice")
	modelDeviceModCmd.Flags().
		String("brand", "", "Sets the brand associated of the modelDevice")
	modelDeviceModCmd.Flags().
		String("class", "", "Sets the deviceclass of the modelDevice")
}
