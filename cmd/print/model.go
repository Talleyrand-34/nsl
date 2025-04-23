package cmd_print

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// devicesCmd represents the devices command
var modelDevicePrintCmd = &cobra.Command{
	Use:   "model",
	Short: "Print the modelDevices",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		fmt.Println(string(service.GetModels()))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(modelDevicePrintCmd)
}
