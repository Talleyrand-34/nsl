// preview.go: `nsl-graph push preview --device <id>`.
//
// Read-only. Produces the rendered patch (Diff + per-change Patch lines) so an
// operator can eyeball what would happen before committing. There is no --apply
// flag — preview IS the dry-run.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package cmd_push

import (
	"fmt"

	"github.com/spf13/cobra"
)

var PreviewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Show the rendered patch without applying it (read-only)",
	Long: `Preview shows the diff between the device's current config and the
intended one, formatted as a human-readable patch. It never opens an SSH
session that could mutate the device.

Example:
  nsl-graph push preview --device R1`,
	Run: func(c *cobra.Command, args []string) {
		deviceID, _ := c.Flags().GetString("device")
		if deviceID == "" {
			_ = c.Usage()
			fmt.Println("\nerror: --device is required")
			return
		}
		fmt.Printf("would preview device %q (phase 5 stub)\n", deviceID)
	},
}

func init() {
	PreviewCmd.Flags().String("device", "", "ID of the device to preview (required)")
	PreviewCmd.MarkFlagRequired("device")
}
