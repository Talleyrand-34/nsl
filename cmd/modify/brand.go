package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// brandModCmd represents the port command
var brandModCmd = &cobra.Command{
	Use:   "brand",
	Short: "brand modifications subcommand",
	Long: `Specify a brand which consists on a name.

		A brand is the commercial name of a hardware provider
		`,
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
		err = service.AddBrand(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing Brand: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(brandModCmd)

	brandModCmd.Flags().
		String("name", "", "Sets the name of the brand")
}
