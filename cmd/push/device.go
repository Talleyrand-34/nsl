// device.go: `nsl-graph push device --device <id> [--apply]`.
//
// One device at a time. The engine loads the intended ConfigData from the DB
// (Phase 4 + YANG projection), fetches the observed one over SSH, runs the
// renderer.Diff and either reports (dry-run) or applies the patch.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package cmd_push

import (
	"fmt"

	"github.com/spf13/cobra"
)

var DeviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Push a configuration to a single device",
	Long: `Push to a single device. Defaults to --dry-run.

Examples:
  # Preview the diff for a device (no SSH traffic):
  nsl-graph push device --device R1

  # Apply the change (writes, syntax-checks, reloads, audit-logs):
  nsl-graph push device --device R1 --apply`,
	Run: func(c *cobra.Command, args []string) {
		deviceID, _ := c.Flags().GetString("device")
		safety, _ := c.Flags().GetString("safety")
		fmt.Printf("would push to device %q at safety=%s (phase 5 stub)\n", deviceID, resolveSafety(safety))
	},
}

func init() {
	DeviceCmd.Flags().String("device", "", "ID of the device to push (required)")
	DeviceCmd.MarkFlagRequired("device")
	DeviceCmd.Flags().String("safety", "dry-run", "safety level: dry-run|staged|apply")
	DeviceCmd.Flags().Bool("apply", false, "alias for --safety=apply (never the default)")
	DeviceCmd.Flags().Bool("dry-run", false, "alias for --safety=dry-run (the default)")
}
