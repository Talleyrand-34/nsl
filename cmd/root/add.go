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
/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)
*/
package cmd_root

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd"
)

// AddCmd represents the add command. It creates network entities (brands,
// devices, zones, connections, …); use the separate `update` command to change
// existing ones. `modify` is kept as a backward-compatible alias.
var AddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"modify"},
	Short:   "Add entities to the network (brands, devices, zones, connections, …)",
	Long: `Add new entities to the network model: brands, model types, zones,
models, devices, ports, connections, VLANs and interfaces.

To change an existing entity, use the "update" command instead. The old name
"modify" still works as an alias.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	cmd.RootCmd.AddCommand(AddCmd)
}
