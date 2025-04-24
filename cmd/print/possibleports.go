package cmd_print

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// devicesCmd represents the devices command
var possiblePortPrintCmd = &cobra.Command{
	Use:   "possibleports",
	Short: "Print the possiblePorts",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		deviceid, err := cmd.Flags().GetString("deviceid")
		if err != nil {

			fmt.Fprintf(os.Stderr, "Error reading flag '%v': %v\n", "all", err)
			os.Exit(1)
		}
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		if deviceid == "" {
			fmt.Println(string(service.GetAllPortsAll()))
		} else {
			fmt.Println(string(service.GetAllPortsDevice(deviceid)))
		}
	},
}

func init() {
	cmd.PrintCmd.AddCommand(possiblePortPrintCmd)
	possiblePortPrintCmd.Flags().
		String("deviceid", "", "Sets the name of the zone")
}
