
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
package cmd_modify

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

// zoneModCmd represents the port command
var zoneModCmd = &cobra.Command{
	Use:   "zone",
	Short: "zone modifications subcommand",
	Long:  `Specify zone which consists on a name, father`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNames := []string{"name", "father", "proprietary", "zonetype", "fatherid"}
		vals := util.Flagproc(
			cmd,
			flagNames,
		)
		name := vals[0]
		father := vals[1]
		proprietary := vals[2]
		zonetype := vals[3]
		fatherid := vals[4]
		service, err := util.ServiceConnection()
		if err != nil {
			return
		}
		err = service.AddZone(name, fatherid, father, proprietary, zonetype)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing zone: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.ModifyCmd.AddCommand(zoneModCmd)

	zoneModCmd.Flags().
		String("name", "", "Sets the name of the zone")
	zoneModCmd.Flags().
		String("father", "", "Sets the fatherzone by name if there is by name")
	zoneModCmd.Flags().
		String("fatherid", "", "Sets the fatherzone by id if there is (prioritized over name)")
	zoneModCmd.Flags().
		String("proprietary", "", "Sets the proprietary of the zone by name")
	zoneModCmd.Flags().
		String("zonetype", "", "Sets the zonetype of the zone by name")
}
