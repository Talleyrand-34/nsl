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

func GenerateD2FocusPorts(devices []e.Device, connections []e.Connection, zones []e.Zone, devicePorts []e.DevicePort, allInterfaces []e.DeviceInterface, ifacePorts []e.InterfacePort, colorPorts bool, showUnusedPorts bool) string {
	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(connections, deviceMap, zoneFullName)
	if showUnusedPorts {
		collectPortsFromDevicePorts(devicePorts, deviceMap)
	}
	assignPortNumbers(deviceMap)

	d2Connections := generateD2ConnectionStrings(connections, deviceMap, zoneFullName, false)

	var d2Devices string
	if colorPorts {
		vlanColorMap := make(map[string]string)
		d2Devices = generateD2DeviceBlocksWithPortColors(deviceMap, devicePorts, allInterfaces, ifacePorts, vlanColorMap)
	} else {
		d2Devices = generateD2DeviceBlocks(deviceMap)
	}

	return d2Connections + "\n" + d2Devices
}

func GenerateD2FocusConnections(devices []e.Device, connections []e.Connection, zones []e.Zone, devicePorts []e.DevicePort, allInterfaces []e.DeviceInterface, ifacePorts []e.InterfacePort, colorPorts bool, showUnusedPorts bool) string {
	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(connections, deviceMap, zoneFullName)
	if showUnusedPorts {
		collectPortsFromDevicePorts(devicePorts, deviceMap)
	}
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

	var d2Devices string
	if colorPorts {
		vlanColorMap := make(map[string]string)
		d2Devices = generateD2DeviceBlocksWithPortColors(deviceMap, devicePorts, allInterfaces, ifacePorts, vlanColorMap)
	} else {
		d2Devices = generateD2DeviceBlocks(deviceMap)
	}

	return d2Connections + "\n" + d2Devices
}

// GenerateD2FocusPortsWithVlans generates a D2 diagram with VLAN support and a legend.
//
// vlanScope controls which VLANs appear on connections:
//   - "untagged" (default): one line per link, colored by the source port's untagged VLAN.
//   - "all": one line per VLAN in the intersection of both ports, multiple lines per link.
//
// colorTarget controls what is colored:
//   - "both" (default): colored connections + colored port nodes.
//   - "connections": colored connections, plain port nodes.
//   - "ports": plain connections, colored port nodes.
func GenerateD2FocusPortsWithVlans(devices []e.Device, connections []e.Connection, zones []e.Zone, devicePorts []e.DevicePort, allInterfaces []e.DeviceInterface, ifacePorts []e.InterfacePort, showUnusedPorts bool, vlanScope, colorTarget string) string {
	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(connections, deviceMap, zoneFullName)
	if showUnusedPorts {
		collectPortsFromDevicePorts(devicePorts, deviceMap)
	}
	assignPortNumbers(deviceMap)

	vlanColorMap := make(map[string]string)

	var d2Connections string
	if colorTarget == "ports" {
		d2Connections = generateD2ConnectionStrings(connections, deviceMap, zoneFullName, false)
	} else {
		d2Connections = generateD2ConnectionStringsWithVlans(connections, deviceMap, zoneFullName, vlanColorMap, devicePorts, allInterfaces, ifacePorts, vlanScope)
	}

	var d2Devices string
	if colorTarget == "connections" {
		d2Devices = generateD2DeviceBlocks(deviceMap)
	} else {
		d2Devices = generateD2DeviceBlocksWithPortColors(deviceMap, devicePorts, allInterfaces, ifacePorts, vlanColorMap)
	}

	d2Legend := generateVlanLegend(vlanColorMap)

	return d2Connections + "\n" + d2Devices + d2Legend
}

// GenerateD2FocusConnectionsWithVlans generates a D2 diagram with sorted connections and VLAN support.
//
// vlanScope controls which VLANs appear on connections:
//   - "untagged" (default): one line per link, colored by the source port's untagged VLAN.
//   - "all": one line per VLAN in the intersection of both ports, multiple lines per link.
//
// colorTarget controls what is colored:
//   - "both" (default): colored connections + colored port nodes.
//   - "connections": colored connections, plain port nodes.
//   - "ports": plain connections, colored port nodes.
func GenerateD2FocusConnectionsWithVlans(devices []e.Device, connections []e.Connection, zones []e.Zone, devicePorts []e.DevicePort, allInterfaces []e.DeviceInterface, ifacePorts []e.InterfacePort, showUnusedPorts bool, vlanScope, colorTarget string) string {
	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(connections, deviceMap, zoneFullName)
	if showUnusedPorts {
		collectPortsFromDevicePorts(devicePorts, deviceMap)
	}
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

	var d2Connections string
	if colorTarget == "ports" {
		d2Connections = generateD2ConnectionStrings(sortedConnections, deviceMap, zoneFullName, false)
	} else {
		d2Connections = generateD2ConnectionStringsWithVlans(sortedConnections, deviceMap, zoneFullName, vlanColorMap, devicePorts, allInterfaces, ifacePorts, vlanScope)
	}

	var d2Devices string
	if colorTarget == "connections" {
		d2Devices = generateD2DeviceBlocks(deviceMap)
	} else {
		d2Devices = generateD2DeviceBlocksWithPortColors(deviceMap, devicePorts, allInterfaces, ifacePorts, vlanColorMap)
	}

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

// collectPortsFromDevicePorts populates device ports from the DevicePort slice,
// adding ports that are not already present (e.g. ports with no connections).
func collectPortsFromDevicePorts(devicePorts []e.DevicePort, deviceMap map[string]*DeviceD2) {
	// Build reverse lookup: device label -> map key
	labelToKey := make(map[string]string)
	for key, dev := range deviceMap {
		labelToKey[dev.Label] = key
	}
	for _, dp := range devicePorts {
		key, ok := labelToKey[dp.DevLabel]
		if !ok {
			continue
		}
		dev := deviceMap[key]
		if _, exists := dev.Ports[dp.PortName]; !exists {
			dev.Ports[dp.PortName] = DevicePort{Name: dp.PortName}
			dev.PortOrder = append(dev.PortOrder, dp.PortName)
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

		fromPath := buildD2DevicePath(zoneFullName[c.FromZoneID], c.FromDevice)
		toPath := buildD2DevicePath(zoneFullName[c.ToZoneID], c.ToDevice)
		from := fmt.Sprintf("%s.%s", fromPath, fromPortNum)
		to := fmt.Sprintf("%s.%s", toPath, toPortNum)

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

// quoteD2Identifier quotes a D2 identifier if it contains special characters like dots
func quoteD2Identifier(name string) string {
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		return fmt.Sprintf(`"%s"`, name)
	}
	return name
}

// buildD2DevicePath creates a properly-quoted D2 path for a device.
// Zone names remain unquoted (they are D2 hierarchy separators).
// The device name is quoted only if it contains special characters.
func buildD2DevicePath(zonePath string, deviceName string) string {
	quoted := quoteD2Identifier(deviceName)
	if zonePath == "" {
		return quoted
	}
	return zonePath + "." + quoted
}

// generateD2DeviceBlocks generates D2 device block definitions with ports
func generateD2DeviceBlocks(deviceMap map[string]*DeviceD2) string {
	var d2Devices strings.Builder

	for _, dev := range deviceMap {
		zonePath := strings.Join(dev.ZoneHierarchy, ".")
		d2Key := buildD2DevicePath(zonePath, dev.Label)
		d2Devices.WriteString(fmt.Sprintf("%s: {\n", d2Key))
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

// getPortUntaggedVlan returns the untagged VLAN number for a specific port via the
// DeviceInterface layer. It accepts the full allInterfaces and ifacePorts slices so
// callers can fetch them once and reuse across many calls.
//
// Logic:
//  1. Find DevicePort matching deviceLabel+portName to get DeviceID+ModelPortID.
//  2. Find all InterfacePort entries where DeviceID+ModelPortID match.
//  3. For each matched DeviceInterface, scan VlanConfigs for an untagged VLAN.
func getPortUntaggedVlan(
	deviceLabel, portName string,
	devicePorts []e.DevicePort,
	allInterfaces []e.DeviceInterface,
	ifacePorts []e.InterfacePort,
) string {
	// Step 1: identify the DevicePort and collect VLANs from it first (physical port's own VLANs)
	var deviceID, modelPortID string
	portVlans := make(map[string]bool) // vlanNumber -> tagged
	for _, dp := range devicePorts {
		if dp.DevLabel == deviceLabel && dp.PortName == portName {
			deviceID = dp.DeviceID
			modelPortID = dp.ModelID
			for _, vc := range dp.VlanConfigs {
				portVlans[vc.VlanNumber] = vc.Tagged
			}
			break
		}
	}
	if deviceID == "" {
		return ""
	}

	// Step 2: collect interface IDs linked to this physical port
	ifaceIDSet := make(map[string]struct{})
	for _, ip := range ifacePorts {
		if ip.DeviceID == deviceID && ip.ModelPortID == modelPortID {
			ifaceIDSet[ip.InterfaceID] = struct{}{}
		}
	}

	// Step 3: scan the matched interfaces for an untagged VLAN
	for _, iface := range allInterfaces {
		if _, ok := ifaceIDSet[iface.ID]; !ok {
			continue
		}
		for _, vc := range iface.VlanConfigs {
			if !vc.Tagged {
				return vc.VlanNumber
			}
		}
	}

	// Step 4: if no untagged VLAN found via InterfacePorts, check DevicePort's own VLANs
	for vlanNum, tagged := range portVlans {
		if !tagged {
			return vlanNum
		}
	}

	return ""
}

// generateD2DeviceBlocksWithPortColors generates D2 device block definitions with colored ports
// Ports are colored by their untagged VLAN color via the DeviceInterface layer.
func generateD2DeviceBlocksWithPortColors(deviceMap map[string]*DeviceD2, devicePorts []e.DevicePort, allInterfaces []e.DeviceInterface, ifacePorts []e.InterfacePort, vlanColorMap map[string]string) string {
	var d2Devices strings.Builder

	for _, dev := range deviceMap {
		zonePath := strings.Join(dev.ZoneHierarchy, ".")
		d2Key := buildD2DevicePath(zonePath, dev.Label)
		d2Devices.WriteString(fmt.Sprintf("%s: {\n", d2Key))
		d2Devices.WriteString(fmt.Sprintf("  shape: %s\n", dev.Shape))
		d2Devices.WriteString(fmt.Sprintf("  label: \"%s\"\n", dev.Label))
		for _, portName := range dev.PortOrder {
			num := dev.PortNumMap[portName]
			// Check if this port has an untagged VLAN via the interface layer
			untaggedVlan := getPortUntaggedVlan(dev.Label, portName, devicePorts, allInterfaces, ifacePorts)
			if untaggedVlan != "" {
				color := getVlanColor(untaggedVlan, vlanColorMap)
				d2Devices.WriteString(fmt.Sprintf("  %s: \"%s\" {\n", num, portName))
				d2Devices.WriteString(fmt.Sprintf("    style.stroke: %s\n", color))
				d2Devices.WriteString("  }\n")
			} else {
				d2Devices.WriteString(fmt.Sprintf("  %s: \"%s\"\n", num, portName))
			}
		}
		d2Devices.WriteString("}\n")
	}

	return d2Devices.String()
}

// GetConnectionVlanInfo returns VLAN information for a connection.
// Returns: intersection VLANs (present on BOTH ports) and missing VLANs (present on ONLY ONE port).
func GetConnectionVlanInfo(
	conn e.Connection,
	devicePorts []e.DevicePort,
	allInterfaces []e.DeviceInterface,
	ifacePorts []e.InterfacePort,
) ([]e.ConnectionVlanInfo, []e.ConnectionVlanInfo) {
	type vlanEntry struct {
		vlanID string
		tagged bool
	}

	getPortVlanMap := func(deviceLabel, portName string) map[string]vlanEntry {
		result := make(map[string]vlanEntry)
		var deviceID, modelPortID string
		for _, dp := range devicePorts {
			if dp.DevLabel == deviceLabel && dp.PortName == portName {
				deviceID = dp.DeviceID
				modelPortID = dp.ModelID
				for _, vc := range dp.VlanConfigs {
					result[vc.VlanNumber] = vlanEntry{vlanID: vc.VlanNumber, tagged: vc.Tagged}
				}
				break
			}
		}
		if deviceID == "" {
			return result
		}
		ifaceIDSet := make(map[string]struct{})
		for _, ip := range ifacePorts {
			if ip.DeviceID == deviceID && ip.ModelPortID == modelPortID {
				ifaceIDSet[ip.InterfaceID] = struct{}{}
			}
		}
		for _, iface := range allInterfaces {
			matched := false
			if _, ok := ifaceIDSet[iface.ID]; ok {
				matched = true
			} else if iface.DeviceID == deviceID && iface.Name == portName {
				matched = true
			}
			if matched {
				for _, vc := range iface.VlanConfigs {
					if _, exists := result[vc.VlanNumber]; !exists {
						result[vc.VlanNumber] = vlanEntry{vlanID: vc.VlanNumber, tagged: vc.Tagged}
					}
				}
			}
		}
		return result
	}

	fromMap := getPortVlanMap(conn.FromDevice, conn.FromModelPort)
	toMap := getPortVlanMap(conn.ToDevice, conn.ToModelPort)

	var intersection, missing []e.ConnectionVlanInfo
	for id, entry := range fromMap {
		if _, exists := toMap[id]; exists {
			intersection = append(intersection, e.ConnectionVlanInfo{VLANID: entry.vlanID, Tagged: entry.tagged})
		} else {
			missing = append(missing, e.ConnectionVlanInfo{VLANID: entry.vlanID, Tagged: entry.tagged})
		}
	}
	for id, entry := range toMap {
		if _, exists := fromMap[id]; !exists {
			missing = append(missing, e.ConnectionVlanInfo{VLANID: entry.vlanID, Tagged: entry.tagged})
		}
	}

	sort.Slice(intersection, func(i, j int) bool { return intersection[i].VLANID < intersection[j].VLANID })
	sort.Slice(missing, func(i, j int) bool { return missing[i].VLANID < missing[j].VLANID })
	return intersection, missing
}

// getConnectionVlans returns the union of VLANs from both device ports in a connection
// via the DeviceInterface layer.
func getConnectionVlans(
	conn e.Connection,
	devicePorts []e.DevicePort,
	allInterfaces []e.DeviceInterface,
	ifacePorts []e.InterfacePort,
) []string {
	vlanSet := make(map[string]bool)

	collectVlans := func(deviceLabel, portName string) {
		// Find the DevicePort
		var deviceID, modelPortID string
		for _, dp := range devicePorts {
			if dp.DevLabel == deviceLabel && dp.PortName == portName {
				deviceID = dp.DeviceID
				modelPortID = dp.ModelID
				break
			}
		}
		if deviceID == "" {
			return
		}
		// Collect interface IDs for this port
		ifaceIDSet := make(map[string]struct{})
		for _, ip := range ifacePorts {
			if ip.DeviceID == deviceID && ip.ModelPortID == modelPortID {
				ifaceIDSet[ip.InterfaceID] = struct{}{}
			}
		}
		// Collect VLANs from matched interfaces
		for _, iface := range allInterfaces {
			if _, ok := ifaceIDSet[iface.ID]; !ok {
				continue
			}
			for _, vc := range iface.VlanConfigs {
				vlanSet[vc.VlanNumber] = true
			}
		}
	}

	collectVlans(conn.FromDevice, conn.FromModelPort)
	collectVlans(conn.ToDevice, conn.ToModelPort)

	// Convert set to sorted slice for consistent output
	vlans := make([]string, 0, len(vlanSet))
	for vlan := range vlanSet {
		vlans = append(vlans, vlan)
	}
	sort.Strings(vlans)

	return vlans
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

// generateD2ConnectionStringsWithVlans generates D2 connection strings with VLAN support.
//
// vlanScope == "untagged": one line per connection colored by the source port's untagged VLAN.
// vlanScope == "all": one line per VLAN in the intersection of both ports; plain line when empty.
func generateD2ConnectionStringsWithVlans(connections []e.Connection, deviceMap map[string]*DeviceD2, zoneFullName map[string]string, vlanColorMap map[string]string, devicePorts []e.DevicePort, allInterfaces []e.DeviceInterface, ifacePorts []e.InterfacePort, vlanScope string) string {
	var d2Connections strings.Builder

	for _, c := range connections {
		fromKey := zoneFullName[c.FromZoneID] + "." + c.FromDevice
		toKey := zoneFullName[c.ToZoneID] + "." + c.ToDevice

		fromPortNum := deviceMap[fromKey].PortNumMap[c.FromModelPort]
		toPortNum := deviceMap[toKey].PortNumMap[c.ToModelPort]

		fromPath := buildD2DevicePath(zoneFullName[c.FromZoneID], c.FromDevice)
		toPath := buildD2DevicePath(zoneFullName[c.ToZoneID], c.ToDevice)
		from := fmt.Sprintf("%s.%s", fromPath, fromPortNum)
		to := fmt.Sprintf("%s.%s", toPath, toPortNum)

		if vlanScope == "all" {
			vlans, _ := GetConnectionVlanInfo(c, devicePorts, allInterfaces, ifacePorts)
			if len(vlans) == 0 {
				d2Connections.WriteString(fmt.Sprintf("%s -- %s\n", from, to))
			} else {
				for _, vlan := range vlans {
					color := getVlanColor(vlan.VLANID, vlanColorMap)
					d2Connections.WriteString(fmt.Sprintf("%s -- %s {\n", from, to))
					d2Connections.WriteString(fmt.Sprintf("    style.stroke: %s\n", color))
					d2Connections.WriteString("}\n")
				}
			}
		} else {
			// "untagged" mode (default): color by source port's untagged VLAN
			untaggedVlan := getPortUntaggedVlan(c.FromDevice, c.FromModelPort, devicePorts, allInterfaces, ifacePorts)
			if untaggedVlan != "" {
				color := getVlanColor(untaggedVlan, vlanColorMap)
				d2Connections.WriteString(fmt.Sprintf("%s -- %s {\n", from, to))
				d2Connections.WriteString(fmt.Sprintf("    style.stroke: %s\n", color))
				d2Connections.WriteString("}\n")
			} else {
				d2Connections.WriteString(fmt.Sprintf("%s -- %s\n", from, to))
			}
		}
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
