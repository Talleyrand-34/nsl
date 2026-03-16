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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

var deviceInterfacePrintCmd = &cobra.Command{
	Use:   "deviceinterface",
	Short: "Print device interfaces",
	Long:  `Print all logical device interfaces. Use --deviceid to filter by device.`,
	Run: func(cmd *cobra.Command, args []string) {
		deviceid, _ := cmd.Flags().GetString("deviceid")

		service, err := util.ServiceConnection()
		if err != nil {
			return
		}

		var result interface{}
		if deviceid != "" {
			result, err = service.GetDeviceInterfaces(deviceid)
		} else {
			result, err = service.GetAllDeviceInterfaces()
		}
		if err != nil {
			fmt.Println("Error getting device interfaces:", err)
			return
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling to JSON:", err)
			return
		}
		fmt.Println(string(jsonBytes))
	},
}

var interfacePortPrintCmd = &cobra.Command{
	Use:   "interfaceport",
	Short: "Print interface-port links",
	Long: `Print interface-to-port associations.

Use --interfaceid to list ports for a given interface.
Use --deviceid and --modelportid to list interfaces for a given physical port.
Without flags, prints all interface-port links.`,
	Run: func(cmd *cobra.Command, args []string) {
		interfaceid, _ := cmd.Flags().GetString("interfaceid")
		deviceid, _ := cmd.Flags().GetString("deviceid")
		modelportid, _ := cmd.Flags().GetString("modelportid")

		service, err := util.ServiceConnection()
		if err != nil {
			return
		}

		var result interface{}
		if interfaceid != "" {
			result, err = service.GetPortsForInterface(interfaceid)
		} else if deviceid != "" && modelportid != "" {
			result, err = service.GetInterfacesForPort(deviceid, modelportid)
		} else {
			result, err = service.GetAllInterfacePorts()
		}
		if err != nil {
			fmt.Println("Error getting interface-port links:", err)
			return
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling to JSON:", err)
			return
		}
		fmt.Println(string(jsonBytes))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(deviceInterfacePrintCmd)
	deviceInterfacePrintCmd.Flags().String("deviceid", "", "Filter by device ID (optional)")

	cmd.PrintCmd.AddCommand(interfacePortPrintCmd)
	interfacePortPrintCmd.Flags().String("interfaceid", "", "Show ports for this interface ID")
	interfacePortPrintCmd.Flags().String("deviceid", "", "Device ID (use with --modelportid)")
	interfacePortPrintCmd.Flags().String("modelportid", "", "Model port ID (use with --deviceid)")
}
