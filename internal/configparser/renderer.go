// renderer.go: the write-side mirror of ConfigParser.
//
// Every vendor already knows how to READ config (Fetch + ParseConfig). This package
// gives them the ability to WRITE it back — a Diff and an Apply that speak the device's
// native grammar via the same Session interface.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package configparser

import (
	"fmt"
	"sort"
)

// ---------------------------------------------------------------------------
// Error type
// ---------------------------------------------------------------------------

// ErrUnsupported means no renderer is registered for the requested OS type.
//
// Callers in cmd/push check this explicitly and surface it as a clean user message
// rather than a panicky nil-pointer on the renderer.
type ErrUnsupported struct {
	OS     string // the OS type the caller asked for
	Reason string // why: xml-patch complexity, FGT-specific admin model, etc.
}

func (e ErrUnsupported) Error() string {
	return fmt.Sprintf("configparser: render unsupported for %s: %s", e.OS, e.Reason)
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

// ConfigChange describes one line of difference between two configs.
type ConfigChange struct {
	Kind  string      // "add", "delete", "set"
	Path  string      // human-readable path; "eth0 VLAN 10 tagged"
	Old   string      // current value; empty when Kind == "add"
	New   string      // intended value; empty when Kind == "delete"
	Patch []string    // raw commands / lines the renderer would send (empty if not yet rendered)
}

// ConfigRenderer is the write-side counterpart to ConfigParser.
//
// It mirrors the read path: Diff says WHAT changed, Render says HOW to apply it.
// The Transport layer (how we reach the device) is shared — the renderer uses the
// same Session interface and SSHTransport that fetch_test.go already exercises.
//
// See internal/configparser/safety.go for the safety gate contract.
type ConfigRenderer interface {
	GetOsType() string                                // mirrors ConfigParser.GetOsType
	SupportsDevice(d any) bool                        // mirrors ConfigParser.SupportsDevice
	Diff(intended *ConfigData, observed *ConfigData) []ConfigChange
	Render(safety SafetyLevel, intended *ConfigData, sess Session, creds SSHCredentials) error
}

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

// RendererRegistry holds all registered ConfigRenderders.
type RendererRegistry struct {
	renderers map[string]ConfigRenderer
}

// NewRendererRegistry creates a new renderer registry with an empty map.
func NewRendererRegistry() *RendererRegistry {
	return &RendererRegistry{
		renderers: make(map[string]ConfigRenderer),
	}
}

// RegisterRenderer adds a renderer to the registry under its canonical os-type key.
func (r *RendererRegistry) RegisterRenderer(renderer ConfigRenderer) {
	r.renderers[renderer.GetOsType()] = renderer
}

// GetRenderer returns the appropriate renderer for an OS type, accepting common
// vendor-name aliases (e.g. "mikrotik" → "routeros").
//
// Uses the existing osTypeAliases map so the operator sees exactly the same aliases
// for read and write.
func (r *RendererRegistry) GetRenderer(osType string) (ConfigRenderer, bool) {
	canonical := resolveOsType(osType)
	rdr, exists := r.renderers[canonical]
	return rdr, exists
}

// GetRendererOrError is like GetRenderer but returns the OS type as a structured
// error instead of a bare false. This is what cmd/push/* calls — callers branch
// on errors.Is(err, ErrUnsupported{}) without needing a magic boolean flag.
func (r *RendererRegistry) GetRendererOrError(osType string) (ConfigRenderer, error) {
	renderer, exists := r.GetRenderer(osType)
	if !exists {
		return nil, ErrUnsupported{OS: osType, Reason: "no renderer registered"}
	}
	return renderer, nil
}

// ListRenderers returns all registered renderer OS types in sorted order.
func (r *RendererRegistry) ListRenderers() []string {
	out := make([]string, len(r.renderers))

	i := 0
	for k := range r.renderers {
		out[i] = k
		i++
	}
	sort.Strings(out)
	return out
}

// DefaultRendererRegistry is the global registry. Renderers register via init().

// Renderers returns the underlying map for callers that need to enumerate
// every registered renderer (e.g. the push Engine).
func (r *RendererRegistry) Renderers() map[string]ConfigRenderer {
	return r.renderers
}

var DefaultRendererRegistry = NewRendererRegistry()
