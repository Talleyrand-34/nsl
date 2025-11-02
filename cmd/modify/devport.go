
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
	Short: "deviceport modifications subcommand",
	Long:  `.`,
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
		allowMultipleUntagged, _ := cmd.Flags().GetBool("allow-multiple-untagged")

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
		err = service.AddDevicePort(deviceid, modelportid, macaddress, vlanConfigs, allowMultipleUntagged)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing deviceport: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(devicePortModCmd)

	devicePortModCmd.Flags().
		String("deviceid", "", "Sets the device ID")
	devicePortModCmd.Flags().
		String("modelportid", "", "Sets the model port ID")
	devicePortModCmd.Flags().
		String("macaddress", "", "Sets the MAC address (optional)")
	devicePortModCmd.Flags().
		StringSlice("vlan-configs", []string{}, "VLAN configurations (format: '100:tagged,200:untagged' or just '100,200' for tagged)")
	devicePortModCmd.Flags().
		Bool("allow-multiple-untagged", false, "Allow multiple untagged VLANs per port")
}
