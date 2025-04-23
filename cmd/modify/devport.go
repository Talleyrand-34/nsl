package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// zoneModCmd represents the port command
var devicePortModCmd = &cobra.Command{
	Use:   "devicePort",
	Short: "devicePort modifications subcommand",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"deviceid", "modelportid"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		deviceid := vals[0]
		modelportid := vals[1]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddDevicePort(deviceid, modelportid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing zone: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(devicePortModCmd)

	devicePortModCmd.Flags().
		String("deviceid", "", "Sets the name of the zone")
	devicePortModCmd.Flags().
		String("modelportid", "", "Sets the fatherzone by name if there is")
}
