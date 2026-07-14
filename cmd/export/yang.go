/*
Copyright © 2026 Talleyrand-34 (t34@t34.dev)

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
package cmdexport

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
)

var exportYangStrict bool

var exportYangCmd = &cobra.Command{
	Use:   "yang",
	Short: "Export the specification as RFC 7951 JSON (YANG)",
	Long: `Export the whole specification as RFC 7951 JSON, conforming to the standard
YANG models: ietf-network and ietf-network-topology (RFC 8345), ietf-l2-topology
(RFC 8944) and IEEE 802.1Q, augmented by nsl-topology and nsl-inventory.

Devices are nodes, device ports are termination points, and each connection becomes
two links -- RFC 8345 links are unidirectional. The output can be validated against
the modules in internal/yang/modules:

    nsl-graph export yang -s real.db > topology.json
    yanglint -t config -p internal/yang/modules \
        internal/yang/modules/ietf-network.yang \
        internal/yang/modules/ietf-network-topology.yang \
        internal/yang/modules/ietf-l2-topology.yang \
        internal/yang/modules/nsl-topology.yang \
        internal/yang/modules/nsl-inventory.yang \
        topology.json

Values the schema will not accept -- a VLAN outside 1..4094, a reference to an owner
that does not exist -- are dropped and reported on stderr, so the emitted document
always validates. Those warnings are worth reading: they list what the hand-rolled
schema has been quietly tolerating. Use --strict to fail on them instead.`,
	RunE: func(c *cobra.Command, args []string) error {
		service, err := util.ServiceConnection()
		if err != nil {
			return fmt.Errorf("connecting to the database: %w", err)
		}

		out, warnings, err := service.ExportYANG()
		if err != nil {
			return err
		}

		for _, w := range warnings {
			fmt.Fprintln(os.Stderr, "warning:", w)
		}
		if exportYangStrict && len(warnings) > 0 {
			return fmt.Errorf("%d value(s) rejected by the schema (--strict)", len(warnings))
		}

		fmt.Println(string(out))
		return nil
	},
}

func init() {
	exportYangCmd.Flags().BoolVar(&exportYangStrict, "strict", false,
		"fail if any value had to be dropped to satisfy the schema")
	cmd.ExportCmd.AddCommand(exportYangCmd)
}
