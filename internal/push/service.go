// SPDX-License-Identifier: MIT
// service.go: the canonical entry point for the push pipeline.
//
// Service is the single surface the HTTP API and the cobra CLI both call
// into. It owns the renderer registry (via Engine), the audit trail (via
// SafetyFloor + PushRepository), the credential vault (via
// CredentialResolver), and a per-device mutex so concurrent pushes to the
// same device can't race.
//
// Phase 1 of webui-integration.md: this file declares the surface and
// stubs each method with a panic. Subsequent phases fill in the bodies
// one method at a time, each with its own TDD cycle.
package push

import (
	"context"
	"sync"
	"time"

	"nsl-graph/internal/configparser"
)

// Service is the canonical push primitive. Construct once at API server
// startup; share across all HTTP handlers. The cobra CLI gets its own
// Service instance per process.
type Service struct {
	repo       PushRepository      // runs + snapshots + backups
	engine     *Engine             // renderer registry; vendor-agnostic
	floor      *SafetyFloor        // backup + snapshot + audit
	creds      CredentialResolver  // SSH + API credentials, vault-backed
	defaultOS  func(deviceID string) (string, error) // resolves a device's OS type
	rendererName string             // "typed" or "lazy"; informational, audit row value

	// perDevLocks serialises pushes to the same device. key = deviceID,
	// value = *sync.Mutex. A sync.Map so reads (the common case — no lock
	// held) are lock-free.
	perDevLocks sync.Map
}

// PushRequest is the canonical input for a single push attempt.
type PushRequest struct {
	DeviceID string
	OS       string
	Safety   configparser.SafetyLevel
	Intent   *configparser.ConfigData
	Renderer string
}

// PushResult is the canonical output. Populated whether the push
// succeeded or failed; the operator inspects ExitStatus + ErrorString.
type PushResult struct {
	RunID         string
	SnapshotID    string
	BackupRelpath string
	DiffText      string // same as Preview output; empty on failure-before-diff
	ExitStatus    string // "success" | "failed"
	ErrorString   string
	StartedAt     time.Time
	FinishedAt    time.Time
}

// SnapshotInfo is the lightweight shape returned by History. We
// deliberately don't include the full *ConfigData — that comes via
// GetSnapshot or the device-detail page.
type SnapshotInfo struct {
	ID            string    `json:"id"`
	DeviceID      string    `json:"device_id"`
	OS            string    `json:"os"`
	CapturedAt    time.Time `json:"captured_at"`
	CapturedByRun string    `json:"captured_by_run"`
	BackupRelpath string    `json:"backup_relpath"`
	Interfaces    int       `json:"interfaces_count"`
}

// CredentialResolver abstracts SSH + API credential lookup. Production
// reads from the vault; tests return fixed creds.
type CredentialResolver interface {
	// SSH returns the SSH credentials for a device. Implementations may
	// consult scan profiles, the vault, or both.
	SSH(deviceID string) (configparser.SSHCredentials, error)
	// API returns the REST credentials (base URL + key + secret) for an
	// OPNsense device. Returns ErrCredentialsUnavailable when the
	// device has no API profile.
	API(deviceID, os string) (baseURL, key, secret string, err error)
}

// ErrCredentialsUnavailable is returned when no credentials can be
// resolved for a device. The HTTP layer translates this to 503 with a
// hint to set up a profile first.
type ErrCredentialsUnavailable struct {
	DeviceID string
	OS       string
}

func (e *ErrCredentialsUnavailable) Error() string {
	return "push: no credentials available for device " + e.DeviceID + " (os=" + e.OS + ")"
}

// NewService builds the canonical Service. Caller provides the
// repository, the engine (use NewDefaultEngine), a CredentialResolver
// (typically vault-backed), and a function that maps a device ID to its
// OS type (so we can pick the right renderer).
func NewService(
	repo PushRepository,
	engine *Engine,
	creds CredentialResolver,
	defaultOS func(deviceID string) (string, error),
) *Service {
	return &Service{
		repo:         repo,
		engine:       engine,
		floor:        nil, // wired in phase 2 when SafetyFloor.NewSafetyFloorWithStore lands
		creds:        creds,
		defaultOS:    defaultOS,
		rendererName: "typed",
	}
}

// deviceLock returns the per-device mutex, creating one on first use.
// Stored as *sync.Mutex inside the sync.Map so callers can defer Unlock.
func (s *Service) deviceLock(deviceID string) *sync.Mutex {
	v, _ := s.perDevLocks.LoadOrStore(deviceID, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// ---------------------------------------------------------------------------
// Method stubs — phase 2 onwards fills these in one at a time, each with
// its own TDD cycle. Until then, calling them panics.
// ---------------------------------------------------------------------------

// Push is the canonical entry point: snapshot the device, write the
// snapshot, run diff, run render, commit per-area, audit. Safety=DryRun
// skips the network round-trip entirely.
func (s *Service) Push(ctx context.Context, req PushRequest) (PushResult, error) {
	panic("Service.Push not yet implemented; see webui-integration.md phase 7")
}

// Preview returns the rendered patch text without snapshotting or
// applying. Always read-only.
//
// Two overloads:
//   - Preview(ctx, deviceID, os)             — convenience; loads observed
//     from the latest snapshot, requires an intent to be supplied via
//     PreviewIntent. Phase 3 wires this.
//   - PreviewDiff(ctx, deviceID, os, intent, observed) — explicit; uses
//     caller-supplied configs. The HTTP layer composes these.
func (s *Service) Preview(ctx context.Context, deviceID, os string) (string, error) {
	panic("Service.Preview (no-args) not yet implemented; see webui-integration.md phase 3")
}

// PreviewDiff returns the rendered patch for an explicit
// (intent, observed) pair. Used by HTTP handlers that have both in
// hand. Read-only.
func (s *Service) PreviewDiff(ctx context.Context, deviceID, os string, intent, observed *configparser.ConfigData) (string, error) {
	return s.engine.Preview(os, intent, observed)
}

// History lists the most recent `limit` snapshots for a device, newest
// first. limit<=0 returns every snapshot.
func (s *Service) History(ctx context.Context, deviceID string, limit int) ([]SnapshotInfo, error) {
	panic("Service.History not yet implemented; see webui-integration.md phase 5")
}

// Rollback loads a snapshot and re-renders it via the typed renderer.
func (s *Service) Rollback(ctx context.Context, deviceID, snapshotID string) (PushResult, error) {
	panic("Service.Rollback not yet implemented; see webui-integration.md phase 8")
}

// LiveConfig fetches the device's current running config and returns
// the parsed *ConfigData.
func (s *Service) LiveConfig(ctx context.Context, deviceID, os string) (*configparser.ConfigData, error) {
	panic("Service.LiveConfig not yet implemented; see webui-integration.md phase 3")
}

// FetchAndSnapshot is the unified fetch+parse+save entry point used by
// both the snapshot-on-push path and the device-detail page.
func (s *Service) FetchAndSnapshot(ctx context.Context, deviceID, os, runID string) (string, error) {
	panic("Service.FetchAndSnapshot not yet implemented; see webui-integration.md phase 4")
}

// Status returns the most recent push result for a device. Lightweight
// poll target for the device-row status indicator.
func (s *Service) Status(ctx context.Context, deviceID string) (PushResult, error) {
	panic("Service.Status not yet implemented; see webui-integration.md phase 9")
}