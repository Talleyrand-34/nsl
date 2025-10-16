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
package cmd_delete

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// connectionTypeDelCmd represents the connection type delete command
var connectionTypeDelCmd = &cobra.Command{
	Use:   "connectiontype [connection_type_name]",
	Short: "Delete a connection type",
	Long:  `Delete a connection type from the database by name.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		connectionTypeName := args[0]
		
		service, err := util.GetServiceConnection(cmd_root.GetSrcDB())
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		err = service.DeleteConnectionType(connectionTypeName)
		if err != nil {
			fmt.Printf("Error deleting connection type: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Connection type '%s' deleted successfully\n", connectionTypeName)
	},
}

func init() {
	cmd_root.DeleteCmd.AddCommand(connectionTypeDelCmd)
}