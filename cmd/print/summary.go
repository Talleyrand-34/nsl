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
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
	e "nsl-graph/internal/repository/entities"
)

// summaryCmd represents the summary command
var summaryPrintCmd = &cobra.Command{
	Use:   "summary",
	Short: "Print network summary with different detail levels",
	Long: `Print network summary with progressive detail levels:
  Level 1: Basic inventory (device counts, zones, connections, VLANs)
  Level 2: Device details (device list, brand/model stats, zone hierarchy)
  Level 3: Network topology (connections, VLAN assignments, port details)`,
	Run: func(cmd *cobra.Command, args []string) {
		level, _ := cmd.Flags().GetInt("level")
		if level < 1 || level > 3 {
			fmt.Println("Error: Level must be between 1 and 3")
			return
		}

		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Println("Error connecting to service:", err)
			return
		}

		// Get all data needed for any level
		devices, err := service.GetDevices()
		if err != nil {
			fmt.Println("Error getting devices:", err)
			return
		}

		zones, err := service.GetZones()
		if err != nil {
			fmt.Println("Error getting zones:", err)
			return
		}

		connections, err := service.GetConnections()
		if err != nil {
			fmt.Println("Error getting connections:", err)
			return
		}

		vlans, err := service.GetVlans()
		if err != nil {
			fmt.Println("Error getting VLANs:", err)
			return
		}

		var devicePorts []e.DevicePort
		if level >= 3 {
			devicePorts, err = service.GetDevicePorts()
			if err != nil {
				fmt.Println("Error getting device ports:", err)
				return
			}
		}

		// Generate summary based on level
		switch level {
		case 1:
			printLevel1Summary(devices, zones, connections, vlans)
		case 2:
			printLevel1Summary(devices, zones, connections, vlans)
			fmt.Println()
			printLevel2Summary(devices, zones)
		case 3:
			printLevel1Summary(devices, zones, connections, vlans)
			fmt.Println()
			printLevel2Summary(devices, zones)
			fmt.Println()
			printLevel3Summary(connections, devicePorts)
		}
	},
}

func printLevel1Summary(devices []e.Device, zones []e.Zone, connections []e.Connection, vlans []e.Vlan) {
	fmt.Println("=== NETWORK SUMMARY - LEVEL 1: BASIC INVENTORY ===")

	// Device counts by class
	classCounts := make(map[string]int)
	for _, device := range devices {
		// Get device class from model (simplified - could be enhanced)
		class := getDeviceClassFromModel(device.Model)
		classCounts[class]++
	}

	fmt.Printf("\nDevices by Type (%d total):\n", len(devices))
	for class, count := range classCounts {
		fmt.Printf("  %-15s: %d\n", class, count)
	}

	// Zone summary
	zoneCounts := make(map[string]int)
	for _, device := range devices {
		zoneCounts[device.ZoneName]++
	}

	fmt.Printf("\nDevices by Zone (%d zones):\n", len(zones))
	for zoneName, count := range zoneCounts {
		fmt.Printf("  %-15s: %d devices\n", zoneName, count)
	}

	// Connection and VLAN counts
	fmt.Printf("\nNetwork Elements:\n")
	fmt.Printf("  Connections    : %d\n", len(connections))
	fmt.Printf("  VLANs          : %d\n", len(vlans))
}

func printLevel2Summary(devices []e.Device, zones []e.Zone) {
	fmt.Println("=== LEVEL 2: DEVICE DETAILS ===")

	// Brand/Model distribution
	brandCounts := make(map[string]int)
	modelCounts := make(map[string]int)
	for _, device := range devices {
		brandCounts[device.Brand]++
		modelCounts[device.Model]++
	}

	fmt.Printf("\nBrand Distribution:\n")
	for brand, count := range brandCounts {
		fmt.Printf("  %-15s: %d devices\n", brand, count)
	}

	fmt.Printf("\nModel Distribution:\n")
	for model, count := range modelCounts {
		fmt.Printf("  %-20s: %d\n", model, count)
	}

	// Device list grouped by zone
	zoneDevices := make(map[string][]e.Device)
	for _, device := range devices {
		zoneDevices[device.ZoneName] = append(zoneDevices[device.ZoneName], device)
	}

	fmt.Printf("\nDevices by Zone:\n")
	for zoneName, zoneDeviceList := range zoneDevices {
		fmt.Printf("\n  Zone: %s\n", zoneName)
		for _, device := range zoneDeviceList {
			ips := strings.Join(device.IPs, ", ")
			if ips == "" {
				ips = "none"
			}
			fmt.Printf("    %-20s [%s/%s] IPs: %s\n",
				device.Name, device.Brand, device.Model, ips)
		}
	}
}

func printLevel3Summary(connections []e.Connection, devicePorts []e.DevicePort) {
	fmt.Println("=== LEVEL 3: NETWORK TOPOLOGY ===")

	// Connection details
	fmt.Printf("\nConnections (%d total):\n", len(connections))
	for i, conn := range connections {
		fmt.Printf("  %d. %s[%s] ←→ %s[%s]\n", i+1,
			conn.FromDevice, conn.FromModelPort,
			conn.ToDevice, conn.ToModelPort)
	}

	if len(devicePorts) > 0 {
		// Group device ports by device
		devicePortMap := make(map[string][]e.DevicePort)
		for _, port := range devicePorts {
			devicePortMap[port.DevLabel] = append(devicePortMap[port.DevLabel], port)
		}

		fmt.Printf("\nDevice Ports and VLANs:\n")
		for deviceName, ports := range devicePortMap {
			fmt.Printf("\n  Device: %s\n", deviceName)

			// Sort ports by name for consistent output
			sort.Slice(ports, func(i, j int) bool {
				return ports[i].PortName < ports[j].PortName
			})

			for _, port := range ports {
				macAddr := port.MacAddress
				if macAddr == "" {
					macAddr = "none"
				}

				fmt.Printf("    %-15s MAC: %-17s", port.PortName, macAddr)
				fmt.Println()
			}
		}
	}
}

// Helper function to determine device class from model name
func getDeviceClassFromModel(model string) string {
	model = strings.ToLower(model)
	switch {
	case strings.Contains(model, "switch"):
		return "Switch"
	case strings.Contains(model, "router"):
		return "Router"
	case strings.Contains(model, "server"):
		return "Server"
	case strings.Contains(model, "printer"):
		return "Printer"
	case strings.Contains(model, "firewall") || strings.Contains(model, "fortigate"):
		return "Firewall"
	case strings.Contains(model, "access point") || strings.Contains(model, "unifi") || strings.Contains(model, "aruba"):
		return "Access Point"
	default:
		return "Generic"
	}
}

func init() {
	cmd.PrintCmd.AddCommand(summaryPrintCmd)
	summaryPrintCmd.Flags().IntP("level", "l", 1, "Detail level (1-3)")
}