package cmd_diagram

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
)

// testCmd represents the test command
var validCmd = &cobra.Command{
	Use:   "valid",
	Short: "Ensures only valid connections are printed",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		// ProcessCmd(getAllValidConnectionsInfo)
	},
}

func init() {
	cmd.DiagramCmd.AddCommand(validCmd)
	cmd.DiagramCmd.Flags().
		BoolP("all-model-ports", "a", false, "Print the device with all the ports, not only those used")
}
