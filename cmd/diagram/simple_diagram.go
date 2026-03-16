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
package cmd_diagram

import (
	"nsl-graph/internal/format"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// rootCmd represents the base command when called without any subcommands
var diagramSimpleAllPorts bool

var DiagramSimpleCmd = &cobra.Command{
	Use:   "port",
	Short: "Generates a diagram from an nsl especification with the focus on ports",
	Long:  `.`,
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
		devicePorts, err := service.GetDevicePorts()
		allInterfaces, err := service.GetAllDeviceInterfaces()
		ifacePorts, err := service.GetAllInterfacePorts()
		if err != nil {
			return
		}
		d2diagram := format.GenerateD2FocusPorts(devices, connections, zones, devicePorts, allInterfaces, ifacePorts, false, diagramSimpleAllPorts)
		format.WriteDiagram(d2diagram, op, of, oi)
	},
}

func init() {
	cmd.DiagramCmd.AddCommand(DiagramSimpleCmd)
	DiagramSimpleCmd.Flags().BoolVar(&diagramSimpleAllPorts, "all-ports", false, "Include ports with no connections in the diagram")
}
