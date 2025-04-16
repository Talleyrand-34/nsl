package cmd_modify

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
)

// connectionCmd represents the connection command
var connectionCmd = &cobra.Command{
	Use:   "connection",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(connectionCmd)

	connectionCmd.Flags().
		String("source-modelid", "", "Print the device with all the ports, not only those used")
	connectionCmd.Flags().String("dest-modelid", "", "Print the device with the ports used")
	connectionCmd.Flags().
		String("source-devid", "", "Print the device with all the ports, not only those used")
	connectionCmd.Flags().String("dest-devid", "", "Print the device with the ports used")
	connectionCmd.Flags().
		String("source-portname", "", "Print the device with all the ports, not only those used")
	connectionCmd.Flags().String("dest-portname", "", "Print the device with the ports used")
	// Mark flags as required
	// devicesCmd.MarkFlagRequired("source")
	// devicesCmd.MarkFlagRequired("dest")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// connectionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// connectionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
