package format

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	d "nslgraph/src/datastructs"
)

// formatDevices formats the devices into the specified output format.

// FormatDevices formats the devices into the specified output format.
func FormatDevices(devices []d.Device) string {
	var sb strings.Builder

	for _, device := range devices {
		fullZonePath := strings.Join(device.ZoneHierarchy, ".")

		sb.WriteString(fmt.Sprintf("%s.%s: {\n", fullZonePath, device.Label))
		sb.WriteString(fmt.Sprintf("  shape: %s\n", device.Shape))
		sb.WriteString(fmt.Sprintf("  label: \"%s\"\n", device.Label))

		if len(device.Ports) > 0 {
			// Collect and sort composite keys (strings) from the map
			keys := make([]string, 0, len(device.Ports))
			for portKey := range device.Ports {
				keys = append(keys, portKey)
			}
			sort.Strings(keys) // Sort composite keys alphabetically

			// Iterate through sorted keys and build the string
			for _, portKey := range keys {
				port := device.Ports[portKey]
				sb.WriteString(fmt.Sprintf("  %s: \"%s\"\n", portKey, port.Name))
			}
		}

		sb.WriteString("}\n")
	}

	return sb.String()
}

// FormatDevicesToJSON formats the devices into the desired JSON structure.

// FormatDevicesToJSON formats devices into a JSON string.
func FormatDevicesToJSON(devices []d.Device, parsePorts bool) (string, error) {
	// Create a map to hold the formatted JSON structure
	formatted := make(map[string]interface{})

	for _, device := range devices {
		fullZonePath := strings.Join(device.ZoneHierarchy, ".")
		deviceKey := fmt.Sprintf("%s.%s", fullZonePath, device.Label)

		if parsePorts {
			// Convert integer keys in device.Ports to string keys
			stringKeyedPorts := make(map[string]d.DevicePort)
			for _, port := range device.Ports {
				stringKeyedPorts[port.Name] = port
			}

			// Add device.ID and ports to the formatted structure
			formatted[deviceKey] = map[string]interface{}{
				"ID":      device.ID,
				"ModelID": device.ModelID,
				"Ports":   stringKeyedPorts,
			}
		} else {
			// Only include ID if parsePorts is false
			formatted[deviceKey] = map[string]interface{}{
				"ID":      device.ID,
				"ModelID": device.ModelID,
			}
		}
	}

	// Convert the map into a JSON string
	jsonData, err := json.MarshalIndent(formatted, "", "  ") // Pretty-print with indentation
	if err != nil {
		return "", fmt.Errorf("failed to marshal devices to JSON: %w", err)
	}

	return string(jsonData), nil
}

// FormatConnections converts connections into the desired string format.
func FormatConnections(connections []d.Connection) string {
	var sb strings.Builder

	for _, conn := range connections {
		fromZone := strings.Join(conn.FromZone, ".")
		toZone := strings.Join(conn.ToZone, ".")

		sb.WriteString(
			fmt.Sprintf(
				"%s.%s.%s -- %s.%s.%s\n",
				fromZone,
				conn.FromDevice,
				conn.FromPort.Name,
				toZone,
				conn.ToDevice,
				conn.ToPort.Name,
			),
		)
	}

	return sb.String()
}
