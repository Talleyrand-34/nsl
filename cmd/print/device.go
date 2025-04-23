package cmd_print

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// devicesCmd represents the devices command
var devicePrintCmd = &cobra.Command{
	Use:   "device",
	Short: "Print the devices",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		devices := service.GetDevices()
		fmt.Println(string(devices))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(devicePrintCmd)
}
