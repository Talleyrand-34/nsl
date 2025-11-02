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
package format

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	e "nsl-graph/internal/repository/entities"
)

type DevicePort struct {
	Name string
}

type DeviceD2 struct {
	ZoneHierarchy []string
	Label         string
	Shape         string
	Ports         map[string]DevicePort // portName -> DevicePort
	PortOrder     []string              // to preserve order of insertion
	PortNumMap    map[string]string     // portName -> numbered port (e.g., "1-1")
}

func GenerateD2FromJSON(devicesJSON, connectionsJSON []byte) string {
	// Parse JSON
	var devices []e.Device
	var connections []e.Connection
	json.Unmarshal(devicesJSON, &devices)
	json.Unmarshal(connectionsJSON, &connections)

	// Build zone name map from device data (simple zone names, not hierarchy)
	zoneNameMap := make(map[string]string)
	for _, d := range devices {
		zoneNameMap[d.ZoneID] = d.ZoneName
	}

	deviceMap := buildDeviceMap(devices, zoneNameMap)
	collectPortsFromConnections(connections, deviceMap, zoneNameMap)
	assignPortNumbers(deviceMap)

	d2Connections := generateD2ConnectionStrings(connections, deviceMap, zoneNameMap, false)
	d2Devices := generateD2DeviceBlocks(deviceMap)

	return d2Connections + "\n" + d2Devices
}

func GenerateD2FocusPorts(devices []e.Device, connections []e.Connection, zones []e.Zone) string {
	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(connections, deviceMap, zoneFullName)
	assignPortNumbers(deviceMap)

	d2Connections := generateD2ConnectionStrings(connections, deviceMap, zoneFullName, false)
	d2Devices := generateD2DeviceBlocks(deviceMap)

	return d2Connections + "\n" + d2Devices
}

func GenerateD2FocusConnections(devices []e.Device, connections []e.Connection, zones []e.Zone) string {
	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(connections, deviceMap, zoneFullName)
	assignPortNumbers(deviceMap)

	// Sort connections for better D2 diagram flow
	sortedConnections := make([]e.Connection, len(connections))
	copy(sortedConnections, connections)

	sort.Slice(sortedConnections, func(i, j int) bool {
		// Create full device keys for comparison
		fromKeyI := zoneFullName[sortedConnections[i].FromZoneID] + "." + sortedConnections[i].FromDevice
		fromKeyJ := zoneFullName[sortedConnections[j].FromZoneID] + "." + sortedConnections[j].FromDevice

		// First sort by source device
		if fromKeyI != fromKeyJ {
			return fromKeyI < fromKeyJ
		}

		// If same source device, sort by target device
		toKeyI := zoneFullName[sortedConnections[i].ToZoneID] + "." + sortedConnections[i].ToDevice
		toKeyJ := zoneFullName[sortedConnections[j].ToZoneID] + "." + sortedConnections[j].ToDevice

		return toKeyI < toKeyJ
	})

	d2Connections := generateD2ConnectionStrings(sortedConnections, deviceMap, zoneFullName, true)
	d2Devices := generateD2DeviceBlocks(deviceMap)

	return d2Connections + "\n" + d2Devices
}

// GenerateD2FocusPortsWithVlans generates a D2 diagram with VLAN support
// Creates multiple colored connections per VLAN, with a legend
func GenerateD2FocusPortsWithVlans(devices []e.Device, connections []e.Connection, zones []e.Zone) string {
	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(connections, deviceMap, zoneFullName)
	assignPortNumbers(deviceMap)

	vlanColorMap := make(map[string]string)
	d2Connections := generateD2ConnectionStringsWithVlans(connections, deviceMap, zoneFullName, vlanColorMap)
	d2Devices := generateD2DeviceBlocks(deviceMap)
	d2Legend := generateVlanLegend(vlanColorMap)

	return d2Connections + "\n" + d2Devices + d2Legend
}

// GenerateD2FocusConnectionsWithVlans generates a D2 diagram with sorted connections and VLAN support
// Creates multiple colored connections per VLAN, with a legend
func GenerateD2FocusConnectionsWithVlans(devices []e.Device, connections []e.Connection, zones []e.Zone) string {
	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(connections, deviceMap, zoneFullName)
	assignPortNumbers(deviceMap)

	// Sort connections for better D2 diagram flow
	sortedConnections := make([]e.Connection, len(connections))
	copy(sortedConnections, connections)

	sort.Slice(sortedConnections, func(i, j int) bool {
		// Create full device keys for comparison
		fromKeyI := zoneFullName[sortedConnections[i].FromZoneID] + "." + sortedConnections[i].FromDevice
		fromKeyJ := zoneFullName[sortedConnections[j].FromZoneID] + "." + sortedConnections[j].FromDevice

		// First sort by source device
		if fromKeyI != fromKeyJ {
			return fromKeyI < fromKeyJ
		}

		// If same source device, sort by target device
		toKeyI := zoneFullName[sortedConnections[i].ToZoneID] + "." + sortedConnections[i].ToDevice
		toKeyJ := zoneFullName[sortedConnections[j].ToZoneID] + "." + sortedConnections[j].ToDevice

		return toKeyI < toKeyJ
	})

	vlanColorMap := make(map[string]string)
	d2Connections := generateD2ConnectionStringsWithVlans(sortedConnections, deviceMap, zoneFullName, vlanColorMap)
	d2Devices := generateD2DeviceBlocks(deviceMap)
	d2Legend := generateVlanLegend(vlanColorMap)

	return d2Connections + "\n" + d2Devices + d2Legend
}

func buildZoneFullNameMap(zones []e.Zone) map[string]string {
	zoneByID := make(map[string]e.Zone)
	for _, z := range zones {
		zoneByID[z.ID] = z
	}

	fullNameByID := make(map[string]string)
	var getFullName func(string) string
	getFullName = func(id string) string {
		// If already computed, return it
		if name, ok := fullNameByID[id]; ok {
			return name
		}
		z, ok := zoneByID[id]
		if !ok {
			return "" // or panic/error
		}
		if z.FatherID == "" || z.FatherID == "0" {
			fullNameByID[id] = z.Name
		} else {
			parentFull := getFullName(z.FatherID)
			fullNameByID[id] = parentFull + "." + z.Name
		}
		return fullNameByID[id]
	}

	// Compute for all zones
	for id := range zoneByID {
		getFullName(id)
	}
	return fullNameByID
}

// buildDeviceMap creates a device map from devices with their zone hierarchy
func buildDeviceMap(devices []e.Device, zoneFullName map[string]string) map[string]*DeviceD2 {
	deviceMap := make(map[string]*DeviceD2)
	for _, d := range devices {
		fullZoneName := zoneFullName[d.ZoneID]
		key := fullZoneName + "." + d.Name
		deviceMap[key] = &DeviceD2{
			ZoneHierarchy: strings.Split(fullZoneName, "."),
			Label:         d.Name,
			Shape:         "rectangle",
			Ports:         make(map[string]DevicePort),
			PortOrder:     []string{},
			PortNumMap:    make(map[string]string),
		}
	}
	return deviceMap
}

// collectPortsFromConnections populates device ports from connections, preserving insertion order
func collectPortsFromConnections(connections []e.Connection, deviceMap map[string]*DeviceD2, zoneFullName map[string]string) {
	for _, c := range connections {
		fromFullZone := zoneFullName[c.FromZoneID]
		toFullZone := zoneFullName[c.ToZoneID]
		fromKey := fromFullZone + "." + c.FromDevice
		toKey := toFullZone + "." + c.ToDevice

		if dev, ok := deviceMap[fromKey]; ok {
			if _, exists := dev.Ports[c.FromModelPort]; !exists {
				dev.Ports[c.FromModelPort] = DevicePort{Name: c.FromModelPort}
				dev.PortOrder = append(dev.PortOrder, c.FromModelPort)
			}
		}
		if dev, ok := deviceMap[toKey]; ok {
			if _, exists := dev.Ports[c.ToModelPort]; !exists {
				dev.Ports[c.ToModelPort] = DevicePort{Name: c.ToModelPort}
				dev.PortOrder = append(dev.PortOrder, c.ToModelPort)
			}
		}
	}
}

// assignPortNumbers assigns numbered port identifiers (e.g., 1-1, 2-1, ...) to each device port
func assignPortNumbers(deviceMap map[string]*DeviceD2) {
	for _, dev := range deviceMap {
		for i, portName := range dev.PortOrder {
			num := fmt.Sprintf("%d-1", i+1)
			dev.PortNumMap[portName] = num
		}
	}
}

// generateD2ConnectionStrings generates D2 connection strings from connections
// If includeIPs is true, IP segments are included as comments
func generateD2ConnectionStrings(connections []e.Connection, deviceMap map[string]*DeviceD2, zoneFullName map[string]string, includeIPs bool) string {
	var d2Connections strings.Builder

	for _, c := range connections {
		fromKey := zoneFullName[c.FromZoneID] + "." + c.FromDevice
		toKey := zoneFullName[c.ToZoneID] + "." + c.ToDevice

		fromPortNum := deviceMap[fromKey].PortNumMap[c.FromModelPort]
		toPortNum := deviceMap[toKey].PortNumMap[c.ToModelPort]

		from := fmt.Sprintf("%s.%s", fromKey, fromPortNum)
		to := fmt.Sprintf("%s.%s", toKey, toPortNum)

		// Add IP addresses as comments if requested and they exist
		// var ipComment string
		// if includeIPs && (c.FromIPSegment != "" || c.ToIPSegment != "") {
		// 	ipComment = fmt.Sprintf(": %s -- %s", c.FromIPSegment, c.ToIPSegment)
		// }

		// d2Connections.WriteString(fmt.Sprintf("%s -- %s%s\n", from, to, ipComment))
		d2Connections.WriteString(fmt.Sprintf("%s -- %s\n", from, to))
	}

	return d2Connections.String()
}

// generateD2DeviceBlocks generates D2 device block definitions with ports
func generateD2DeviceBlocks(deviceMap map[string]*DeviceD2) string {
	var d2Devices strings.Builder

	for key, dev := range deviceMap {
		d2Devices.WriteString(fmt.Sprintf("%s: {\n", key))
		d2Devices.WriteString(fmt.Sprintf("  shape: %s\n", dev.Shape))
		d2Devices.WriteString(fmt.Sprintf("  label: \"%s\"\n", dev.Label))
		for _, portName := range dev.PortOrder {
			num := dev.PortNumMap[portName]
			d2Devices.WriteString(fmt.Sprintf("  %s: \"%s\"\n", num, portName))
		}
		d2Devices.WriteString("}\n")
	}

	return d2Devices.String()
}

// getVlanColor returns a color for a VLAN ID from a predefined list
// Colors cycle through: red, pink, purple, brown, blue, green, orange
func getVlanColor(vlanID string, vlanColorMap map[string]string) string {
	if color, exists := vlanColorMap[vlanID]; exists {
		return color
	}

	colors := []string{"red", "pink", "purple", "brown", "blue", "green", "orange"}
	color := colors[len(vlanColorMap)%len(colors)]
	vlanColorMap[vlanID] = color
	return color
}

// generateD2ConnectionStringsWithVlans generates D2 connection strings with VLAN support
// Creates multiple colored connections per VLAN, or a single connection if no VLANs
func generateD2ConnectionStringsWithVlans(connections []e.Connection, deviceMap map[string]*DeviceD2, zoneFullName map[string]string, vlanColorMap map[string]string) string {
	var d2Connections strings.Builder

	for _, c := range connections {
		fromKey := zoneFullName[c.FromZoneID] + "." + c.FromDevice
		toKey := zoneFullName[c.ToZoneID] + "." + c.ToDevice

		fromPortNum := deviceMap[fromKey].PortNumMap[c.FromModelPort]
		toPortNum := deviceMap[toKey].PortNumMap[c.ToModelPort]

		from := fmt.Sprintf("%s.%s", fromKey, fromPortNum)
		to := fmt.Sprintf("%s.%s", toKey, toPortNum)

		// If connection has VLANs, create one colored connection per VLAN
		// if len(c.VlanCon) > 0 {
		// 	for _, vlanID := range c.VlanCon {
		// 		color := getVlanColor(vlanID, vlanColorMap)
		// 		d2Connections.WriteString(fmt.Sprintf("%s -- %s{\n", from, to))
		// 		d2Connections.WriteString(fmt.Sprintf("    style.stroke: %s\n", color))
		// 		d2Connections.WriteString("}\n")
		// 	}
		// } else {
		// 	// No VLANs - create normal connection
		// 	d2Connections.WriteString(fmt.Sprintf("%s -- %s\n", from, to))
		// }
		d2Connections.WriteString(fmt.Sprintf("%s -- %s\n", from, to))
	}

	return d2Connections.String()
}

// generateVlanLegend generates a D2 legend section showing VLAN to color mappings
func generateVlanLegend(vlanColorMap map[string]string) string {
	if len(vlanColorMap) == 0 {
		return ""
	}

	var legend strings.Builder
	legend.WriteString("vars: {\n\n")
	legend.WriteString("  d2-legend: {\n\n")
	legend.WriteString("    a.style.opacity: 0\n\n")
	legend.WriteString("    b.style.opacity: 0\n\n")

	// Sort VLANs for consistent output
	vlans := make([]string, 0, len(vlanColorMap))
	for vlan := range vlanColorMap {
		vlans = append(vlans, vlan)
	}
	sort.Strings(vlans)

	for _, vlan := range vlans {
		color := vlanColorMap[vlan]
		legend.WriteString(fmt.Sprintf("    a -- b: VLAN %s {\n", vlan))
		legend.WriteString(fmt.Sprintf("      style.stroke: %s\n", color))
		legend.WriteString("    }\n")
	}

	legend.WriteString("  }\n")
	legend.WriteString("}\n")

	return legend.String()
}
