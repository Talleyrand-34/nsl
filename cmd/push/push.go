// push.go: cobra command tree for `nsl-graph push`.
//
// Three subcommands:
//   - push device --device <id> [--apply]
//   - push devices --zone <id> [--apply]
//   - push preview --device <id>   (read-only)
//
// Every command defaults to dry-run. --apply is required for any mutation.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package cmd_push

import (
	"github.com/spf13/cobra"

	cmd "nsl-graph/cmd"
)

var PushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push a configuration to one or more devices (dry-run by default)",
	Long: `Push a configuration to network devices.

Every push passes a SafetyLevel gate. The contract is:
  dry-run (default): report the diff; never modify the device
  staged:            write the patch file to the device; do not reload
  apply:             write, syntax-check, reload, audit-log

A miscompiled call that drops --apply defaults to dry-run, never to apply.`,
}

func init() {
	cmd.RootCmd.AddCommand(PushCmd)
	PushCmd.AddCommand(DeviceCmd)
	PushCmd.AddCommand(DevicesCmd)
	PushCmd.AddCommand(PreviewCmd)
}

// resolveSafety maps the --safety flag to its string value, defaulting to dry-run
// when the string is empty (the cobra zero-value case).
func resolveSafety(s string) string {
	switch s {
	case "staged", "apply":
		return s
	default:
		return "dry-run"
	}
}
