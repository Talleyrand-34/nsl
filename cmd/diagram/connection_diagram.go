// SPDX-License-Identifier: AGPL-3.0-or-later
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
	diagramConnectionAllPorts    bool
	diagramConnectionVlan        bool
	diagramConnectionVlanScope   string
	diagramConnectionColorTarget string
)

// DiagramConnectionCmd represents the "diagram connection" command.
var DiagramConnectionCmd = &cobra.Command{
	Use:   "connection",
	Short: "Generate a connection-focused diagram (use --vlan for VLAN colors)",
	Long: `Generate a network diagram focused on connections.

With --vlan the connections are colored per VLAN; --vlan-scope and --color-target
tune the coloring.`,
	Run: func(cmd *cobra.Command, args []string) {
		vals := util.Flagproc(cmd, []string{"outPath", "outFile", "outImage"})
		op, of, oi := vals[0], vals[1], vals[2]
		ascii, _ := cmd.Flags().GetBool("ascii")
		unicode := mustCharsetUnicode(cmd)
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
		if diagramConnectionVlan {
			d2diagram = format.GenerateD2FocusConnectionsWithVlans(devices, connections, zones, devicePorts, allInterfaces, ifacePorts, diagramConnectionAllPorts, diagramConnectionVlanScope, diagramConnectionColorTarget)
		} else {
			d2diagram = format.GenerateD2FocusConnections(devices, connections, zones, devicePorts, allInterfaces, ifacePorts, false, diagramConnectionAllPorts)
		}
		format.WriteDiagram(d2diagram, op, of, oi, ascii, unicode)
	},
}

// mustCharsetUnicode reads the shared --charset flag and reports whether the
// unicode (box-drawing) charset was requested. Unknown values default to unicode.
func mustCharsetUnicode(c *cobra.Command) bool {
	cs, _ := c.Flags().GetString("charset")
	return cs != "ascii"
}

func init() {
	cmd.DiagramCmd.AddCommand(DiagramConnectionCmd)
	f := DiagramConnectionCmd.Flags()
	f.BoolVar(&diagramConnectionAllPorts, "all-ports", false, "Include ports with no connections in the diagram")
	f.BoolVar(&diagramConnectionVlan, "vlan", false, "Color connections by VLAN")
	f.StringVar(&diagramConnectionVlanScope, "vlan-scope", "untagged", "With --vlan: 'untagged' (one line per link) or 'all' (one line per VLAN in the intersection)")
	f.StringVar(&diagramConnectionColorTarget, "color-target", "both", "With --vlan: what to color — 'both', 'connections', or 'ports'")
}
