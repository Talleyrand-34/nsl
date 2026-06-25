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
package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// modelDeviceModCmd represents the port command
var modelPortModCmd = &cobra.Command{
	Use:   "modelport",
	Short: "Add a model port",
	Long: `Add a port to a device model, defining the model's port layout: a name, a
grid position (--posx/--posy), the type (wired by default, or 'wifi') and whether
the port allows multiple connections.

Example:
  nsl-graph add modelport --modelname "ISR4431" --name "GigE0/0/0" --posx 0 --posy 0`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"name", "posx", "posy", "modelname"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		name := vals[0]
		posx := vals[1]
		posy := vals[2]
		modelname := vals[3]
		allowMultiple, _ := cmd.Flags().GetBool("allow-multiple")
		portType, _ := cmd.Flags().GetString("port-type")
		band, _ := cmd.Flags().GetString("band")
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddModelPort(name, posx, posy, modelname, allowMultiple, portType, band)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing modelPort: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.AddCmd.AddCommand(modelPortModCmd)

	modelPortModCmd.Flags().
		String("name", "", "Sets the name of the modelPort")
	modelPortModCmd.Flags().
		String("posx", "", "Sets the position X of the modelPort")
	modelPortModCmd.Flags().
		String("posy", "", "Sets the position Y of the modelPort")
	modelPortModCmd.Flags().
		String("modelname", "", "Sets the model name of the modelPort")
	modelPortModCmd.Flags().
		Bool("allow-multiple", false, "Allow multiple connections to this port")
	modelPortModCmd.Flags().
		String("port-type", "", "Port type: '' (wired) or 'wifi' (radio)")
	modelPortModCmd.Flags().
		String("band", "", "WiFi band: '2.4GHz', '5GHz', or '6GHz' (wifi ports only)")
}
