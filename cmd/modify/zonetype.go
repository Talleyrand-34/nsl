package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// zonetypeModCmd represents the port command
var zonetypeModCmd = &cobra.Command{
	Use:   "zonetype",
	Short: "zonetype modifications subcommand",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		flag := "name"
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag '%v': %v\n", flag, err)
			os.Exit(1)
		}
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddZoneType(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing zonetype: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(zonetypeModCmd)

	zonetypeModCmd.Flags().
		String("name", "", "Sets the name of the zonetype")
}
