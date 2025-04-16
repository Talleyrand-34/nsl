/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd_root

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmd "nsl-graph/cmd"
)

// configCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("config called")
		fmt.Println("Configuration values:")
		for _, key := range viper.AllKeys() {
			fmt.Printf("%s: %v\n", key, viper.Get(key))
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(ConfigCmd)
}
