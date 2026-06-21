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
package cmd_print

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// DeviceWithInterfaces represents a device with its interfaces
type DeviceWithInterfaces struct {
	ID          string            `json:"id"`
	Name        string            `json:"label"`
	Model       string            `json:"model"`
	Brand       string            `json:"brand"`
	ZoneID      string            `json:"zoneid"`
	ZoneName    string            `json:"zonename"`
	ZoneFather  string            `json:"zonefathername"`
	Proprietary string            `json:"proprietary"`
	Interfaces  []InterfaceNameIP `json:"interfaces"`
}

// InterfaceNameIP represents an interface with its IP address
type InterfaceNameIP struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

// devicesCmd represents the devices command
var devicePrintCmd = &cobra.Command{
	Use:   "device",
	Short: "Print the devices",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		devs, err := service.GetDevices()
		if err != nil {
			fmt.Println("Error getting Devices:", err)
			return
		}

		// Build result with interfaces
		result := make([]DeviceWithInterfaces, len(devs))
		for i, d := range devs {
			result[i] = DeviceWithInterfaces{
				ID:          d.ID,
				Name:        d.Name,
				Model:       d.Model,
				Brand:       d.Brand,
				ZoneID:      d.ZoneID,
				ZoneName:    d.ZoneName,
				ZoneFather:  d.ZoneFather,
				Proprietary: d.Proprietary,
				Interfaces:  []InterfaceNameIP{},
			}

			// Get interfaces for this device
			ifaces, err := service.GetDeviceInterfaces(d.ID)
			if err != nil {
				continue
			}
			for _, iface := range ifaces {
				// IPs are owned by the DeviceInterface.
				var ip string
				if len(iface.IPAddresses) > 0 {
					ip = iface.IPAddresses[0]
				}
				result[i].Interfaces = append(result[i].Interfaces, InterfaceNameIP{
					Name: iface.Name,
					IP:   ip,
				})
			}
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling Devices to JSON:", err)
			return
		}
		fmt.Println(string(jsonBytes))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(devicePrintCmd)
}
