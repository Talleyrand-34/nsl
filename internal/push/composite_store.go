// composite_store.go: FetchAndSnapshot orchestrator.
// Fetches raw bytes from a device, persists them to the backup store, and
// writes a ConfigSnapshot row. This is the entry point every push uses to
// materialise the historical record.
package push

import (
	"fmt"
	"time"

	"nsl-graph/internal/configparser"
)

// FetchFn captures raw config bytes and returns (raw, ext, error).
type FetchFn func() (raw []byte, ext string, err error)

// ParseFn parses raw bytes into a ConfigData.
type ParseFn func(os string, raw []byte) (*configparser.ConfigData, error)

// FetchAndSnapshot fetches, saves the raw backup, and writes a ConfigSnapshot.
// Returns the snapshot ID. The runID is used as both the push run ID and the
// snapshot ID for a one-to-one mapping.
func FetchAndSnapshot(deviceID, os string, fetch FetchFn, parse ParseFn, store *Store, runID string) (string, error) {
	raw, ext, err := fetch()
	if err != nil {
		return "", fmt.Errorf("fetch: %w", err)
	}

	// Save raw backup file.
	relpath, err := store.Backup.Save(deviceID, runID, ext, raw)
	if err != nil {
		return "", fmt.Errorf("backup save: %w", err)
	}

	// Retention: keep last 50 backups per device.
	store.Backup.KeepN(deviceID, 50)

	// Parse config (optional — parse errors are non-fatal for snapshot).
	var parsedCfg any
	if parse != nil {
		cfg, err := parse(os, raw)
		if err == nil {
			parsedCfg = cfg
		}
	}

	// Write snapshot row.
	snap := ConfigSnapshot{
		ID:             runID,
		DeviceID:       deviceID,
		OS:             os,
		CapturedAt:     time.Now(),
		CapturedByRun:  runID,
		ParserVersion:  "v1",
		Sha256Raw:      Sha256(raw),
		RawExt:         ext,
		BackupRelpath: relpath,
		Config:         parsedCfg,
	}
	id, err := store.Snap.Save(snap)
	if err != nil {
		return "", fmt.Errorf("snapshot save: %w", err)
	}
	return id, nil
}
