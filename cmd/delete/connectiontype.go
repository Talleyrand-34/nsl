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

	"github.com/spf13/cobra"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// connectionTypeDelCmd represents the connection type delete command
var connectionTypeDelCmd = &cobra.Command{
	Use:   "connectiontype [connection_type_name]",
	Short: "Delete a connection type",
	Long:  `Delete a connection type from the database by name. Connection types have no dependencies, so --cascade has no effect.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		connectionTypeName := args[0]
		
		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			return fmt.Errorf("error connecting to database: %w", err)
		}

		err = service.DeleteConnectionType(connectionTypeName)
		if err != nil {
			return fmt.Errorf("error deleting connection type: %w", err)
		}

		fmt.Printf("Connection type '%s' deleted successfully\n", connectionTypeName)
		return nil
	},
}

func init() {
	cmd_root.DeleteCmd.AddCommand(connectionTypeDelCmd)
	connectionTypeDelCmd.Flags().Bool("cascade", false, "No effect for connection types (they have no dependencies)")
}