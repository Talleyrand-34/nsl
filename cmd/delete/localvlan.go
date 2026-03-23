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

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

var localVlanDelCmd = &cobra.Command{
	Use:   "localvlan [local_vlan_id]",
	Short: "Delete a local VLAN",
	Long: `Delete a local VLAN entry by ID.
Use --by-device to delete all local VLANs for a given device ID.
Use --by-vlan to delete all local VLANs for a given global VLAN ID.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		byDevice, _ := cmd.Flags().GetBool("by-device")
		byVlan, _ := cmd.Flags().GetBool("by-vlan")

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			return fmt.Errorf("error connecting to database: %w", err)
		}

		switch {
		case byDevice:
			err = service.DeleteLocalVlansByDevice(id)
			if err == nil {
				fmt.Printf("All local VLANs for device '%s' deleted successfully\n", id)
			}
		case byVlan:
			err = service.DeleteLocalVlansByVlanID(id)
			if err == nil {
				fmt.Printf("All local VLANs for global VLAN '%s' deleted successfully\n", id)
			}
		default:
			err = service.DeleteLocalVlan(id)
			if err == nil {
				fmt.Printf("Local VLAN '%s' deleted successfully\n", id)
			}
		}

		if err != nil {
			return fmt.Errorf("error deleting local VLAN: %w", err)
		}
		return nil
	},
}

func init() {
	cmd_root.DeleteCmd.AddCommand(localVlanDelCmd)
	localVlanDelCmd.Flags().Bool("by-device", false, "Delete all local VLANs belonging to the given device ID")
	localVlanDelCmd.Flags().Bool("by-vlan", false, "Delete all local VLANs associated with the given global VLAN ID")
}
