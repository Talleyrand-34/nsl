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

// ModelUpdateCmd represents the model update command
var ModelUpdateCmd = &cobra.Command{
	Use:   "model",
	Short: "Update an existing device model",
	Long: `Update an existing device model by its ID.
	
You must specify the model ID and can update the model name, brand, and model type.
Brand and model type should be provided as IDs.

Examples:
  nsl-graph update model --id 1 --name "ISR4431-V2"
  nsl-graph update model --id 1 --name "ISR4431" --brand-id 2 --modeltype-id 3`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required model ID
		modelId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if modelId == "" {
			fmt.Fprintf(os.Stderr, "Model ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		// Get update fields
		newModelName, _ := cmd.Flags().GetString("name")
		newBrandId, _ := cmd.Flags().GetString("brand-id")
		newModelTypeId, _ := cmd.Flags().GetString("model-type-id")
		newOsTypeId, _ := cmd.Flags().GetString("os-type-id")

		// For this implementation, require all fields
		if newModelName == "" || newBrandId == "" || newModelTypeId == "" {
			fmt.Fprintf(os.Stderr, "All fields are required: --name, --brand-id, --model-type-id.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the model
		err = service.UpdateModel(modelId, newModelName, newBrandId, newModelTypeId, newOsTypeId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating model: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated model with ID %s\n", modelId)
		fmt.Printf("  New name: %s\n", newModelName)
		fmt.Printf("  New brand ID: %s\n", newBrandId)
		fmt.Printf("  New model type ID: %s\n", newModelTypeId)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(ModelUpdateCmd)

	ModelUpdateCmd.Flags().String("id", "", "ID of the model to update (required)")
	ModelUpdateCmd.Flags().String("name", "", "New name for the model (required)")
	ModelUpdateCmd.Flags().String("brand-id", "", "New brand ID for the model (required)")
	ModelUpdateCmd.Flags().String("os-type-id", "", "New OS type ID for the model (optional)")
	ModelUpdateCmd.Flags().String("model-type-id", "", "New model type ID for the model (required)")
	// ModelUpdateCmd.MarkFlagRequired("id")
	// ModelUpdateCmd.MarkFlagRequired("name")
	// ModelUpdateCmd.MarkFlagRequired("brand-id")
	// ModelUpdateCmd.MarkFlagRequired("model-type-id")
}
