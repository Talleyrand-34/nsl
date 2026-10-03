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

// rootCmd represents the base command when called without any subcommands
var DiagramCmd = &cobra.Command{
	Use:   "diagram",
	Short: "Generate a network diagram",
	Long: `Generate a network diagram. Use a subcommand to choose the focus:
"connection" or "port" (add --vlan for VLAN coloring).

By default an SVG is rendered. With --ascii the same diagram is rendered as
ASCII art: it is printed to stdout and also written to a .txt file alongside the
.d2 source. Use --charset ascii for plain characters instead of box-drawing.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("diagram called - use a specific subcommand to do diagrams")
		cmd.Help()
	},
}

func init() {
	cmd.RootCmd.AddCommand(DiagramCmd)
	// Persistent flags shared by all diagram subcommands.
	DiagramCmd.PersistentFlags().Bool("ascii", false, "Render ASCII art (printed to stdout and written to a .txt file) instead of SVG")
	DiagramCmd.PersistentFlags().String("charset", "unicode", "ASCII charset when --ascii is set: 'unicode' (box-drawing) or 'ascii' (plain + - |)")
}
