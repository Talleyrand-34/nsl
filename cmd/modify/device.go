package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// zoneModCmd represents the port command
var DeviceModCmd = &cobra.Command{
	Use:   "device",
	Short: "device modifications subcommand",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"name", "model", "zoneid", "zonename", "proprietary"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		label := vals[0]
		model := vals[1]
		zoneid := vals[2]
		zonename := vals[3]
		proprietary := vals[4]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddDevice(label, model, zoneid, zonename, proprietary)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing zone: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(DeviceModCmd)

	DeviceModCmd.Flags().
		String("name", "", "Sets the name of the zone")
	DeviceModCmd.Flags().
		String("model", "", "Sets the fatherzone by name if there is")
	DeviceModCmd.Flags().
		String("zoneid", "", "Sets the fatherzone by id if there is")
	DeviceModCmd.Flags().
		String("zonename", "", "Sets the proprietary of the zone")
	DeviceModCmd.Flags().
		String("proprietary", "", "Sets the zonetype of the zone")
}
