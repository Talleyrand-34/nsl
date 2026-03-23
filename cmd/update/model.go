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
	
You must specify the model ID and can update the model name, brand, and device class.
Brand and device class should be provided as IDs.

Examples:
  nsl-graph update model --id 1 --name "ISR4431-V2"
  nsl-graph update model --id 1 --name "ISR4431" --brand-id 2 --deviceclass-id 3`,
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
		newDeviceClassId, _ := cmd.Flags().GetString("deviceclass-id")

		// For this implementation, require all fields
		if newModelName == "" || newBrandId == "" || newDeviceClassId == "" {
			fmt.Fprintf(os.Stderr, "All fields are required: --name, --brand-id, --deviceclass-id.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the model
		err = service.UpdateModel(modelId, newModelName, newBrandId, newDeviceClassId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating model: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated model with ID %s\n", modelId)
		fmt.Printf("  New name: %s\n", newModelName)
		fmt.Printf("  New brand ID: %s\n", newBrandId)
		fmt.Printf("  New device class ID: %s\n", newDeviceClassId)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(ModelUpdateCmd)

	ModelUpdateCmd.Flags().String("id", "", "ID of the model to update (required)")
	ModelUpdateCmd.Flags().String("name", "", "New name for the model (required)")
	ModelUpdateCmd.Flags().String("brand-id", "", "New brand ID for the model (required)")
	ModelUpdateCmd.Flags().String("deviceclass-id", "", "New device class ID for the model (required)")
	// ModelUpdateCmd.MarkFlagRequired("id")
	// ModelUpdateCmd.MarkFlagRequired("name")
	// ModelUpdateCmd.MarkFlagRequired("brand-id")
	// ModelUpdateCmd.MarkFlagRequired("deviceclass-id")
}