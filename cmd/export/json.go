package cmd_export

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// exportJsonCmd represents the port command
var exportJsonCmd = &cobra.Command{
	Use:   "json",
	Short: "brand modifications subcommand",
	Long: `Specify a brand which consists on a name.

		A brand is the commercial name of a hardware provider
		`,
	Run: func(cmd *cobra.Command, args []string) {
		// flagNames := []string{"outPath", "outFile", "outImage"}
		// vals := util.Flagproc(
		// 	cmd,
		// 	flagNames,
		// )
		// op := vals[0]
		// of := vals[1]
		// oi := vals[2]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		exportjson := service.ExportAllStructs()
		fmt.Println(string(exportjson))
	},
}

func init() {
	cmd.ExportCmd.AddCommand(exportJsonCmd)
}
