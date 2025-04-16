package cmd_print

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
)

// devicesCmd represents the devices command
var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Print the devices",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// p, _ := db.DeviceRepository.FetchDevices()
		// fmt.Println(p)
	},
}

func init() {
	cmd.PrintCmd.AddCommand(devicesCmd)

	devicesCmd.Flags().
		BoolP("all-model-ports", "a", false, "Print the device with all the ports, not only those used")
	devicesCmd.Flags().
		BoolP("info-ports", "i", false, "Print the device with the ports used")
	devicesCmd.Flags().
		BoolP("possible-ports", "p", false, "Print the cartesian product of the possible ports of the model and the device ports")
}
