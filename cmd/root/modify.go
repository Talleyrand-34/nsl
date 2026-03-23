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

// ModifyCmd represents the modify command
var ModifyCmd = &cobra.Command{
	Use:   "modify",
	Short: "Make modificatons into the network structure",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	cmd.RootCmd.AddCommand(ModifyCmd)
}
