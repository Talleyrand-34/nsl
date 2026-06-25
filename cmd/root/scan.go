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
package cmd_root

import (
	"github.com/spf13/cobra"

	"nsl-graph/cmd"
)

var ScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Query network devices via SNMP to collect inventory",
	Long: `Query network devices using SNMP to collect deterministic inventory:
device names, interfaces, MAC addresses, and VLAN assignments.

Requires SNMP to be enabled on target devices (default community: public).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cmd.RootCmd.AddCommand(ScanCmd)
}
