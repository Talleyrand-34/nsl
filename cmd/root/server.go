/*
  Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU Affero General Public License as published
  by the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  GNU Affero General Public License for more details.

  You should have received a copy of the GNU Affero General Public License
  along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd_root

import (
	"fmt"

	"nsl-graph/cmd"

	"github.com/spf13/cobra"

	srv "nsl-graph/internal/api"

	c "nsl-graph/cmd"
)

var port int

// printCmd represents the print command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Gets information about the network",
	Long:  `Print info about any table in the db`,
	Run: func(cmd *cobra.Command, args []string) {
		// start http server
		fmt.Println("Starting HTTP server...")
		srv.StartServer(c.Srcdbpath, port)
	},
}

func init() {
	cmd.RootCmd.AddCommand(serverCmd)
	serverCmd.Flags().
		IntVar(&port, "port", 8080, "Port where the service will be exposed")
}
