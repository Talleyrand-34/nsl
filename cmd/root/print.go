
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
/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd_root

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd"
)

// printCmd represents the print command
var PrintCmd = &cobra.Command{
	Use:   "print",
	Short: "Gets information about the network",
	Long:  `Print info about any table in the db`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("print called")
	},
}

func init() {
	cmd.RootCmd.AddCommand(PrintCmd)
}
