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
