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

// zoneModCmd represents the port command
var devicePortModCmd = &cobra.Command{
	Use:   "deviceport",
	Short: "Add a device port",
	Long: `Attach a model port to a device instance, creating a device port. The MAC
address and VLAN configs are optional; if the (device, model-port) pair already
exists it is updated instead of duplicated.

VLAN configs accept '100:tagged,200:untagged' (or just '100,200' for tagged).

Example:
  nsl-graph add deviceport --deviceid <id> --modelportid <id> \
    --macaddress 00:11:22:33:44:55 --vlan-configs 100:tagged,200:untagged`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"deviceid", "modelportid", "macaddress"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		deviceid := vals[0]
		modelportid := vals[1]
		macaddress := vals[2]

		// Get optional VLAN configs
		// Format: "100:tagged,200:untagged" or just "100,200" (defaults to tagged)
		vlanConfigsStr, _ := cmd.Flags().GetStringSlice("vlan-configs")

		// Parse VLAN configs
		var vlanConfigs []e.PortVlanConfig
		for _, vcStr := range vlanConfigsStr {
			if vcStr == "" {
				continue
			}
			// Check if format is "number:tagged" or "number:untagged"
			parts := strings.Split(vcStr, ":")
			vlanNum := parts[0]
			tagged := true // default to tagged
			if len(parts) == 2 {
				tagged = strings.ToLower(parts[1]) == "tagged"
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

		// Check if deviceport already exists
		existingPort, err := service.GetDevicePortByIDs(deviceid, modelportid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking deviceport: %v\n", err)
			os.Exit(1)
		}

		if existingPort != nil {
			// Update existing deviceport
			if err := service.UpdateDevicePort(deviceid, modelportid, macaddress, vlanConfigs); err != nil {
				fmt.Fprintf(os.Stderr, "Error updating deviceport: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Updated deviceport %s:%s\n", existingPort.DevLabel, existingPort.PortName)
		} else {
			// Create new deviceport
			if _, err := service.AddDevicePort(deviceid, modelportid, macaddress, vlanConfigs); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing deviceport: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created deviceport %s:%s\n", deviceid, modelportid)
		}
	},
}

func init() {
	cmd.AddCmd.AddCommand(devicePortModCmd)

	devicePortModCmd.Flags().
		String("deviceid", "", "Sets the device ID")
	devicePortModCmd.Flags().
		String("modelportid", "", "Sets the model port ID")
	devicePortModCmd.Flags().
		String("macaddress", "", "Sets the MAC address (optional)")
	devicePortModCmd.Flags().
		StringSlice("vlan-configs", []string{}, "VLAN configurations (format: '100:tagged,200:untagged' or just '100,200' for tagged)")
}
