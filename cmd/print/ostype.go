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
package cmd_print

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// devicesCmd represents the devices command
var ostypePrintCmd = &cobra.Command{
	Use:   "ostype",
	Short: "Print the OS types",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}

		ostype, err := service.GetOsTypes()
		if err != nil {
			fmt.Println("Error getting OsType:", err)
			return
		}
		jsonBytes, err := json.MarshalIndent(ostype, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling OsType to JSON:", err)
			return
		}
		fmt.Println(string(jsonBytes))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(ostypePrintCmd)
}
