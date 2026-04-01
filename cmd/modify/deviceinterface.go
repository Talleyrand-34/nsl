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
	"strings"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
	e "nsl-graph/internal/repository/entities"
)

var deviceInterfaceModCmd = &cobra.Command{
	Use:   "deviceinterface",
	Short: "Add a logical interface to a device",
	Long: `Add a named logical interface to a device.

A device interface carries VLAN configuration and IP addresses.
It is linked to one or more physical device ports via interfaceport.

VLAN config format: VLAN_NUMBER:tagged|untagged (e.g. "10:untagged" or "20:tagged").
Multiple values accepted: --vlan-configs 10:untagged --vlan-configs 20:tagged`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"deviceid", "name", "description"}
		vals := util.Flagproc(cmd, flagNames)
		deviceid := vals[0]
		name := vals[1]
		description := vals[2]

		vlanConfigsStr, _ := cmd.Flags().GetStringSlice("vlan-configs")
		ipAddresses, _ := cmd.Flags().GetStringSlice("ips")
		ssid, _ := cmd.Flags().GetString("ssid")
		securityMode, _ := cmd.Flags().GetString("security-mode")

		var vlanConfigs []e.PortVlanConfig
		for _, vcStr := range vlanConfigsStr {
			if vcStr == "" {
				continue
			}
			parts := strings.Split(vcStr, ":")
			vlanNum := parts[0]
			tagged := true
			if len(parts) == 2 {
				tagged = strings.ToLower(parts[1]) != "untagged"
			}
			vlanConfigs = append(vlanConfigs, e.PortVlanConfig{
				VlanNumber: vlanNum,
				Tagged:     tagged,
			})
		}

		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddDeviceInterface(deviceid, name, description, "", vlanConfigs, ipAddresses, ssid, securityMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding device interface: %v\n", err)
			os.Exit(1)
		}
	},
}

var interfacePortModCmd = &cobra.Command{
	Use:   "interfaceport",
	Short: "Link a device interface to a physical port",
	Long:  `Create an association between a logical device interface and a physical device port.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"interfaceid", "deviceid", "modelportid"}
		vals := util.Flagproc(cmd, flagNames)
		interfaceid := vals[0]
		deviceid := vals[1]
		modelportid := vals[2]

		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddInterfacePort(interfaceid, deviceid, modelportid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding interface-port link: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(deviceInterfaceModCmd)
	deviceInterfaceModCmd.Flags().String("deviceid", "", "Device ID")
	deviceInterfaceModCmd.Flags().String("name", "", "Interface name (e.g. lan, vlan10, eth0.10)")
	deviceInterfaceModCmd.Flags().String("description", "", "Interface description (optional)")
	deviceInterfaceModCmd.Flags().StringSlice("vlan-configs", []string{}, "VLAN configs (format: 'NUMBER:tagged' or 'NUMBER:untagged')")
	deviceInterfaceModCmd.Flags().StringSlice("ips", []string{}, "IP addresses assigned to this interface")
	deviceInterfaceModCmd.Flags().String("ssid", "", "WiFi SSID (for wireless interfaces)")
	deviceInterfaceModCmd.Flags().String("security-mode", "", "WiFi security mode: open, wpa2, or wpa3")

	cmd.ModifyCmd.AddCommand(interfacePortModCmd)
	interfacePortModCmd.Flags().String("interfaceid", "", "Device interface ID")
	interfacePortModCmd.Flags().String("deviceid", "", "Device ID")
	interfacePortModCmd.Flags().String("modelportid", "", "Model port ID")
}
