// SPDX-License-Identifier: AGPL-3.0-or-later
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
package cmd_print

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// vlanPrintCmd represents the vlan print command
var vlanPrintCmd = &cobra.Command{
	Use:   "vlan",
	Short: "Print all VLANs",
	Long:  `Display all configured VLANs in JSON format.`,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			return
		}

		vlans, err := service.GetVlans()
		if err != nil {
			fmt.Println("Error getting VLANs:", err)
			return
		}

		jsonBytes, err := json.MarshalIndent(vlans, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling VLANs to JSON:", err)
			return
		}

		fmt.Println(string(jsonBytes))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(vlanPrintCmd)
}
