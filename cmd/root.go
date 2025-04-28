
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
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	Srcdbpath string
	outpath   string
	outFile   string
	outImage  string
	Verbose   bool
	Debug     bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "nsl-graph",
	Short: "Manages networks specifications in nsl",
	Long:  `i`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		message := viper.GetString("message")
		fmt.Println(message)
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
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().
		StringVarP(&cfgFile, "config-file", "c", "", "Specify the config file")
	RootCmd.PersistentFlags().
		StringVarP(&Srcdbpath, "source", "s", "test.db", "database file")
	viper.BindPFlag("source", RootCmd.PersistentFlags().Lookup("source"))
	RootCmd.PersistentFlags().
		StringVar(&outpath, "outPath", "out/", "output path for files")
	viper.BindPFlag("outPath", RootCmd.PersistentFlags().Lookup("outPath"))
	RootCmd.PersistentFlags().
		StringVar(&outFile, "outFile", "out.d2", "output script filepath")
	viper.BindPFlag("outFile", RootCmd.PersistentFlags().Lookup("outFile"))
	RootCmd.PersistentFlags().
		StringVar(&outImage, "outImage", "out.svg", "output image filepath")
	viper.BindPFlag("outImage", RootCmd.PersistentFlags().Lookup("outImage"))
	RootCmd.PersistentFlags().
		BoolVarP(&Verbose, "verbose", "v", false, "Display more verbose output in console output. (default: false)")
	viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	RootCmd.PersistentFlags().
		BoolVarP(&Debug, "debug", "d", false, "Display debugging output in the console. (default: false)")
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
}

func GetSrcDB() string {
	return Srcdbpath
}

func initConfig() {
	if cfgFile != "" {
		// Use specified config file
		viper.SetConfigFile(cfgFile)
	} else {
		// Default config file settings
		viper.SetConfigName("config") // Name without extension
		viper.AddConfigPath(".")      // Current directory
		viper.AddConfigPath("$HOME")  // Home directory
	}

	viper.AutomaticEnv() // Read environment variables

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
	}
}
