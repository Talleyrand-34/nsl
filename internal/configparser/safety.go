// safety.go: the gate every push must pass.
//
// SafetyLevel governs whether a renderer.Render call (or any future cmd/push/* code path)
// is allowed to touch the device. The contract is:
//
//   - SafetyDryRun (zero value): report the diff and the rendered patch; never open SSH
//     for an apply, never reload config, never `uci commit`.
//
//   - SafetyStaged: write the rendered patch to a temp path on the device and report it,
//     but do not reload. Lets the operator eyeball the artifact before saying `--apply`.
//
//   - SafetyApply: write, syntax-check, reload, audit-log. This is the only level that
//     mutates the running configuration.
//
// The zero value MUST be SafetyDryRun. A miscompiled caller that drops the flag falls
// back to "report only", not "apply silently". See internal/configparser/safety_test.go
// for the test that pins this contract.
package configparser

import (
	"fmt"
	"strings"
)

// SafetyLevel is the gate every push passes through.
//
// Order is load-bearing: do not reorder these. Higher levels are a superset of lower
// levels' effects, and `String()` reports them in that order.
type SafetyLevel int

const (
	// SafetyDryRun is the zero value. Default to no apply.
	SafetyDryRun SafetyLevel = iota
	// SafetyStaged writes the rendered patch to the device but does not reload.
	SafetyStaged
	// SafetyApply writes, syntax-checks, reloads, and audit-logs.
	SafetyApply
)

// String renders the level for logs, errors and CLI output.
func (s SafetyLevel) String() string {
	switch s {
	case SafetyDryRun:
		return "dry-run"
	case SafetyStaged:
		return "staged"
	case SafetyApply:
		return "apply"
	default:
		return "unknown"
	}
}

// ParseSafetyLevel parses the three flag spellings plus their common aliases.
//
// The empty string parses as SafetyDryRun, not as an error: an HTTP handler that did
// not receive a `safety` field should report, not apply.
func ParseSafetyLevel(s string) (SafetyLevel, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "dry-run", "dryrun":
		return SafetyDryRun, nil
	case "staged":
		return SafetyStaged, nil
	case "apply":
		return SafetyApply, nil
	default:
		return SafetyDryRun, fmt.Errorf("configparser: unknown safety level %q (want dry-run|staged|apply)", s)
	}
}
