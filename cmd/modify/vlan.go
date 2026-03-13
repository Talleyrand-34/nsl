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

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// vlanModCmd represents the vlan command
var vlanModCmd = &cobra.Command{
	Use:   "vlan",
	Short: "Create a new VLAN",
	Long: `Create a VLAN (Virtual LAN) by specifying a VLAN ID and name.

A VLAN is a logical network segmentation that can be assigned to connections.

Example:
  nsl-graph modify vlan --vlan-id 100 --name "Management VLAN"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get required VLAN ID
		vlanID, err := cmd.Flags().GetString("vlan-id")
		if err != nil || vlanID == "" {
			fmt.Fprintf(os.Stderr, "VLAN ID is required. Use --vlan-id flag.\n")
			os.Exit(1)
		}

		// Get required VLAN name
		vlanName, err := cmd.Flags().GetString("name")
		if err != nil || vlanName == "" {
			fmt.Fprintf(os.Stderr, "VLAN name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get optional IP segment
		ipSegment, _ := cmd.Flags().GetString("ip-segment")

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Create the VLAN
		err = service.AddVlan(vlanID, vlanName, ipSegment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating VLAN: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created VLAN %s (%s)\n", vlanID, vlanName)
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(vlanModCmd)

	vlanModCmd.Flags().String("vlan-id", "", "VLAN ID (required, e.g., 100)")
	vlanModCmd.Flags().String("name", "", "VLAN name (required, e.g., 'Management VLAN')")
	vlanModCmd.Flags().String("ip-segment", "", "IP segment associated with this VLAN (e.g., '192.168.1.0/24')")

	vlanModCmd.MarkFlagRequired("vlan-id")
	vlanModCmd.MarkFlagRequired("name")
}
