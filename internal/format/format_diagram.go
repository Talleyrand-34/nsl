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

	// Build device map
	deviceMap := make(map[string]*DeviceD2)
	for _, d := range devices {
		key := d.ZoneName + "." + d.Name
		deviceMap[key] = &DeviceD2{
			ZoneHierarchy: []string{d.ZoneName},
			Label:         d.Name,
			Shape:         "rectangle",
			Ports:         make(map[string]DevicePort),
			PortOrder:     []string{},
			PortNumMap:    make(map[string]string),
		}
	}

	// Collect ports and preserve insertion order
	for _, c := range connections {
		fromKey := c.FromZoneName + "." + c.FromDevice
		toKey := c.ToZoneName + "." + c.ToDevice
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

	// Assign numbered ports per device (e.g., 1-1, 2-1, ...)
	for _, dev := range deviceMap {
		for i, portName := range dev.PortOrder {
			num := fmt.Sprintf("%d-1", i+1)
			dev.PortNumMap[portName] = num
		}
	}

	// Generate D2 connections using numbered ports
	var d2Connections, d2Devices strings.Builder
	for _, c := range connections {
		fromKey := c.FromZoneName + "." + c.FromDevice
		toKey := c.ToZoneName + "." + c.ToDevice
		fromPortNum := deviceMap[fromKey].PortNumMap[c.FromModelPort]
		toPortNum := deviceMap[toKey].PortNumMap[c.ToModelPort]
		from := fmt.Sprintf("%s.%s", fromKey, fromPortNum)
		to := fmt.Sprintf("%s.%s", toKey, toPortNum)
		d2Connections.WriteString(fmt.Sprintf("%s -- %s\n", from, to))
	}

	// Generate D2 device blocks
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

	return d2Connections.String() + "\n" + d2Devices.String()
}

// func GenerateD2FromStruct(devices []e.Device, connections []e.Connection, zones []e.Zone) string {
// 	// Build device map
// 	deviceMap := make(map[string]*DeviceD2)
// 	for _, d := range devices {
// 		key := d.ZoneName + "." + d.Name
// 		deviceMap[key] = &DeviceD2{
// 			ZoneHierarchy: []string{d.ZoneName},
// 			Label:         d.Name,
// 			Shape:         "rectangle",
// 			Ports:         make(map[string]DevicePort),
// 			PortOrder:     []string{},
// 			PortNumMap:    make(map[string]string),
// 		}
// 	}
//
// 	// Collect ports and preserve insertion order
// 	for _, c := range connections {
// 		fromKey := c.FromZoneName + "." + c.FromDevice
// 		toKey := c.ToZoneName + "." + c.ToDevice
// 		if dev, ok := deviceMap[fromKey]; ok {
// 			if _, exists := dev.Ports[c.FromModelPort]; !exists {
// 				dev.Ports[c.FromModelPort] = DevicePort{Name: c.FromModelPort}
// 				dev.PortOrder = append(dev.PortOrder, c.FromModelPort)
// 			}
// 		}
// 		if dev, ok := deviceMap[toKey]; ok {
// 			if _, exists := dev.Ports[c.ToModelPort]; !exists {
// 				dev.Ports[c.ToModelPort] = DevicePort{Name: c.ToModelPort}
// 				dev.PortOrder = append(dev.PortOrder, c.ToModelPort)
// 			}
// 		}
// 	}
//
// 	// Assign numbered ports per device (e.g., 1-1, 2-1, ...)
// 	for _, dev := range deviceMap {
// 		for i, portName := range dev.PortOrder {
// 			num := fmt.Sprintf("%d-1", i+1)
// 			dev.PortNumMap[portName] = num
// 		}
// 	}
//
// 	// Generate D2 connections using numbered ports
// 	var d2Connections, d2Devices strings.Builder
// 	for _, c := range connections {
// 		fromKey := c.FromZoneName + "." + c.FromDevice
// 		toKey := c.ToZoneName + "." + c.ToDevice
// 		fromPortNum := deviceMap[fromKey].PortNumMap[c.FromModelPort]
// 		toPortNum := deviceMap[toKey].PortNumMap[c.ToModelPort]
// 		from := fmt.Sprintf("%s.%s", fromKey, fromPortNum)
// 		to := fmt.Sprintf("%s.%s", toKey, toPortNum)
// 		d2Connections.WriteString(fmt.Sprintf("%s -- %s\n", from, to))
// 	}
//
// 	// Generate D2 device blocks
// 	for key, dev := range deviceMap {
// 		d2Devices.WriteString(fmt.Sprintf("%s: {\n", key))
// 		d2Devices.WriteString(fmt.Sprintf("  shape: %s\n", dev.Shape))
// 		d2Devices.WriteString(fmt.Sprintf("  label: \"%s\"\n", dev.Label))
// 		for _, portName := range dev.PortOrder {
// 			num := dev.PortNumMap[portName]
// 			d2Devices.WriteString(fmt.Sprintf("  %s: \"%s\"\n", num, portName))
// 		}
// 		d2Devices.WriteString("}\n")
// 	}
//
// 	return d2Connections.String() + "\n" + d2Devices.String()
// }

func GenerateD2FromStruct(devices []e.Device, connections []e.Connection, zones []e.Zone) string {
	zoneFullName := buildZoneFullNameMap(zones)

	// Build device map
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

	// Collect ports and preserve insertion order
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

	// Assign numbered ports per device (e.g., 1-1, 2-1, ...)
	for _, dev := range deviceMap {
		for i, portName := range dev.PortOrder {
			num := fmt.Sprintf("%d-1", i+1)
			dev.PortNumMap[portName] = num
		}
	}

	// Generate D2 connections using numbered ports
	var d2Connections, d2Devices strings.Builder

	for _, c := range connections {
		//fromKey := c.FromZoneName + "." + c.FromDevice
		//toKey := c.ToZoneName + "." + c.ToDevice
		fromKey := zoneFullName[c.FromZoneID] + "." + c.FromDevice
		toKey := zoneFullName[c.ToZoneID] + "." + c.ToDevice

		// Lookups unchanged
		fromPortNum := deviceMap[fromKey].PortNumMap[c.FromModelPort]
		toPortNum := deviceMap[toKey].PortNumMap[c.ToModelPort]

		// For output, attach full zone path to the device name
		//fromZonePath := zoneFullName[c.FromZoneID]
		//toZonePath := zoneFullName[c.ToZoneID]
		// Remove the last part (device name) if it's already in fromKey/toKey
		// But in your case, just use the full path and device name

		from := fmt.Sprintf("%s.%s", fromKey, fromPortNum)
		to := fmt.Sprintf("%s.%s", toKey, toPortNum)
		d2Connections.WriteString(fmt.Sprintf("%s -- %s\n", from, to))

	}
	// Generate D2 device blocks
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

	return d2Connections.String() + "\n" + d2Devices.String()
}

func buildZoneFullNameMap(zones []e.Zone) map[int]string {
	zoneByID := make(map[int]e.Zone)
	for _, z := range zones {
		zoneByID[z.ID] = z
	}

	fullNameByID := make(map[int]string)
	var getFullName func(int) string
	getFullName = func(id int) string {
		// If already computed, return it
		if name, ok := fullNameByID[id]; ok {
			return name
		}
		z, ok := zoneByID[id]
		if !ok {
			return "" // or panic/error
		}
		if z.FatherID == 0 {
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
