package cmd_print

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// devicesCmd represents the devices command
var zonePrintCmd = &cobra.Command{
	Use:   "zone",
	Short: "Print the zones",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		zones := service.GetZones()
		fmt.Println(string(zones))
	},
}

func init() {
	cmd.PrintCmd.AddCommand(zonePrintCmd)
}
