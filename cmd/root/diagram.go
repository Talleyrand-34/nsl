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
package cmd_root

import (
	"nsl-graph/internal/format"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd"
	util "nsl-graph/cmd/utils"
)

// rootCmd represents the base command when called without any subcommands
var DiagramCmd = &cobra.Command{
	Use:   "diagram",
	Short: "Generates a diagram from an nsl especification",
	Long:  `.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"outPath", "outFile", "outImage"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		op := vals[0]
		of := vals[1]
		oi := vals[2]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		connections, err := service.GetConnections()
		devices, err := service.GetDevices()
		zones, err := service.GetZones()
		d2diagram := format.GenerateD2FromStruct(devices, connections, zones)
		format.WriteDiagram(d2diagram, op, of, oi)
	},
}

func init() {
	cmd.RootCmd.AddCommand(DiagramCmd)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.modtest.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
