package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// modelDeviceModCmd represents the port command
var modelDeviceModCmd = &cobra.Command{
	Use:   "model",
	Short: "model modifications subcommand",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"name", "brand", "class"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		name := vals[0]
		brand := vals[1]
		class := vals[2]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddModel(name, brand, class)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing modelDevice: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(modelDeviceModCmd)

	modelDeviceModCmd.Flags().
		String("name", "", "Sets the name of the modelDevice")
	modelDeviceModCmd.Flags().
		String("brand", "", "Sets the name of the modelDevice")
	modelDeviceModCmd.Flags().
		String("class", "", "Sets the name of the modelDevice")
}
