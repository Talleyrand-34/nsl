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
	Long:  `Generate a D2/SVG network diagram. Use a subcommand to choose the focus: "connection" or "port" (add --vlan for VLAN coloring).`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("diagram called - use a specific subcommand to do diagrams")
		cmd.Help()
	},
}

func init() {
	cmd.RootCmd.AddCommand(DiagramCmd)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.modtest.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
