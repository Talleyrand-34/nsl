
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

// modelDeviceModCmd represents the model creation command
var modelDeviceModCmd = &cobra.Command{
	Use:   "model",
	Short: "Create a new device model",
	Long: `Create a new device model with a specified name, brand, and device class.
	
A model represents a specific network device template (like "ISR4431" or "Catalyst2960") 
that defines the capabilities and characteristics of devices.

Examples:
  nsl-graph modify model --name "ISR4431" --brand "Cisco" --class "Router"
  nsl-graph modify model --name "Catalyst2960" --brand "Cisco" --class "Switch"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get required model name
		modelName, err := cmd.Flags().GetString("name")
		if err != nil || modelName == "" {
			fmt.Fprintf(os.Stderr, "Model name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get required brand name
		brandName, err := cmd.Flags().GetString("brand")
		if err != nil || brandName == "" {
			fmt.Fprintf(os.Stderr, "Brand name is required. Use --brand flag.\n")
			os.Exit(1)
		}

		// Get required device class
		deviceClass, err := cmd.Flags().GetString("class")
		if err != nil || deviceClass == "" {
			fmt.Fprintf(os.Stderr, "Device class is required. Use --class flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Create the model
		err = service.AddModel(modelName, brandName, deviceClass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating model: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created model '%s' from brand '%s' with class '%s'\n", 
			modelName, brandName, deviceClass)
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(modelDeviceModCmd)

	modelDeviceModCmd.Flags().String("name", "", "Model name/identifier (required)")
	modelDeviceModCmd.Flags().String("brand", "", "Brand name for the model (required)")
	modelDeviceModCmd.Flags().String("class", "", "Device class name for the model (required)")
	
	modelDeviceModCmd.MarkFlagRequired("name")
	modelDeviceModCmd.MarkFlagRequired("brand")
	modelDeviceModCmd.MarkFlagRequired("class")
}
