
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
package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// proprietaryModCmd represents the proprietary creation command
var proprietaryModCmd = &cobra.Command{
	Use:   "proprietary",
	Short: "Add a proprietary owner",
	Long: `Create a new proprietary owner entity.

A proprietary represents the owner or responsible party for devices and zones in the network.
This could be a department, organization, or individual responsible for network assets.

Examples:
  nsl-graph add proprietary --name "IT Department"
  nsl-graph add proprietary --name "Network Operations Team"
  nsl-graph add proprietary --name "Security Division"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get required proprietary name
		proprietaryName, err := cmd.Flags().GetString("name")
		if err != nil || proprietaryName == "" {
			fmt.Fprintf(os.Stderr, "Proprietary name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Create the proprietary
		err = service.AddProprietary(proprietaryName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating proprietary: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created proprietary owner '%s'\n", proprietaryName)
	},
}

func init() {
	cmd.AddCmd.AddCommand(proprietaryModCmd)

	proprietaryModCmd.Flags().String("name", "", "Proprietary owner name (required)")
	
	proprietaryModCmd.MarkFlagRequired("name")
}
