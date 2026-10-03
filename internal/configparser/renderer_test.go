// renderer_test.go: TDD for the ConfigRenderer interface, the
// DefaultRendererRegistry and ErrUnsupported.
//
// The renderer mirrors ConfigParser: it knows WHAT to say to a device and what to
// compute as a diff. It does not know HOW to reach the device -- that is the
// Transport's job, and it is reused as-is from transport.go.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package configparser

import (
	"errors"
	"fmt"
	"testing"
)

// stubRenderer is the minimum a ConfigRenderer must implement. Used by registry
// tests; OpenWrt's real renderer ships in Phase 2.
type stubRenderer struct{ os string }

func (s stubRenderer) GetOsType() string                                         { return s.os }
func (s stubRenderer) SupportsDevice(d any) bool                                { return true }
func (s stubRenderer) Diff(intended, observed *ConfigData) []ConfigChange       { return nil }
func (s stubRenderer) Render(safety SafetyLevel, intended *ConfigData, sess Session, creds SSHCredentials) error {
	return nil
}

func TestErrUnsupported_StructureAndIs(t *testing.T) {
	err := ErrUnsupported{OS: "opnsense", Reason: "xml-patch complexity"}

	if err.OS != "opnsense" {
		t.Errorf("ErrUnsupported.OS = %q, want %q", err.OS, "opnsense")
	}
	if err.Reason == "" {
		t.Errorf("ErrUnsupported.Reason must not be empty; callers use it to tell the user *why*")
	}

	msg := err.Error()
	if msg == "" {
		t.Fatal("ErrUnsupported.Error() must not return empty string")
	}
	if !contains(msg, "opnsense") {
		t.Errorf("Error() = %q, must contain OS name", msg)
	}
	if !contains(msg, "xml-patch complexity") {
		t.Errorf("Error() = %q, must contain Reason", msg)
	}

	// errors.As must match a wrapped ErrUnsupported so cmd/push can branch on the type.
	wrapped := fmt.Errorf("render: %w", err)
	var target ErrUnsupported
	if !errors.As(wrapped, &target) {
		t.Fatalf("errors.As on wrapped ErrUnsupported = false; got target=%v", target)
	}
	if target.OS != "opnsense" {
		t.Errorf("errors.As target.OS = %q, want opnsense", target.OS)
	}
}

func TestRendererRegistry_RegisterAndLookup(t *testing.T) {
	r := NewRendererRegistry()
	r.RegisterRenderer(stubRenderer{os: "openwrt"})

	got, ok := r.GetRenderer("openwrt")
	if !ok {
		t.Fatal("GetRenderer(openwrt) returned !ok after RegisterRenderer(stubRenderer{openwrt})")
	}
	if got.GetOsType() != "openwrt" {
		t.Errorf("renderer's GetOsType() = %q, want openwrt", got.GetOsType())
	}
}

func TestRendererRegistry_AliasLookup(t *testing.T) {
	// Reuses the same osTypeAliases map the parser registry uses. A user who
	// types "mikrotik" must get the same renderer as "routeros".
	r := NewRendererRegistry()
	r.RegisterRenderer(stubRenderer{os: "routeros"})

	got, ok := r.GetRenderer("mikrotik")
	if !ok {
		t.Fatal("alias 'mikrotik' must resolve to the 'routeros' renderer (reuses parser aliases)")
	}
	if got.GetOsType() != "routeros" {
		t.Errorf("alias lookup returned %q, want routeros", got.GetOsType())
	}

	got, ok = r.GetRenderer("ROS") // case-insensitive
	if !ok || got.GetOsType() != "routeros" {
		t.Errorf("alias 'ROS' must resolve case-insensitively to routeros")
	}
}

func TestRendererRegistry_UnknownReturnsErrUnsupported(t *testing.T) {
	r := NewRendererRegistry()
	r.RegisterRenderer(stubRenderer{os: "openwrt"})

	got, err := r.GetRendererOrError("netscaler")
	if err == nil {
		t.Fatal("GetRendererOrError(netscaler) returned no error; unknown OS must surface as ErrUnsupported")
	}
	if got != nil {
		t.Errorf("GetRendererOrError on unknown returned non-nil renderer: %v", got)
	}
	var unsup ErrUnsupported
	if !errors.As(err, &unsup) {
		t.Fatalf("err is not ErrUnsupported: %v", err)
	}
	if unsup.OS != "netscaler" {
		t.Errorf("ErrUnsupported.OS = %q, want netscaler", unsup.OS)
	}
}

func TestDefaultRendererRegistry_IsCallable(t *testing.T) {
	// Phase 2+ will register the OpenWrt renderer into the default registry.
	// For now, the registry must exist and be non-nil.
	if DefaultRendererRegistry == nil {
		t.Fatal("DefaultRendererRegistry must be wired (var, not lazy) so init() ordering cannot leak a nil registry")
	}
	if got := DefaultRendererRegistry.ListRenderers(); got == nil {
		t.Errorf("ListRenderers() returned nil slice; callers range over it expecting len()==0 to be safe")
	}
}

// contains is the tiny helper ErrUnsupported.Error() tests rely on. stdlib has
// strings.Contains but the import loop is annoying for a single test file.
func contains(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
