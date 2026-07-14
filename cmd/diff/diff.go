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

// Package cmddiff implements `nsl-graph diff`: audit the network against its
// specification.
package cmddiff

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd"
	util "nsl-graph/cmd/utils"
	"nsl-graph/internal/datastore"
	s "nsl-graph/internal/scanner"
)

var (
	diffScanFile string
	diffJSON     bool
	diffQuiet    bool
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare the network against its specification",
	Long: `Audit the network: report how what a scan actually FOUND differs from what the
specification says SHOULD be there.

It compares the two as canonical YANG trees (RFC 8345), so a difference is reported
against the same node a schema error would name:

    nsl-graph scan network --subnet 10.0.2.0/24 --json --out scan.json
    nsl-graph diff --scan scan.json -s real.db

What it reports:

    - a device in the specification that the scan did not find (off, unplugged, gone)
    - a device on the network that nobody specified (rogue, or undocumented)
    - a port missing from a device, or one nobody expected
    - a VLAN absent from a trunk, or present where it should not be
    - a VLAN whose TAGGING differs -- the port is in the right VLAN but trunking where
      it should be an access port. Frames flow, and they flow wrong, which is worse
      than not flowing at all, because everything looks up.

Cabling is not compared here: a device scan sees interfaces, not cables. Adjacency comes
from correlating LLDP/CDP/FDB across several hosts, which is what "scan connections"
does.

Exit status is 1 when the network differs from the specification, so this can gate a
pipeline. Use --quiet to get only that status.`,
	RunE: func(c *cobra.Command, args []string) error {
		if diffScanFile == "" {
			return fmt.Errorf("--scan is required: a scan-result JSON file to compare the specification against")
		}

		devices, err := loadScan(diffScanFile)
		if err != nil {
			return err
		}

		service, err := util.ServiceConnection()
		if err != nil {
			return fmt.Errorf("connecting to the database: %w", err)
		}

		changes, warnings, err := service.DiffAgainstScan(devices)
		if err != nil {
			return err
		}

		for _, w := range warnings {
			fmt.Fprintln(os.Stderr, "warning:", w)
		}

		if diffJSON {
			out, err := json.MarshalIndent(changes, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
		} else if !diffQuiet {
			report(changes)
		}

		if len(changes) > 0 {
			// Silence cobra's usage dump: a network that differs from its specification
			// is a finding, not a misuse of the command.
			c.SilenceUsage = true
			return fmt.Errorf("the network differs from its specification in %d place(s)", len(changes))
		}
		return nil
	},
}

// loadScan reads a scan-result JSON file. It accepts either the raw ScanResult object
// that `scan network --json` emits, or a bare array of devices, because both shapes are
// in circulation.
func loadScan(path string) ([]s.SNMPDevice, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading scan file: %w", err)
	}

	var result s.ScanResult
	if err := json.Unmarshal(data, &result); err == nil && len(result.Devices) > 0 {
		return result.Devices, nil
	}

	var devices []s.SNMPDevice
	if err := json.Unmarshal(data, &devices); err == nil && len(devices) > 0 {
		return devices, nil
	}

	return nil, fmt.Errorf("%s does not look like a scan result: expected a ScanResult "+
		"object or an array of devices", path)
}

func report(changes []datastore.Change) {
	if len(changes) == 0 {
		fmt.Println("The network matches its specification.")
		return
	}

	var missing, unexpected, differs int
	for _, ch := range changes {
		switch ch.Op {
		case datastore.OpMissing:
			missing++
		case datastore.OpUnexpected:
			unexpected++
		case datastore.OpDiffers:
			differs++
		}
	}

	for _, ch := range changes {
		fmt.Println(ch)
	}

	fmt.Printf("\n%d difference(s): %d intended but not observed, %d observed but not intended, %d differ.\n",
		len(changes), missing, unexpected, differs)
}

func init() {
	f := diffCmd.Flags()
	f.StringVar(&diffScanFile, "scan", "", "scan-result JSON to compare against (from `scan network --json`)")
	f.BoolVar(&diffJSON, "json", false, "emit the differences as JSON")
	f.BoolVar(&diffQuiet, "quiet", false, "print nothing; report only via exit status")

	cmd.RootCmd.AddCommand(diffCmd)
}
