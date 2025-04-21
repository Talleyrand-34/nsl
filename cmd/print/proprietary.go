package cmd_print

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// devicesCmd represents the devices command
var proprietaryPrintCmd = &cobra.Command{
	Use:   "proprietary",
	Short: "Print the proprietarys",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		util.PrintStringArrayPrettyJson(service.GetProperties())
	},
}

func init() {
	cmd.PrintCmd.AddCommand(proprietaryPrintCmd)
}
