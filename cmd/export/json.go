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
package cmdexport

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// exportJsonCmd represents the port command
var exportJsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Export all network data as JSON",
	Long: `Specify a brand which consists on a name.

		A brand is the commercial name of a hardware provider
		`,
	Run: func(cmd *cobra.Command, args []string) {
		// flagNames := []string{"outPath", "outFile", "outImage"}
		// vals := util.Flagproc(
		// 	cmd,
		// 	flagNames,
		// )
		// op := vals[0]
		// of := vals[1]
		// oi := vals[2]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		exportjson := service.ExportAllStructs()
		fmt.Println(string(exportjson))
	},
}

func init() {
	cmd.ExportCmd.AddCommand(exportJsonCmd)
}
