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
	e "nsl-graph/internal/repository/entities"
)

// VerboseInterfacePort is an enriched InterfacePort with resolved names
type VerboseInterfacePort struct {
	ID            string             `json:"id"`
	InterfaceID   string             `json:"interface_id"`
	InterfaceName string             `json:"interface_name,omitempty"`
	DeviceID      string             `json:"device_id"`
	DeviceName    string             `json:"device_name,omitempty"`
	ModelPortID   string             `json:"model_port_id"`
	PortName      string             `json:"port_name,omitempty"`
	VlanConfigs   []e.PortVlanConfig `json:"vlan_configs,omitempty"`
	IPAddresses   []string           `json:"ip_addresses,omitempty"`
}

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
	Long: `Print interface-to-port associations with resolved names.

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

		var verbosePorts []VerboseInterfacePort

		// Get lookup maps
		devices, _ := service.GetDevices()
		modelPorts, _ := service.GetModelPorts()
		interfaces, _ := service.GetAllDeviceInterfaces()

		deviceNameMap := make(map[string]string)
		for _, d := range devices {
			deviceNameMap[d.ID] = d.Name
		}
		portNameMap := make(map[string]string)
		for _, mp := range modelPorts {
			portNameMap[mp.ID] = mp.Name
		}
		ifaceNameMap := make(map[string]string)
		// VLANs and IPs are owned by the DeviceInterface; resolve them by interface ID.
		ifaceByID := make(map[string]e.DeviceInterface)
		for _, iface := range interfaces {
			ifaceNameMap[iface.ID] = iface.Name
			ifaceByID[iface.ID] = iface
		}

		if interfaceid != "" {
			// GetPortsForInterface returns []DevicePort - get the physical ports for this interface
			devicePorts, err := service.GetPortsForInterface(interfaceid)
			if err != nil {
				fmt.Println("Error getting ports for interface:", err)
				return
			}
			verbosePorts = make([]VerboseInterfacePort, len(devicePorts))
			for i, dp := range devicePorts {
				verbosePorts[i] = VerboseInterfacePort{
					InterfaceID:   interfaceid,
					InterfaceName: ifaceNameMap[interfaceid],
					DeviceID:      dp.DeviceID,
					DeviceName:    deviceNameMap[dp.DeviceID],
					ModelPortID:   dp.ModelID,
					PortName:      portNameMap[dp.ModelID],
				}
			}
		} else if deviceid != "" && modelportid != "" {
			// GetInterfacesForPort returns []DeviceInterface - get interfaces on a physical port
			deviceIfaces, err := service.GetInterfacesForPort(deviceid, modelportid)
			if err != nil {
				fmt.Println("Error getting interfaces for port:", err)
				return
			}
			// List the join records for this physical port; VLANs/IPs come from
			// the linked DeviceInterface.
			allIPs, _ := service.GetAllInterfacePorts()
			ifaceNameMap2 := make(map[string]string)
			for _, di := range deviceIfaces {
				ifaceNameMap2[di.ID] = di.Name
			}
			verbosePorts = make([]VerboseInterfacePort, 0)
			for _, ip := range allIPs {
				if ip.DeviceID == deviceid && ip.ModelPortID == modelportid {
					iface := ifaceByID[ip.InterfaceID]
					verbosePorts = append(verbosePorts, VerboseInterfacePort{
						ID:            ip.ID,
						InterfaceID:   ip.InterfaceID,
						InterfaceName: ifaceNameMap2[ip.InterfaceID],
						DeviceID:      ip.DeviceID,
						DeviceName:    deviceNameMap[ip.DeviceID],
						ModelPortID:   ip.ModelPortID,
						PortName:      portNameMap[ip.ModelPortID],
						VlanConfigs:   iface.VlanConfigs,
						IPAddresses:   iface.IPAddresses,
					})
				}
			}
		} else {
			// Get all interface ports
			ports, err := service.GetAllInterfacePorts()
			if err != nil {
				fmt.Println("Error getting interface ports:", err)
				return
			}
			verbosePorts = make([]VerboseInterfacePort, len(ports))
			for i, p := range ports {
				iface := ifaceByID[p.InterfaceID]
				verbosePorts[i] = VerboseInterfacePort{
					ID:            p.ID,
					InterfaceID:   p.InterfaceID,
					InterfaceName: ifaceNameMap[p.InterfaceID],
					DeviceID:      p.DeviceID,
					DeviceName:    deviceNameMap[p.DeviceID],
					ModelPortID:   p.ModelPortID,
					PortName:      portNameMap[p.ModelPortID],
					VlanConfigs:   iface.VlanConfigs,
					IPAddresses:   iface.IPAddresses,
				}
			}
		}

		jsonBytes, err := json.MarshalIndent(verbosePorts, "", "  ")
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
