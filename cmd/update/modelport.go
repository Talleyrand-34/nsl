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

// ModelPortUpdateCmd represents the model port update command
var ModelPortUpdateCmd = &cobra.Command{
	Use:   "modelport",
	Short: "Update an existing model port",
	Long: `Update an existing model port by its ID.
	
You can update the port name, position coordinates, and the model it belongs to.

Example:
  nsl-graph update modelport --id 1 --name "GigE1/0/2" --position-x 2 --position-y 0 --model-id 1`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required model port ID
		modelPortId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if modelPortId == "" {
			fmt.Fprintf(os.Stderr, "Model port ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		// Get update fields
		newPortName, _ := cmd.Flags().GetString("name")
		newPositionX, _ := cmd.Flags().GetString("position-x")
		newPositionY, _ := cmd.Flags().GetString("position-y")
		newModelId, _ := cmd.Flags().GetString("model-id")
		allowMultiple, _ := cmd.Flags().GetBool("allow-multiple")

		// For this implementation, require all fields
		if newPortName == "" || newPositionX == "" || newPositionY == "" || newModelId == "" {
			fmt.Fprintf(os.Stderr, "All fields are required: --name, --position-x, --position-y, --model-id.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the model port
		err = service.UpdateModelPort(modelPortId, newPortName, newPositionX, newPositionY, newModelId, allowMultiple)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating model port: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated model port with ID %s\n", modelPortId)
		fmt.Printf("  New name: %s\n", newPortName)
		fmt.Printf("  New position: (%s, %s)\n", newPositionX, newPositionY)
		fmt.Printf("  New model ID: %s\n", newModelId)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(ModelPortUpdateCmd)

	ModelPortUpdateCmd.Flags().String("id", "", "ID of the model port to update (required)")
	ModelPortUpdateCmd.Flags().String("name", "", "New name for the port (required)")
	ModelPortUpdateCmd.Flags().String("position-x", "", "New X position coordinate (required)")
	ModelPortUpdateCmd.Flags().String("position-y", "", "New Y position coordinate (required)")
	ModelPortUpdateCmd.Flags().String("model-id", "", "New model ID for the port (required)")
	ModelPortUpdateCmd.Flags().Bool("allow-multiple", false, "Allow multiple connections to this port")
	// ModelPortUpdateCmd.MarkFlagRequired("id")
	// ModelPortUpdateCmd.MarkFlagRequired("name")
	// ModelPortUpdateCmd.MarkFlagRequired("position-x")
	// ModelPortUpdateCmd.MarkFlagRequired("position-y")
	// ModelPortUpdateCmd.MarkFlagRequired("model-id")
}
