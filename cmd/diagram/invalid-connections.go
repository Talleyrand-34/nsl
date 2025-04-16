package cmd_diagram

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
)

// testCmd represents the test command
var invalidCmd = &cobra.Command{
	Use:   "invalid",
	Short: "Generates diagram with invalid connections printed",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		// ProcessCmd(getAllConnectionsInfo)
	},
}

func init() {
	cmd.DiagramCmd.AddCommand(invalidCmd)
}
