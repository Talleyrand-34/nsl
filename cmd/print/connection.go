package cmd_print

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// connectionsCmd represents the devices command
var connectionPrintCmd = &cobra.Command{
	Use:   "connection",
	Short: "Print the connections",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		connections := service.GetConnections()
		fmt.Println(string(connections))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(connectionPrintCmd)
}
