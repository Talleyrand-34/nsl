package cmd_modify

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
)

// portCmd represents the port command
var portCmd = &cobra.Command{
	Use:   "port",
	Short: "Port modifications subcommand",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(portCmd)

	portCmd.Flags().
		String("devid", "", "Print the device with all the ports, not only those used")
	portCmd.Flags().
		String("portname", "", "Print the device with the ports used")

	portCmd.Flags().
		String("modelid", "", "Print the device with all the ports, not only those used")
}
