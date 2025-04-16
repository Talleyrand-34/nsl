package cmd_print

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
	q "nsl-graph/internal/repository/application"
)

// devicesCmd represents the devices command
var brandCmd = &cobra.Command{
	Use:   "brand",
	Short: "Print the brands",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		repository, err := util.RepositoryConnection()
		if err != nil {
			log.Fatalf("Error connecting to the database: %v", err)
		}
		service := q.NewNetService(repository)
		fmt.Println("This are the Brands: ", service.GetBrands())
	},
}

func init() {
	cmd.PrintCmd.AddCommand(brandCmd)
}
