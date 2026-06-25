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

// modelTypeUpdateCmd represents the model type update command
var modelTypeUpdateCmd = &cobra.Command{
	Use:   "modeltype",
	Short: "Update an existing model type",
	Long: `Update an existing model type by its ID.
	
A model type categorizes network equipment types (e.g., router, switch, firewall).

Example:
  nsl-graph update modeltype --id 1 --name "Layer3Switch"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required flags
		modelTypeId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if modelTypeId == "" {
			fmt.Fprintf(os.Stderr, "Model type ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		newModelTypeName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'name': %v\n", err)
			os.Exit(1)
		}
		if newModelTypeName == "" {
			fmt.Fprintf(os.Stderr, "New model type name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the model type
		err = service.UpdateModelType(modelTypeId, newModelTypeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating model type: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated model type with ID %s to '%s'\n", modelTypeId, newModelTypeName)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(modelTypeUpdateCmd)

	modelTypeUpdateCmd.Flags().String("id", "", "ID of the model type to update (required)")
	modelTypeUpdateCmd.Flags().String("name", "", "New name for the model type (required)")
	modelTypeUpdateCmd.MarkFlagRequired("id")
	modelTypeUpdateCmd.MarkFlagRequired("name")
}
