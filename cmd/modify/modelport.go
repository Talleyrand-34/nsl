package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// modelDeviceModCmd represents the port command
var modelPortModCmd = &cobra.Command{
	Use:   "modelport",
	Short: "model modifications subcommand",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"name", "posx", "posy", "modelname"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		name := vals[0]
		posx := vals[1]
		posy := vals[2]
		modelname := vals[3]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddModelPort(name, posx, posy, modelname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing modelPort: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(modelPortModCmd)

	modelPortModCmd.Flags().
		String("name", "", "Sets the name of the modelPort")
	modelPortModCmd.Flags().
		String("posx", "", "Sets the name of the modelPort")
	modelPortModCmd.Flags().
		String("posy", "", "Sets the name of the modelPort")
	modelPortModCmd.Flags().
		String("modelname", "", "Sets the name of the modelPort")
}
