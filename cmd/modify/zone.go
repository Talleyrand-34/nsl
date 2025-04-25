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
	Long:  `Specify zone which consists on a name, father`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"name", "father", "proprietary", "zonetype", "fatherid"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		name := vals[0]
		father := vals[1]
		proprietary := vals[2]
		zonetype := vals[3]
		fatherid := vals[4]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddZone(name, fatherid, father, proprietary, zonetype)
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
		String("father", "", "Sets the fatherzone by name if there is by name")
	zoneModCmd.Flags().
		String("fatherid", "", "Sets the fatherzone by id if there is (prioritized over name)")
	zoneModCmd.Flags().
		String("proprietary", "", "Sets the proprietary of the zone by name")
	zoneModCmd.Flags().
		String("zonetype", "", "Sets the zonetype of the zone by name")
}
