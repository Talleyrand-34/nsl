// SPDX-License-Identifier: MIT
// openwrt_renderer_test.go: TDD for the OpenWrt UCI renderer's route path.
package parsers

import (
	"strings"
	"testing"

	"nsl-graph/internal/configparser"
)

// TestOpenWrtRenderer_DiffRoutes_RouteAddChange pins the diff shape.
func TestOpenWrtRenderer_DiffRoutes_RouteAddChange(t *testing.T) {
	intended := &configparser.ConfigData{
		Routes: []configparser.ConfigRoute{{Network: "10.0.0.0/8", Gateway: "192.168.1.1"}},
	}
	diffs := diffRoutes(intended, nil)
	if len(diffs) != 1 {
		t.Fatalf("want 1 diff, got %d", len(diffs))
	}
	if diffs[0].Kind != "route-add" {
		t.Errorf("kind=%q, want route-add", diffs[0].Kind)
	}
	if !strings.Contains(diffs[0].Patch[0], "network=10.0.0.0/8") {
		t.Errorf("patch missing network: %v", diffs[0].Patch)
	}
}

// TestOpenWrtRenderer_RouteAddCommands pins the UCI command shape so any
// change to the per-vendor router Apply path is caught.
func TestOpenWrtRenderer_RouteAddCommands(t *testing.T) {
	r := NewOpenWrtRenderer()
	d := configparser.ConfigChange{
		Kind: "route-add",
		Path: "10.0.50.0/24",
		New:  "10.0.50.0/24",
		Old:  "10.0.0.1",
		Patch: []string{
			"network=10.0.50.0/24",
			"gateway=10.0.0.1",
			"interface=wan",
		},
	}
	cmds, err := r.routeAddCommands(d)
	if err != nil {
		t.Fatalf("routeAddCommands: %v", err)
	}
	want := []string{
		"uci set network.route_10_0_50_0_24=route",
		"uci set network.route_10_0_50_0_24.target=10.0.50.0/24",
		"uci set network.route_10_0_50_0_24.gateway=10.0.0.1",
		"uci set network.route_10_0_50_0_24.device=wan",
		"uci commit network",
	}
	if !equalSlices(cmds, want) {
		t.Errorf("commands mismatch:\nwant: %v\ngot:  %v", want, cmds)
	}
}

// TestOpenWrtRenderer_RouteDelCommands pins the delete path.
func TestOpenWrtRenderer_RouteDelCommands(t *testing.T) {
	r := NewOpenWrtRenderer()
	d := configparser.ConfigChange{
		Kind: "route-del",
		Path: "10.0.50.0/24",
	}
	cmds, err := r.routeDelCommands(d)
	if err != nil {
		t.Fatalf("routeDelCommands: %v", err)
	}
	want := []string{
		"uci del network.route_10_0_50_0_24",
		"uci commit network",
	}
	if !equalSlices(cmds, want) {
		t.Errorf("commands mismatch:\nwant: %v\ngot:  %v", want, cmds)
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}