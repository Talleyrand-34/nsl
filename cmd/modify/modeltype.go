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

// brandModCmd represents the port command
var modeltypeModCmd = &cobra.Command{
	Use:   "modeltype",
	Short: "Add a model type",
	Long: `Specify the class of a device which consists of a name.

		A modeltype is the type of device for example router, switch...`,
	Run: func(cmd *cobra.Command, args []string) {
		flag := "name"
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag '%v': %v\n", flag, err)
			os.Exit(1)
		}
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddModelType(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing modeltype: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.AddCmd.AddCommand(modeltypeModCmd)

	modeltypeModCmd.Flags().
		String("name", "", "Sets the name of the modeltype")
}
