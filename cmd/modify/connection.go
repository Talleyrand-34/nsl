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

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
	application "nsl-graph/internal/repository/application"
)

// connectionModCmd represents the connection creation command
var connectionModCmd = &cobra.Command{
	Use:   "connection",
	Short: "Add a connection between two device ports",
	Long: `Create a connection between two device ports.

Identify endpoints by name (device label or IP + port name):
  nsl-graph add connection --from-device 10.0.0.245 --from-modelport eth0 \
                              --to-device router.local --to-modelport igc1

Or by database IDs (legacy):
  nsl-graph add connection --from-device-id <id> --from-modelport-id <id> \
                              --to-device-id <id> --to-modelport-id <id>`,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		fromDeviceName, _ := cmd.Flags().GetString("from-device")
		fromPortName, _ := cmd.Flags().GetString("from-modelport")
		toDeviceName, _ := cmd.Flags().GetString("to-device")
		toPortName, _ := cmd.Flags().GetString("to-modelport")

		fromDeviceId, _ := cmd.Flags().GetString("from-device-id")
		fromModelPortId, _ := cmd.Flags().GetString("from-modelport-id")
		toDeviceId, _ := cmd.Flags().GetString("to-device-id")
		toModelPortId, _ := cmd.Flags().GetString("to-modelport-id")

		// Resolve source device
		if fromDeviceName != "" {
			fromDeviceId, err = resolveDeviceID(service, fromDeviceName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error resolving source device: %v\n", err)
				os.Exit(1)
			}
		}
		if fromDeviceId == "" {
			fmt.Fprintf(os.Stderr, "Source device is required. Use --from-device (name/IP) or --from-device-id.\n")
			os.Exit(1)
		}

		// Resolve source port
		if fromPortName != "" {
			fromModelPortId, err = resolveModelPortID(service, fromPortName, fromDeviceId)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error resolving source port: %v\n", err)
				os.Exit(1)
			}
		}
		if fromModelPortId == "" {
			fmt.Fprintf(os.Stderr, "Source model port is required. Use --from-modelport (name) or --from-modelport-id.\n")
			os.Exit(1)
		}

		// Resolve destination device
		if toDeviceName != "" {
			toDeviceId, err = resolveDeviceID(service, toDeviceName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error resolving destination device: %v\n", err)
				os.Exit(1)
			}
		}
		if toDeviceId == "" {
			fmt.Fprintf(os.Stderr, "Destination device is required. Use --to-device (name/IP) or --to-device-id.\n")
			os.Exit(1)
		}

		// Resolve destination port
		if toPortName != "" {
			toModelPortId, err = resolveModelPortID(service, toPortName, toDeviceId)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error resolving destination port: %v\n", err)
				os.Exit(1)
			}
		}
		if toModelPortId == "" {
			fmt.Fprintf(os.Stderr, "Destination model port is required. Use --to-modelport (name) or --to-modelport-id.\n")
			os.Exit(1)
		}

		// Get deviceport IDs
		fromDeviceportId, err := resolveDeviceportID(service, fromDeviceId, fromModelPortId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding source deviceport: %v\n", err)
			os.Exit(1)
		}
		toDeviceportId, err := resolveDeviceportID(service, toDeviceId, toModelPortId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding destination deviceport: %v\n", err)
			os.Exit(1)
		}

		err = service.AddConnection(fromDeviceportId, toDeviceportId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating connection: %v\n", err)
			os.Exit(1)
		}

		fromLabel := fromDeviceName
		if fromLabel == "" {
			fromLabel = fromDeviceId
		}
		toLabel := toDeviceName
		if toLabel == "" {
			toLabel = toDeviceId
		}
		fromPortLabel := fromPortName
		if fromPortLabel == "" {
			fromPortLabel = fromModelPortId
		}
		toPortLabel := toPortName
		if toPortLabel == "" {
			toPortLabel = toModelPortId
		}
		fmt.Printf("Successfully created connection from %s (%s) to %s (%s)\n",
			fromLabel, fromPortLabel, toLabel, toPortLabel)
	},
}

func init() {
	cmd.AddCmd.AddCommand(connectionModCmd)

	// Name-based flags (preferred)
	connectionModCmd.Flags().String("from-device", "", "Label or IP of the source device")
	connectionModCmd.Flags().String("from-modelport", "", "Port name of the source device")
	connectionModCmd.Flags().String("to-device", "", "Label or IP of the destination device")
	connectionModCmd.Flags().String("to-modelport", "", "Port name of the destination device")

	// ID-based flags (legacy)
	connectionModCmd.Flags().String("from-device-id", "", "ID of the source device")
	connectionModCmd.Flags().String("from-modelport-id", "", "ID of the source device's model port")
	connectionModCmd.Flags().String("to-device-id", "", "ID of the destination device")
	connectionModCmd.Flags().String("to-modelport-id", "", "ID of the destination device's model port")

	connectionModCmd.Flags().Bool("allow-vlan-union", false, "Allow connection if VLANs have any overlap (default: strict matching)")
}

// resolveDeviceID finds a device by label or IP and returns its database ID.
func resolveDeviceID(service application.NetServiceInt, nameOrIP string) (string, error) {
	devices, err := service.GetDevices()
	if err != nil {
		return "", fmt.Errorf("could not fetch devices: %w", err)
	}
	for _, d := range devices {
		if d.Name == nameOrIP {
			return d.ID, nil
		}
		// Search by interface IPs
		ifaces, err := service.GetDeviceInterfaces(d.ID)
		if err != nil {
			continue
		}
		for _, iface := range ifaces {
			for _, ip := range iface.IPAddresses {
				if ip == nameOrIP {
					return d.ID, nil
				}
			}
		}
	}
	return "", fmt.Errorf("device not found: %s", nameOrIP)
}

// resolveModelPortID finds a model port by name that belongs to the model of the given device.
func resolveModelPortID(service application.NetServiceInt, portName string, deviceID string) (string, error) {
	devices, err := service.GetDevices()
	if err != nil {
		return "", fmt.Errorf("could not fetch devices: %w", err)
	}
	var modelName string
	for _, d := range devices {
		if d.ID == deviceID {
			modelName = d.Model
			break
		}
	}
	if modelName == "" {
		return "", fmt.Errorf("device ID not found or has no model: %s", deviceID)
	}

	ports, err := service.GetModelPorts()
	if err != nil {
		return "", fmt.Errorf("could not fetch model ports: %w", err)
	}
	for _, p := range ports {
		if p.Name == portName && p.Model == modelName {
			return p.ID, nil
		}
	}
	return "", fmt.Errorf("port %q not found on model %q", portName, modelName)
}

// resolveDeviceportID finds a deviceport by device ID and model port ID.
func resolveDeviceportID(service application.NetServiceInt, deviceID string, modelPortID string) (string, error) {
	deviceports, err := service.GetDevicePorts()
	if err != nil {
		return "", fmt.Errorf("could not fetch device ports: %w", err)
	}
	for _, dp := range deviceports {
		if dp.DeviceID == deviceID && dp.ModelID == modelPortID {
			return dp.ID, nil
		}
	}
	return "", fmt.Errorf("deviceport not found for device %s and model port %s", deviceID, modelPortID)
}
