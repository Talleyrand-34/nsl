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
	"encoding/json"
	"fmt"
	"os"

	"nsl-graph/internal/observ"
	q "nsl-graph/internal/repository/application"
	s "nsl-graph/internal/scanner"
)

// The CLI scan commands share one output convention, matching `scan connections`
// and the web UI workflow: the machine-readable artifact goes to stdout, all
// progress/status text goes to stderr (so the JSON pipes cleanly), and -H/--human
// switches to a readable summary plus interactive review/import. The canonical
// pipeline artifact is a JSON array of import plans (s.DeviceImportPlan) — exactly
// what the web UI lets the user edit before importing — so the round-trip is:
//
//	scan run ... > plan.json   # analyze → emit the editable plan
//	$EDITOR plan.json          # tweak subnet / vlan_number / name / zone
//	scan import plan.json      # execute the (edited) plan

// stderrEmitter reports scan progress and status to stderr, keeping stdout clean
// for the JSON artifact. It satisfies observ.Emitter.
type stderrEmitter struct{}

func (stderrEmitter) Emit(level, msg string, fields ...any) {
	line := fmt.Sprintf("[%s] %s", level, msg)
	for i := 0; i+1 < len(fields); i += 2 {
		line += fmt.Sprintf(" %v=%v", fields[i], fields[i+1])
	}
	fmt.Fprintln(os.Stderr, line)
}

func (stderrEmitter) Progress(done, total int) {
	if total > 0 {
		fmt.Fprintf(os.Stderr, "\rprogress: %d/%d", done, total)
		if done >= total {
			fmt.Fprintln(os.Stderr)
		}
	}
}

// analyzeDevicesToPlans turns discovered devices into editable import plans, the
// same per-device analysis the web `/scan/analyze` endpoint performs.
func analyzeDevicesToPlans(service q.NetServiceInt, devs []s.DiscoveredDevice) ([]s.DeviceImportPlan, error) {
	plans := make([]s.DeviceImportPlan, 0, len(devs))
	for _, d := range devs {
		plan, err := service.AnalyzeDeviceForImport(d)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze %s: %w", d.Device.IP, err)
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// emitScanJSON writes v as indented JSON to out (a file) or, when out is empty, to
// stdout. A trailing newline is added for stdout friendliness.
func emitScanJSON(v any, out string) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if out != "" {
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", out, err)
		}
		fmt.Fprintf(os.Stderr, "Wrote %s\n", out)
		return nil
	}
	_, err = os.Stdout.Write(append(data, '\n'))
	return err
}

// emitOrReviewDevices implements the unified device-scan output convention.
// human → readable summary + interactive review/import (reuses the import.go
// review loop); otherwise → analyze and emit the import-plan JSON (stdout or out).
func emitOrReviewDevices(service q.NetServiceInt, devs []s.DiscoveredDevice, human bool, out string, options s.ImportOptions) error {
	if len(devs) == 0 {
		fmt.Fprintln(os.Stderr, "No devices discovered.")
		return nil
	}
	if human {
		return importDevicesWithInteractiveVLANMapping(service, devs, options)
	}
	plans, err := analyzeDevicesToPlans(service, devs)
	if err != nil {
		return err
	}
	return emitScanJSON(plans, out)
}

// executePlans imports a list of (possibly user-edited) import plans. By default
// each plan is executed as-is (non-interactive); with human it drops into the
// same Approve/Edit/Skip/Quit review used elsewhere. Status goes to stderr.
func executePlans(service q.NetServiceInt, plans []s.DeviceImportPlan, options s.ImportOptions, human bool) error {
	for _, plan := range plans {
		name := plan.Device.SuggestedName
		if name == "" {
			name = plan.Device.Device.IP
		}
		if human {
			fmt.Fprintf(os.Stderr, "\n=== %s (%s) ===\n", name, plan.Device.Device.IP)
			_ = displayImportPlan(plan)
			action, modified, err := getUserImportDecision(plan)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting user input: %v\n", err)
				continue
			}
			switch action {
			case s.ActionApprove:
				plan = modified
			case s.ActionSkip, s.ActionSkipDevice:
				fmt.Fprintf(os.Stderr, "Skipping %s\n", name)
				continue
			case s.ActionQuit:
				fmt.Fprintln(os.Stderr, "Import cancelled by user.")
				return nil
			}
		}
		if err := service.ExecuteApprovedImportPlan(plan, options); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to import %s: %v\n", name, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "Imported %s\n", name)
	}
	return nil
}

// stderrScanEmitter is the shared emitter instance for CLI scans.
var stderrScanEmitter observ.Emitter = stderrEmitter{}
