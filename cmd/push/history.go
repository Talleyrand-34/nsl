// history.go: `nsl-graph push history` — list snapshot history for a device.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package cmd_push

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"nsl-graph/internal/push"
)

var historyStore *push.Store

var HistoryCmd = &cobra.Command{
	Use:   "history --device <id>",
	Short: "List snapshot history for a device",
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceID, _ := cmd.Flags().GetString("device")
		if deviceID == "" {
			return fmt.Errorf("--device is required")
		}

		if historyStore == nil || historyStore.Snap == nil {
			return fmt.Errorf("snapshot store not configured")
		}

		const limit = 50
		snaps, err := historyStore.Snap.List(deviceID, limit)
		if err != nil {
			return fmt.Errorf("list snapshots: %w", err)
		}

		if len(snaps) == 0 {
			fmt.Fprintf(os.Stderr, "No snapshots found for device %q\n", deviceID)
			return nil
		}

		fmt.Fprintf(os.Stdout, "Snapshots for device %q (%d most recent):\n", deviceID, len(snaps))
		fmt.Fprintf(os.Stdout, "%-12s  %-10s  %s\n", "ID", "OS", "Captured At")
		fmt.Fprintln(os.Stdout, "------------  ----------  ------------------------")
		for _, s := range snaps {
			fmt.Fprintf(os.Stdout, "%-12s  %-10s  %s\n",
				s.ID, s.OS, s.CapturedAt.Format("2006-01-02 15:04:05"))
		}
		return nil
	},
}

func init() {
	PushCmd.AddCommand(HistoryCmd)
	HistoryCmd.Flags().StringP("device", "d", "", "Device ID (required)")
}
