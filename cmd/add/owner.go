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
package cmd_add

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// ownerModCmd represents the owner creation command
var ownerModCmd = &cobra.Command{
	Use:   "owner",
	Short: "Add a owner owner",
	Long: `Create a new owner owner entity.

A owner represents the owner or responsible party for devices and zones in the network.
This could be a department, organization, or individual responsible for network assets.

Examples:
  nsl-graph add owner --name "IT Department"
  nsl-graph add owner --name "Network Operations Team"
  nsl-graph add owner --name "Security Division"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get required owner name
		ownerName, err := cmd.Flags().GetString("name")
		if err != nil || ownerName == "" {
			fmt.Fprintf(os.Stderr, "Owner name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Create the owner
		err = service.AddOwner(ownerName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating owner: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created owner owner '%s'\n", ownerName)
	},
}

func init() {
	cmd.AddCmd.AddCommand(ownerModCmd)

	ownerModCmd.Flags().String("name", "", "Owner owner name (required)")

	ownerModCmd.MarkFlagRequired("name")
}
