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
package cmd_root

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd"
)

// UpdateCmd represents the update command
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update existing entities in the network structure",
	Long: `Update existing network entities by their ID. 
	
All update operations require the entity ID and the new values for the fields you want to change.
Examples:
  nsl-graph update brand --id 1 --name "New Brand Name"
  nsl-graph update device --id 1 --label "New Device Name" --model-id 2`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update called - use a specific subcommand to update entities")
		cmd.Help()
	},
}

func init() {
	cmd.RootCmd.AddCommand(UpdateCmd)
}
