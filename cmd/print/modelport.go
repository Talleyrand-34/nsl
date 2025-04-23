package cmd_print

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// devicesCmd represents the devices command
var modelPortPrintCmd = &cobra.Command{
	Use:   "modelport",
	Short: "Print the modelPorts",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		fmt.Println(string(service.GetModelPorts()))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(modelPortPrintCmd)
}
