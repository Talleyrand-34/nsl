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
package cmd_update

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// OwnerUpdateCmd represents the owner update command
var OwnerUpdateCmd = &cobra.Command{
	Use:   "owner",
	Short: "Update an existing owner entity",
	Long: `Update an existing owner entity by its ID.
	
A owner entity represents ownership or management responsibility for network resources.

Example:
  nsl-graph update owner --id 1 --name "Network Operations Center"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required flags
		ownerId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if ownerId == "" {
			fmt.Fprintf(os.Stderr, "Owner ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		newOwnerName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'name': %v\n", err)
			os.Exit(1)
		}
		if newOwnerName == "" {
			fmt.Fprintf(os.Stderr, "New owner name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the owner
		err = service.UpdateOwner(ownerId, newOwnerName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating owner: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated owner with ID %s to '%s'\n", ownerId, newOwnerName)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(OwnerUpdateCmd)

	OwnerUpdateCmd.Flags().String("id", "", "ID of the owner to update (required)")
	OwnerUpdateCmd.Flags().String("name", "", "New name for the owner (required)")
	// OwnerUpdateCmd.MarkFlagRequired("id")
	// OwnerUpdateCmd.MarkFlagRequired("name")
}
