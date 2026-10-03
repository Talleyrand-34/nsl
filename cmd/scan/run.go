// SPDX-License-Identifier: AGPL-3.0-or-later
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
package cmd_scan

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
	q "nsl-graph/internal/repository/application"
	s "nsl-graph/internal/scanner"
)

var (
	runMethod      string
	runTarget      string
	runCommunity   string
	runSNMPVersion string
	runSNMPPort    uint16
	runTimeout     int
	runProfile     string
	runOsType      string
	runHuman       bool
	runOutput      string
	runDefaultZone string
)

var runScanCmd = &cobra.Command{
	Use:   "run [target]",
	Short: "Unified scan that emits an editable import plan (mirrors the web UI)",
	Long: `Discover devices the same way the web "Import devices" page does, then emit an
editable import plan as JSON. This is the CLI half of the scan → edit → import
round-trip:

  nsl-graph scan run --method ssh --target 10.0.2.0/24 --profile owrt --os-type openwrt > plan.json
  $EDITOR plan.json          # adjust subnet / vlan_number / suggested name / zone
  nsl-graph scan import plan.json

The method (snmp|ssh) is explicit; single host vs subnet sweep is inferred from
the target (a bare IP is one host; a CIDR or comma/space-separated list is a
batch). Output is JSON on stdout by default (status goes to stderr, so it pipes
cleanly); pass -H/--human for a readable summary plus interactive review/import.

Examples:
  nsl-graph scan run 10.0.2.245
  nsl-graph scan run --method snmp --target 10.0.2.0/24 --community public > plan.json
  nsl-graph scan run --method ssh --target 10.0.2.0/24 --profile owrt --os-type openwrt > plan.json
  nsl-graph scan run 10.0.2.245 -H        # human summary + interactive import`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := runTarget
		if len(args) == 1 && target == "" {
			target = args[0]
		}
		if target == "" {
			fmt.Fprintln(os.Stderr, "Error: a target IP, CIDR, or list is required (positional arg or --target)")
			os.Exit(1)
		}

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		opts := q.RunScanOptions{
			Method:      runMethod,
			Target:      target,
			Community:   runCommunity,
			SNMPVersion: runSNMPVersion,
			SNMPPort:    runSNMPPort,
			TimeoutSec:  runTimeout,
			Profile:     runProfile,
			OsType:      runOsType,
		}

		devs, err := service.RunScan(opts, stderrScanEmitter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
			os.Exit(1)
		}

		options := s.ImportOptions{
			AutoImport:        true,
			CreateZones:       true,
			DefaultZone:       runDefaultZone,
			SkipExisting:      true,
			InteractiveVLANs:  true,
			VLANAccuracyLevel: 2,
		}

		if err := emitOrReviewDevices(service, devs, runHuman, runOutput, options); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd_root.ScanCmd.AddCommand(runScanCmd)
	f := runScanCmd.Flags()
	f.StringVar(&runMethod, "method", "snmp", "Scan method: snmp or ssh")
	f.StringVar(&runTarget, "target", "", "Target IP, CIDR, or comma/space-separated list (alternative to positional arg)")
	f.StringVar(&runCommunity, "community", "", "SNMP community (falls back to a matching profile, else 'public')")
	f.StringVar(&runSNMPVersion, "snmp-version", "", "SNMP version (v1|v2c)")
	f.Uint16Var(&runSNMPPort, "snmp-port", 0, "SNMP UDP port")
	f.IntVar(&runTimeout, "timeout", 0, "Per-host timeout in seconds")
	f.StringVar(&runProfile, "profile", "", "Saved scan profile supplying credentials/parameters (required for ssh)")
	f.StringVar(&runOsType, "os-type", "", "OS type for ssh scans (overrides the profile's; required for generic profiles)")
	f.StringVar(&runDefaultZone, "default-zone", "Discovered", "Default zone used when importing in --human mode")
	f.BoolVarP(&runHuman, "human", "H", false, "Human-readable summary + interactive review (default output is the plan JSON)")
	f.StringVarP(&runOutput, "output", "o", "", "Write the plan JSON to this file instead of stdout")
}
