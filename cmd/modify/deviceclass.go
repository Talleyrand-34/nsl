package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// brandModCmd represents the port command
var deviceclassModCmd = &cobra.Command{
	Use:   "deviceclass",
	Short: "deviceclass modifications subcommand",
	Long: `Specify the class of a device which consists on a name.

		A deviceclass is the type of device for example router, switch...`,
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
		err = service.AddDeviceClass(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing deviceclass: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(deviceclassModCmd)

	deviceclassModCmd.Flags().
		String("name", "", "Sets the name of the deviceclass")
}
