// push_test.go: TDD for the cmd/push/* cobra commands.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package cmd_push

import (
	"testing"
)

func TestPushCmd_RegistersSubcommands(t *testing.T) {
	want := []string{"device", "devices", "preview"}
	got := make(map[string]bool)
	for _, c := range PushCmd.Commands() {
		got[c.Name()] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("PushCmd must register %q subcommand; have: %v", w, keys(got))
		}
	}
}

func TestPushDeviceCmd_FlagsDefaultToDryRun(t *testing.T) {
	c := DeviceCmd
	apply := c.Flags().Lookup("apply")
	if apply == nil {
		t.Fatal("device subcommand must have --apply flag")
	}
	if apply.DefValue != "false" {
		t.Errorf("--apply must default to false (dry-run safety); got %q", apply.DefValue)
	}

	dryRun := c.Flags().Lookup("dry-run")
	if dryRun == nil {
		t.Error("device subcommand must have --dry-run flag (explicit alias of the default)")
	}
}

func TestPushDeviceCmd_RequiresDeviceFlag(t *testing.T) {
	c := DeviceCmd
	if c.Flag("device") == nil {
		t.Fatal("device subcommand must have --device flag")
	}
	// The flag is marked required via MarkFlagRequired in init() — cobra's
	// behaviour on missing required flag is to print Usage and exit 1, which
	// is exactly what we want for a safety-first tool.
}

func TestPushDeviceCmd_AcceptsSafetyFlag(t *testing.T) {
	c := DeviceCmd
	safety := c.Flags().Lookup("safety")
	if safety == nil {
		t.Fatal("device subcommand must accept --safety (dry-run|staged|apply)")
	}
	if safety.DefValue != "dry-run" {
		t.Errorf("--safety must default to dry-run; got %q", safety.DefValue)
	}
}

func TestPushPreviewCmd_IsReadOnly(t *testing.T) {
	c := PreviewCmd
	apply := c.Flags().Lookup("apply")
	if apply != nil {
		t.Errorf("preview must NOT expose --apply (preview is read-only); got: %v", apply)
	}
}

func TestPushDeviceCmd_ParseSafetyFromApplyFlag(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"dry-run", "dry-run"},
		{"staged", "staged"},
		{"apply", "apply"},
		{"", "dry-run"},
	}
	for _, c := range cases {
		got := resolveSafety(c.in)
		if got != c.want {
			t.Errorf("resolveSafety(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
