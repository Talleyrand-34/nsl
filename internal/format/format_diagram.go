
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
