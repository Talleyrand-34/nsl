// diff_engine_test.go: TDD for the push diff engine.
//
// The engine is the bridge between the read side (intented ConfigData) and the
// renderer's Diff(). Tests use a fake renderer so the engine logic itself is
// under test, not the per-vendor grammar.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package push

import (
	"errors"
	"strings"
	"testing"

	"nsl-graph/internal/configparser"
)

// fakeRenderer records what was Diffed and what was Rendered so the engine's
// wiring is observable in tests.
type fakeRenderer struct {
	os        string
	diffs     []configparser.ConfigChange
	gotIntent *configparser.ConfigData
	gotObs    *configparser.ConfigData
}

func (f *fakeRenderer) GetOsType() string { return f.os }
func (f *fakeRenderer) SupportsDevice(d any) bool { return true }
func (f *fakeRenderer) Diff(intent *configparser.ConfigData, obs *configparser.ConfigData) []configparser.ConfigChange {
	f.gotIntent = intent
	f.gotObs = obs
	return f.diffs
}
func (f *fakeRenderer) Render(safety configparser.SafetyLevel, intent *configparser.ConfigData, sess configparser.Session, creds configparser.SSHCredentials) error {
	return nil
}

func TestEngine_PicksRendererByOsType(t *testing.T) {
	r := &fakeRenderer{os: "openwrt"}
	engine := NewEngine(map[string]configparser.ConfigRenderer{"openwrt": r})

	_, err := engine.Diff("openwrt", &configparser.ConfigData{}, &configparser.ConfigData{})
	if err != nil {
		t.Fatalf("Diff returned: %v", err)
	}
	if r.gotIntent == nil {
		t.Error("engine did not pass intended ConfigData to the renderer")
	}
	if r.gotObs == nil {
		t.Error("engine did not pass observed ConfigData to the renderer")
	}
}

func TestEngine_UnknownOsReturnsErrUnsupported(t *testing.T) {
	engine := NewEngine(map[string]configparser.ConfigRenderer{})
	_, err := engine.Diff("netscaler", &configparser.ConfigData{}, &configparser.ConfigData{})
	if err == nil {
		t.Fatal("engine must return ErrUnsupported for unknown OS")
	}
	var unsup configparser.ErrUnsupported
	if !errors.As(err, &unsup) {
		t.Errorf("err is not ErrUnsupported: %v", err)
	}
	if unsup.OS != "netscaler" {
		t.Errorf("ErrUnsupported.OS = %q, want netscaler", unsup.OS)
	}
}

func TestEngine_PreviewReturnsRenderedPatch(t *testing.T) {
	diffs := []configparser.ConfigChange{
		{Kind: "add", Path: "br-lan VLAN 30", New: "30", Patch: []string{`uci add_list network.br_lan.vlan_members="30 t"`}},
		{Kind: "delete", Path: "br-lan VLAN 20", Old: "20", Patch: []string{`uci del_list network.br_lan.vlan_members="20"`}},
	}
	r := &fakeRenderer{os: "openwrt", diffs: diffs}
	engine := NewEngine(map[string]configparser.ConfigRenderer{"openwrt": r})

	patch, err := engine.Preview("openwrt", &configparser.ConfigData{}, &configparser.ConfigData{})
	if err != nil {
		t.Fatalf("Preview returned: %v", err)
	}
	if !strings.Contains(patch, "uci add_list network.br_lan.vlan_members=\"30 t\"") {
		t.Errorf("Preview missing add line; got: %q", patch)
	}
	if !strings.Contains(patch, "uci del_list network.br_lan.vlan_members=\"20\"") {
		t.Errorf("Preview missing del line; got: %q", patch)
	}
}

func TestEngine_PreviewReturnsEmptyOnNoChanges(t *testing.T) {
	r := &fakeRenderer{os: "openwrt", diffs: nil}
	engine := NewEngine(map[string]configparser.ConfigRenderer{"openwrt": r})

	patch, err := engine.Preview("openwrt", &configparser.ConfigData{}, &configparser.ConfigData{})
	if err != nil {
		t.Fatalf("Preview returned: %v", err)
	}
	if patch != "" {
		t.Errorf("Preview with no diffs must be empty, got: %q", patch)
	}
}

func TestEngine_DefaultRegistryIsUsedWhenNoOverrides(t *testing.T) {
	// Without an explicit registry map, the engine picks up the global registry.
	// This is the call path cmd/push/* uses. The opnsense entry is the lazy
	// renderer: Diff() computes the patch (no network), Render() refuses
	// because the registry has no typed client — the engine is expected to
	// construct the typed renderer for the lab path.
	r := NewDefaultEngine()
	changes, err := r.Diff("opnsense", &configparser.ConfigData{}, &configparser.ConfigData{})
	if err != nil {
		t.Fatalf("Diff(opnsense) on lazy registry entry must succeed (Diff is offline); got err: %v", err)
	}
	_ = changes // patch shape is renderer-specific; engine contract is no-error
}

// fakeTypedOpnsenseRenderer is a recognisable stand-in for the typed
// opnsense renderer the engine constructs when env vars are present. We
// don't reach into the real opnsense.Client here — we only need a marker
// the engine can install so callers see "the typed renderer is in place".
type fakeTypedOpnsenseRenderer struct{ marked bool }

func (f *fakeTypedOpnsenseRenderer) GetOsType() string { return "opnsense" }
func (f *fakeTypedOpnsenseRenderer) SupportsDevice(any) bool { return true }
func (f *fakeTypedOpnsenseRenderer) Diff(*configparser.ConfigData, *configparser.ConfigData) []configparser.ConfigChange { return nil }
func (f *fakeTypedOpnsenseRenderer) Render(configparser.SafetyLevel, *configparser.ConfigData, configparser.Session, configparser.SSHCredentials) error { return nil }

func TestEngine_DefaultRegistry_RespectsOpnsenseEnvOverride(t *testing.T) {
	// The handoff calls for a typed opnsense renderer wired from env vars.
	// Without a live OPNsense we substitute a marker renderer and verify the
	// engine installs it in place of the lazy one.
	t.Setenv("OPNSENSE_URL", "https://opnsense.local")
	t.Setenv("OPNSENSE_KEY", "k")
	t.Setenv("OPNSENSE_SECRET", "s")
	override := WithOpnsenseFactory(func(_, _, _ string) configparser.ConfigRenderer {
		return &fakeTypedOpnsenseRenderer{marked: true}
	})
	e := NewDefaultEngine(override)

	r, ok := e.renderers["opnsense"]
	if !ok {
		t.Fatal("opnsense renderer must be registered in default engine")
	}
	if _, typed := r.(*fakeTypedOpnsenseRenderer); !typed {
		t.Errorf("opnsense renderer must be the typed one when env vars are set; got %T", r)
	}
}

func TestEngine_DefaultRegistry_NoEnvVars_KeepsLazyRenderer(t *testing.T) {
	t.Setenv("OPNSENSE_URL", "")
	t.Setenv("OPNSENSE_KEY", "")
	t.Setenv("OPNSENSE_SECRET", "")

	e := NewDefaultEngine()
	r, ok := e.renderers["opnsense"]
	if !ok {
		t.Fatal("opnsense renderer must be registered even without env vars (lazy entry)")
	}
	if _, typed := r.(*fakeTypedOpnsenseRenderer); typed {
		t.Errorf("opnsense renderer must NOT be the typed one without env vars; got %T", r)
	}
}
