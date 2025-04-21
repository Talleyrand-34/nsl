package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// zoneModCmd represents the port command
var zoneModCmd = &cobra.Command{
	Use:   "zone",
	Short: "zone modifications subcommand",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"name", "father", "proprietary", "zonetype"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		name := vals[0]
		father := vals[1]
		proprietary := vals[2]
		zonetype := vals[3]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddZone(name, father, proprietary, zonetype)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing zone: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(zoneModCmd)

	zoneModCmd.Flags().
		String("name", "", "Sets the name of the zone")
	zoneModCmd.Flags().
		String("father", "", "Sets the name of the zone")
	zoneModCmd.Flags().
		String("proprietary", "", "Sets the name of the zone")
	zoneModCmd.Flags().
		String("zonetype", "", "Sets the name of the zone")
}
