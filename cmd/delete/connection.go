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
package cmd_delete

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// connectionDelCmd represents the connection delete command
var connectionDelCmd = &cobra.Command{
	Use:   "connection [connection_id]",
	Short: "Delete a connection",
	Long:  `Delete a connection from the database by ID. Connections have no dependencies, so --cascade has no effect.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		connectionId := args[0]
		cascade, _ := cmd.Flags().GetBool("cascade")

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			return fmt.Errorf("error connecting to database: %w", err)
		}

		if cascade {
			err = service.DeleteConnectionCascade(connectionId)
		} else {
			err = service.DeleteConnection(connectionId)
		}

		if err != nil {
			return fmt.Errorf("error deleting connection: %w", err)
		}

		fmt.Printf("Connection with ID '%s' deleted successfully\n", connectionId)
		return nil
	},
}

func init() {
	cmd_root.DeleteCmd.AddCommand(connectionDelCmd)
	connectionDelCmd.Flags().Bool("cascade", false, "No effect for connections (they have no dependencies)")
}