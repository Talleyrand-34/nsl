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
package cmd_delete

import (
	"fmt"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// vlanDeleteCmd represents the vlan delete command
var vlanDeleteCmd = &cobra.Command{
	Use:   "vlan",
	Short: "Delete a VLAN",
	Long: `Delete a VLAN by its database ID. Use --cascade to also remove VLAN configurations from all device ports.

Example:
  nsl-graph delete vlan --id 1
  nsl-graph delete vlan --id 1 --cascade`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the required VLAN database ID
		vlanDbId, err := cmd.Flags().GetString("id")
		if err != nil {
			return fmt.Errorf("error reading flag 'id': %w", err)
		}
		if vlanDbId == "" {
			return fmt.Errorf("VLAN database ID is required. Use --id flag")
		}

		cascade, _ := cmd.Flags().GetBool("cascade")

		// Get service connection
		service, err := util.ServiceConnection()
		if err != nil {
			return fmt.Errorf("error connecting to service: %w", err)
		}

		// Delete the VLAN
		if cascade {
			err = service.DeleteVlanCascade(vlanDbId)
			if err == nil {
				fmt.Printf("Successfully deleted VLAN with database ID %s and all port configurations\n", vlanDbId)
			}
		} else {
			err = service.DeleteVlan(vlanDbId)
			if err == nil {
				fmt.Printf("Successfully deleted VLAN with database ID %s\n", vlanDbId)
			}
		}

		if err != nil {
			return fmt.Errorf("error deleting VLAN: %w", err)
		}
		return nil
	},
}

func init() {
	cmd.DeleteCmd.AddCommand(vlanDeleteCmd)
	vlanDeleteCmd.Flags().Bool("cascade", false, "Remove VLAN configurations from all device ports")

	vlanDeleteCmd.Flags().String("id", "", "Database ID of the VLAN to delete (required)")
	vlanDeleteCmd.MarkFlagRequired("id")
}
