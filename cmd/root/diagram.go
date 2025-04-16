package cmd_root

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd"
)

// rootCmd represents the base command when called without any subcommands
var DiagramCmd = &cobra.Command{
	Use:   "diagram",
	Short: "Generates a diagram from an nsl especification",
	Long:  `.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
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
