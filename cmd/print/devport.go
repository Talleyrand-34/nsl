package cmd_print

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// devicePortsCmd represents the devices command
var devicePortPrintCmd = &cobra.Command{
	Use:   "devicePort",
	Short: "Print the devicePorts",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		devicePorts := service.GetDevicePorts()
		fmt.Println(string(devicePorts))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(devicePortPrintCmd)
}
