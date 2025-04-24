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
	Ports         map[string]DevicePort
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
		}
	}

	// Collect ports
	for _, c := range connections {
		fromKey := c.FromZoneName + "." + c.FromDevice
		toKey := c.ToZoneName + "." + c.ToDevice
		if dev, ok := deviceMap[fromKey]; ok {
			dev.Ports[c.FromModelPort] = DevicePort{Name: c.FromModelPort}
		}
		if dev, ok := deviceMap[toKey]; ok {
			dev.Ports[c.ToModelPort] = DevicePort{Name: c.ToModelPort}
		}
	}

	// Generate D2
	var d2Connections, d2Devices strings.Builder
	for _, c := range connections {
		from := fmt.Sprintf("%s.%s.%s", c.FromZoneName, c.FromDevice, c.FromModelPort)
		to := fmt.Sprintf("%s.%s.%s", c.ToZoneName, c.ToDevice, c.ToModelPort)
		d2Connections.WriteString(fmt.Sprintf("%s -- %s\n", from, to))
	}
	for key, dev := range deviceMap {
		d2Devices.WriteString(fmt.Sprintf("%s: {\n", key))
		d2Devices.WriteString(fmt.Sprintf("  shape: %s\n", dev.Shape))
		d2Devices.WriteString(fmt.Sprintf("  label: \"%s\"\n", dev.Label))
		i := 1
		for portName := range dev.Ports {
			d2Devices.WriteString(fmt.Sprintf("  %d-1: \"%s\"\n", i, portName))
			i++
		}
		d2Devices.WriteString("}\n")
	}

	return d2Connections.String() + "\n" + d2Devices.String()
}
