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

// BrandUpdateCmd represents the brand update command
var BrandUpdateCmd = &cobra.Command{
	Use:   "brand",
	Short: "Update an existing brand",
	Long: `Update an existing brand by its ID.
	
You must specify both the brand ID and the new name for the brand.

Example:
  nsl-graph update brand --id 1 --name "Cisco Systems"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the required flags
		brandId, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'id': %v\n", err)
			os.Exit(1)
		}
		if brandId == "" {
			fmt.Fprintf(os.Stderr, "Brand ID is required. Use --id flag.\n")
			os.Exit(1)
		}

		newBrandName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag 'name': %v\n", err)
			os.Exit(1)
		}
		if newBrandName == "" {
			fmt.Fprintf(os.Stderr, "New brand name is required. Use --name flag.\n")
			os.Exit(1)
		}

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to service: %v\n", err)
			os.Exit(1)
		}

		// Update the brand
		err = service.UpdateBrand(brandId, newBrandName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating brand: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated brand with ID %s to '%s'\n", brandId, newBrandName)
	},
}

func init() {
	cmd.UpdateCmd.AddCommand(BrandUpdateCmd)

	BrandUpdateCmd.Flags().String("id", "", "ID of the brand to update (required)")
	BrandUpdateCmd.Flags().String("name", "", "New name for the brand (required)")
	// brandUpdateCmd.MarkFlagRequired("id")
	// brandUpdateCmd.MarkFlagRequired("name")
}
