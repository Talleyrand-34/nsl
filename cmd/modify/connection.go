package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// zoneModCmd represents the port command
var connectionModCmd = &cobra.Command{
	Use:   "connection",
	Short: "connection modifications subcommand",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"from-device", "from-model-port-id", "to-device", "to-model-port-id"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		fromDevice := vals[0]
		fromModelPortId := vals[1]
		toDevice := vals[2]
		toModelPortId := vals[3]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddConnection(fromDevice, fromModelPortId, toDevice, toModelPortId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing zone: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(connectionModCmd)

	connectionModCmd.Flags().
		String("from-device", "", "Sets the fatherzone by name if there is")
	connectionModCmd.Flags().
		String("to-device", "", "Sets the fatherzone by id if there is")
	connectionModCmd.Flags().
		String("from-model-port-id", "", "Sets the name of the zone")
	connectionModCmd.Flags().
		String("to-model-port-id", "", "Sets the proprietary of the zone")
}
