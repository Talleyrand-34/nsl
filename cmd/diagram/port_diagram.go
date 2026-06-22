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
package cmd_diagram

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
	"nsl-graph/internal/format"
)

var (
	diagramPortAllPorts    bool
	diagramPortVlan        bool
	diagramPortVlanScope   string
	diagramPortColorTarget string
)

// DiagramPortCmd represents the "diagram port" command.
var DiagramPortCmd = &cobra.Command{
	Use:   "port",
	Short: "Generate a port-focused diagram (use --vlan for VLAN colors)",
	Long: `Generate a network diagram focused on device ports (lists all ports per device).

With --vlan the ports are colored per VLAN; --vlan-scope and --color-target tune
the coloring.`,
	Run: func(cmd *cobra.Command, args []string) {
		vals := util.Flagproc(cmd, []string{"outPath", "outFile", "outImage"})
		op, of, oi := vals[0], vals[1], vals[2]
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
		var d2diagram string
		if diagramPortVlan {
			d2diagram = format.GenerateD2FocusPortsWithVlans(devices, connections, zones, devicePorts, allInterfaces, ifacePorts, diagramPortAllPorts, diagramPortVlanScope, diagramPortColorTarget)
		} else {
			d2diagram = format.GenerateD2FocusPorts(devices, connections, zones, devicePorts, allInterfaces, ifacePorts, false, diagramPortAllPorts)
		}
		format.WriteDiagram(d2diagram, op, of, oi)
	},
}

func init() {
	cmd.DiagramCmd.AddCommand(DiagramPortCmd)
	f := DiagramPortCmd.Flags()
	f.BoolVar(&diagramPortAllPorts, "all-ports", false, "Include ports with no connections in the diagram")
	f.BoolVar(&diagramPortVlan, "vlan", false, "Color ports by VLAN")
	f.StringVar(&diagramPortVlanScope, "vlan-scope", "untagged", "With --vlan: 'untagged' or 'all'")
	f.StringVar(&diagramPortColorTarget, "color-target", "both", "With --vlan: what to color — 'both', 'connections', or 'ports'")
}
