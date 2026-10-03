// SPDX-License-Identifier: AGPL-3.0-or-later
/*
  Copyright © 2025 Talleyrand-34 (t34@t34.dev)

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
Copyright © 2025 Talleyrand-34 (t34@t34.dev)
*/
package cmd_root

import (
	"fmt"
	"os"

	"nsl-graph/cmd"

	"github.com/spf13/cobra"

	srv "nsl-graph/internal/api"
	"nsl-graph/internal/observ"

	c "nsl-graph/cmd"
)

var (
	port      int
	logLevel  string
	logFormat string
	logFile   string
)

// printCmd represents the print command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP API server",
	Long:  `Print info about any table in the db`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := observ.Init(logLevel, logFormat, logFile); err != nil {
			fmt.Fprintln(os.Stderr, "Error configuring logging:", err)
			os.Exit(1)
		}
		srv.StartServer(c.Srcdbpath, port)
	},
}

func init() {
	cmd.RootCmd.AddCommand(serverCmd)
	f := serverCmd.Flags()
	f.IntVar(&port, "port", 8080, "Port where the service will be exposed")
	f.StringVar(&logLevel, "log-level", "info", "Log level: debug|info|warn|error")
	f.StringVar(&logFormat, "log-format", "text", "Log format: text|json")
	f.StringVar(&logFile, "log-file", "", "Also write logs to this file (in addition to stdout)")
}
