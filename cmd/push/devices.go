// devices.go: `nsl-graph push devices --zone <id> [--apply]`.
//
// All devices in a zone. Same safety gate as `push device`. Useful when a
// zone-wide change is the deployment unit (e.g. re-IP an access LAN).
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package cmd_push

import (
	"fmt"

	"github.com/spf13/cobra"
)

var DevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Push a configuration to every device in a zone",
	Long: `Push to every device in a zone. Defaults to --dry-run.

Examples:
  # Preview the diff for every device in a zone:
  nsl-graph push devices --zone access-lan

  # Apply (one device failure stops the whole run by default):
  nsl-graph push devices --zone access-lan --apply`,
	Run: func(c *cobra.Command, args []string) {
		zone, _ := c.Flags().GetString("zone")
		if zone == "" {
			_ = c.Usage()
			fmt.Println("\nerror: --zone is required")
			return
		}
		safety, _ := c.Flags().GetString("safety")
		fmt.Printf("would push to devices in zone %q at safety=%s (phase 5 stub)\n", zone, resolveSafety(safety))
	},
}

func init() {
	DevicesCmd.Flags().String("zone", "", "zone ID (required)")
	DevicesCmd.MarkFlagRequired("zone")
	DevicesCmd.Flags().String("safety", "dry-run", "safety level: dry-run|staged|apply")
	DevicesCmd.Flags().Bool("apply", false, "alias for --safety=apply (never the default)")
}
