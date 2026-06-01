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
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Srcdbpath  Path to the db
	Srcdbpath string
	// Backend  Database backend type (cloverdb)
	Backend  string
	outpath  string
	outFile  string
	outImage string
	// Verbose verbose
	Verbose bool
	// Debug debug
	Debug bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "nsl-graph",
	Short:   "Manages networks specifications in nsl",
	Long:    `i`,
	Version: "0.1v",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().
		StringVarP(&Srcdbpath, "source", "s", "test.db", "database file or directory path")
	RootCmd.PersistentFlags().
		StringVarP(&Backend, "backend", "b", "cloverdb", "database backend type (cloverdb)")
	RootCmd.PersistentFlags().
		StringVar(&outpath, "outPath", "out/", "output path for files")
	RootCmd.PersistentFlags().
		StringVar(&outFile, "outFile", "out.d2", "output script filepath")
	RootCmd.PersistentFlags().
		StringVar(&outImage, "outImage", "out.svg", "output image filepath")
	RootCmd.PersistentFlags().
		BoolVarP(&Verbose, "verbose", "v", false, "Display more verbose output in console output. (default: false)")
	RootCmd.PersistentFlags().
		BoolVarP(&Debug, "debug", "d", false, "Display debugging output in the console. (default: false)")
}

func GetSrcDB() string {
	return Srcdbpath
}
