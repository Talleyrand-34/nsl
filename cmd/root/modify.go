/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd_root

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd"
)

// ModifyCmd represents the modify command
var ModifyCmd = &cobra.Command{
	Use:   "modify",
	Short: "Make modificatons into the network structure",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("modify called")
	},
}

func init() {
	cmd.RootCmd.AddCommand(ModifyCmd)
}
