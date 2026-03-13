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
package cmd_print

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// localvlanCmd represents the localvlan command for printing
var localvlanCmd = &cobra.Command{
	Use:   "localvlan",
	Short: "Print local VLAN mappings",
	Long: `Print local VLAN mappings from the database.

Local VLANs show device-specific VLAN names that may differ from the global VLAN names.
You can filter by device ID or VLAN ID to show specific mappings.

Examples:
  nsl-graph print localvlan                    # All local VLANs
  nsl-graph print localvlan --device-id abc123 # Local VLANs for specific device
  nsl-graph print localvlan --vlan-id 100      # Local VLANs for specific VLAN`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get optional filter parameters
		deviceID, _ := cmd.Flags().GetString("device-id")
		vlanID, _ := cmd.Flags().GetString("vlan-id")

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		var localVlans interface{}

		// Apply filters based on provided parameters
		if deviceID != "" {
			localVlans, err = service.GetLocalVlansByDevice(deviceID)
		} else if vlanID != "" {
			localVlans, err = service.GetLocalVlansByVlanID(vlanID)
		} else {
			localVlans, err = service.GetLocalVlans()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving local VLANs: %v\n", err)
			os.Exit(1)
		}

		// Print as JSON
		output, err := json.MarshalIndent(localVlans, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(output))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(localvlanCmd)

	localvlanCmd.Flags().String("device-id", "", "Filter by device ID")
	localvlanCmd.Flags().String("vlan-id", "", "Filter by VLAN ID")
}