package cmd_root

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd"
	util "nsl-graph/cmd/utils"
	"nsl-graph/internal/format"
)

// rootCmd represents the base command when called without any subcommands
var DiagramCmd = &cobra.Command{
	Use:   "diagram",
	Short: "Generates a diagram from an nsl especification",
	Long:  `.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"outPath", "outFile", "outImage"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		op := vals[0]
		of := vals[1]
		oi := vals[2]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		connections := service.GetConnections()
		devices := service.GetDevices()
		d2diagram := format.GenerateD2FromJSON(devices, connections)
		format.WriteDiagram(d2diagram, op, of, oi)
	},
}

func init() {
	cmd.RootCmd.AddCommand(DiagramCmd)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.modtest.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
