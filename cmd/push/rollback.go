// rollback.go: `nsl-graph push rollback` — re-render a snapshot onto a device.
package cmd_push

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"nsl-graph/internal/configparser"
	"nsl-graph/internal/push"
)

var rollbackStore *push.Store
var rollbackEngine *push.Engine

var RollbackCmd = &cobra.Command{
	Use:   "rollback --device <id> --to <snapshot-id>",
	Short: "Re-render an old snapshot onto a device",
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceID, _ := cmd.Flags().GetString("device")
		snapshotID, _ := cmd.Flags().GetString("to")

		if deviceID == "" {
			return fmt.Errorf("--device is required")
		}
		if snapshotID == "" {
			return fmt.Errorf("--to <snapshot-id> is required")
		}

		if rollbackStore == nil || rollbackStore.Snap == nil {
			return fmt.Errorf("snapshot store not configured")
		}
		if rollbackEngine == nil {
			return fmt.Errorf("engine not configured")
		}

		// Load the snapshot.
		snap, err := rollbackStore.Snap.Get(snapshotID)
		if err != nil {
			return fmt.Errorf("load snapshot %q: %w", snapshotID, err)
		}
		if snap.DeviceID != deviceID {
			return fmt.Errorf("snapshot %q belongs to device %q, not %q", snapshotID, snap.DeviceID, deviceID)
		}

		// The snapshot's Config is a *ConfigData (set at capture time).
		cfg, ok := snap.Config.(*configparser.ConfigData)
		if !ok || cfg == nil {
			return fmt.Errorf("snapshot %q has no usable Config for rollback", snapshotID)
		}

		// Get the renderer for this OS.
		renderer, ok := rollbackEngine.Renderers()[snap.OS]
		if !ok {
			return fmt.Errorf("no renderer registered for os %q", snap.OS)
		}

		fmt.Fprintf(os.Stderr, "Rollback: device=%s snapshot=%s os=%s\n", deviceID, snapshotID, snap.OS)
		if err := renderer.Render(configparser.SafetyApply, cfg, nil, configparser.SSHCredentials{}); err != nil {
			return fmt.Errorf("render: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Rollback complete: snapshot %q applied to %q\n", snapshotID, deviceID)
		return nil
	},
}

func init() {
	PushCmd.AddCommand(RollbackCmd)
	RollbackCmd.Flags().StringP("device", "d", "", "Device ID (required)")
	RollbackCmd.Flags().StringP("to", "t", "", "Snapshot ID to roll back to (required)")
}
