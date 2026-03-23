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

// ProprietaryUpdateCmd represents the proprietary update command
var ProprietaryUpdateCmd = &cobra.Command{
	Use:   "proprietary",
	Short: "Update an existing proprietary entity",
	Long: `Update an existing proprietary entity by its ID.
	
A proprietary entity represents ownership or management responsibility for network resources.

Example:
  nsl-graph update proprietary --id 1 --name "Network Operations Center"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required flags
		proprietaryId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if proprietaryId == "" {
			fmt.Fprintf(os.Stderr, "Proprietary ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		newProprietaryName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'name': %v\n", err)
			os.Exit(1)
		}
		if newProprietaryName == "" {
			fmt.Fprintf(os.Stderr, "New proprietary name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the proprietary
		err = service.UpdateProprietary(proprietaryId, newProprietaryName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating proprietary: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated proprietary with ID %s to '%s'\n", proprietaryId, newProprietaryName)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(ProprietaryUpdateCmd)

	ProprietaryUpdateCmd.Flags().String("id", "", "ID of the proprietary to update (required)")
	ProprietaryUpdateCmd.Flags().String("name", "", "New name for the proprietary (required)")
	// ProprietaryUpdateCmd.MarkFlagRequired("id")
	// ProprietaryUpdateCmd.MarkFlagRequired("name")
}