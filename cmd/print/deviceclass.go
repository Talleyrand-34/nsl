package cmd_print

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// devicesCmd represents the devices command
var deviceclassPrintCmd = &cobra.Command{
	Use:   "deviceclass",
	Short: "Print the deviceclasss",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		fmt.Println(string(service.GetDeviceClasses()))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(deviceclassPrintCmd)
}
