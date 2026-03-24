/*
Copyright © 2026 Talleyrand-34 (t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package format

import (
	"strings"
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// ── Test fixtures ─────────────────────────────────────────────────────────────

// twoSwitchFixture returns a minimal network: two switches in the same zone
// connected by one link. SW-A has an untagged VLAN 20 on its port; SW-B
// has only tagged VLANs.
func twoSwitchFixture() (
	zones []e.Zone,
	devices []e.Device,
	conns []e.Connection,
	dps []e.DevicePort,
	ifaces []e.DeviceInterface,
	ifacePorts []e.InterfacePort,
) {
	zones = []e.Zone{
		{ID: "zone-dc", Name: "DC"},
	}
	devices = []e.Device{
		{ID: "dev-a", Name: "SW-A", ZoneID: "zone-dc", ZoneName: "DC"},
		{ID: "dev-b", Name: "SW-B", ZoneID: "zone-dc", ZoneName: "DC"},
	}
	conns = []e.Connection{
		{
			FromDevice:    "SW-A",
			FromModelPort: "eth0",
			FromZoneID:    "zone-dc",
			ToDevice:      "SW-B",
			ToModelPort:   "eth0",
			ToZoneID:      "zone-dc",
		},
	}
	dps = []e.DevicePort{
		{DeviceID: "dev-a", ModelID: "mp-a", DevLabel: "SW-A", PortName: "eth0"},
		{DeviceID: "dev-b", ModelID: "mp-b", DevLabel: "SW-B", PortName: "eth0"},
	}
	ifaces = []e.DeviceInterface{
		{
			ID:       "iface-a",
			DeviceID: "dev-a",
			Name:     "access-vlan20",
			VlanConfigs: []e.PortVlanConfig{
				{VlanNumber: "10", Tagged: true},
				{VlanNumber: "20", Tagged: false}, // untagged → colors the port/connection
			},
		},
		{
			ID:       "iface-b",
			DeviceID: "dev-b",
			Name:     "trunk",
			VlanConfigs: []e.PortVlanConfig{
				{VlanNumber: "10", Tagged: true},
				{VlanNumber: "20", Tagged: true}, // all tagged → no color
			},
		},
	}
	ifacePorts = []e.InterfacePort{
		{InterfaceID: "iface-a", DeviceID: "dev-a", ModelPortID: "mp-a"},
		{InterfaceID: "iface-b", DeviceID: "dev-b", ModelPortID: "mp-b"},
	}
	return
}

// ── getVlanColor ──────────────────────────────────────────────────────────────

func TestGetVlanColor_AssignsColor(t *testing.T) {
	m := make(map[string]string)
	color := getVlanColor("10", m)
	if color == "" {
		t.Fatal("expected a non-empty color")
	}
	if m["10"] != color {
		t.Errorf("vlanColorMap not updated: got %q, want %q", m["10"], color)
	}
}

func TestGetVlanColor_SameVlanSameColor(t *testing.T) {
	m := make(map[string]string)
	c1 := getVlanColor("10", m)
	c2 := getVlanColor("10", m)
	if c1 != c2 {
		t.Errorf("same VLAN returned different colors: %q vs %q", c1, c2)
	}
}

func TestGetVlanColor_DifferentVlansDifferentColors(t *testing.T) {
	m := make(map[string]string)
	c10 := getVlanColor("10", m)
	c20 := getVlanColor("20", m)
	if c10 == c20 {
		t.Errorf("different VLANs got same color: %q", c10)
	}
}

func TestGetVlanColor_CyclesAfterSeven(t *testing.T) {
	m := make(map[string]string)
	// Fill all 7 palette slots
	for i := range 7 {
		getVlanColor(string(rune('A'+i)), m)
	}
	// 8th VLAN should reuse the first color
	c1 := getVlanColor("A", m) // already assigned
	c8 := getVlanColor("H", m) // new, cycles back
	if c1 != c8 {
		t.Errorf("color did not cycle: VLAN A=%q, VLAN H=%q", c1, c8)
	}
}

// ── getPortUntaggedVlan ───────────────────────────────────────────────────────

func TestGetPortUntaggedVlan_ReturnsUntaggedVlan(t *testing.T) {
	_, _, _, dps, ifaces, ifacePorts := twoSwitchFixture()

	vlan := getPortUntaggedVlan("SW-A", "eth0", dps, ifaces, ifacePorts)
	if vlan != "20" {
		t.Errorf("expected untagged VLAN 20, got %q", vlan)
	}
}

func TestGetPortUntaggedVlan_NoUntaggedReturnsEmpty(t *testing.T) {
	_, _, _, dps, ifaces, ifacePorts := twoSwitchFixture()

	// SW-B only has tagged VLANs
	vlan := getPortUntaggedVlan("SW-B", "eth0", dps, ifaces, ifacePorts)
	if vlan != "" {
		t.Errorf("expected empty string for all-tagged port, got %q", vlan)
	}
}

func TestGetPortUntaggedVlan_UnknownDeviceReturnsEmpty(t *testing.T) {
	_, _, _, dps, ifaces, ifacePorts := twoSwitchFixture()

	vlan := getPortUntaggedVlan("SW-UNKNOWN", "eth0", dps, ifaces, ifacePorts)
	if vlan != "" {
		t.Errorf("expected empty string for unknown device, got %q", vlan)
	}
}

func TestGetPortUntaggedVlan_NoInterfaceLinkedReturnsEmpty(t *testing.T) {
	_, _, _, dps, ifaces, _ := twoSwitchFixture()

	// Pass empty ifacePorts — nothing is linked
	vlan := getPortUntaggedVlan("SW-A", "eth0", dps, ifaces, []e.InterfacePort{})
	if vlan != "" {
		t.Errorf("expected empty string when no interface linked, got %q", vlan)
	}
}

// ── getConnectionVlans ────────────────────────────────────────────────────────

func TestGetConnectionVlans_ReturnsUnionFromBothPorts(t *testing.T) {
	_, _, conns, dps, ifaces, ifacePorts := twoSwitchFixture()

	vlans := getConnectionVlans(conns[0], dps, ifaces, ifacePorts)

	want := map[string]bool{"10": true, "20": true}
	if len(vlans) != len(want) {
		t.Fatalf("expected %d VLANs, got %d: %v", len(want), len(vlans), vlans)
	}
	for _, v := range vlans {
		if !want[v] {
			t.Errorf("unexpected VLAN %q in result", v)
		}
	}
}

func TestGetConnectionVlans_EmptyWhenNoInterfaces(t *testing.T) {
	_, _, conns, dps, ifaces, _ := twoSwitchFixture()

	vlans := getConnectionVlans(conns[0], dps, ifaces, []e.InterfacePort{})
	if len(vlans) != 0 {
		t.Errorf("expected no VLANs, got %v", vlans)
	}
}

// ── generateD2ConnectionStringsWithVlans ─────────────────────────────────────

func TestGenerateD2ConnectionStrings_ColoredWhenSourceHasUntaggedVlan(t *testing.T) {
	zones, devices, conns, dps, ifaces, ifacePorts := twoSwitchFixture()

	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(conns, deviceMap, zoneFullName)
	assignPortNumbers(deviceMap)

	colorMap := make(map[string]string)
	out := generateD2ConnectionStringsWithVlans(conns, deviceMap, zoneFullName, colorMap, dps, ifaces, ifacePorts, "untagged")

	if !strings.Contains(out, "style.stroke:") {
		t.Errorf("expected style.stroke in output, got:\n%s", out)
	}
	if len(colorMap) == 0 {
		t.Error("vlanColorMap should be populated after generating colored connections")
	}
}

func TestGenerateD2ConnectionStrings_NoColorWhenAllTagged(t *testing.T) {
	zones, devices, _, dps, ifaces, ifacePorts := twoSwitchFixture()

	// Build a connection where the SOURCE port (SW-B) has only tagged VLANs
	conns := []e.Connection{
		{
			FromDevice:    "SW-B",
			FromModelPort: "eth0",
			FromZoneID:    "zone-dc",
			ToDevice:      "SW-A",
			ToModelPort:   "eth0",
			ToZoneID:      "zone-dc",
		},
	}

	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(conns, deviceMap, zoneFullName)
	assignPortNumbers(deviceMap)

	colorMap := make(map[string]string)
	out := generateD2ConnectionStringsWithVlans(conns, deviceMap, zoneFullName, colorMap, dps, ifaces, ifacePorts, "untagged")

	if strings.Contains(out, "style.stroke:") {
		t.Errorf("expected no style.stroke for all-tagged source port, got:\n%s", out)
	}
	if len(colorMap) != 0 {
		t.Errorf("vlanColorMap should be empty, got %v", colorMap)
	}
}

// ── generateVlanLegend ────────────────────────────────────────────────────────

func TestGenerateVlanLegend_EmptyMapReturnsEmpty(t *testing.T) {
	out := generateVlanLegend(map[string]string{})
	if out != "" {
		t.Errorf("expected empty string for empty map, got %q", out)
	}
}

func TestGenerateVlanLegend_ContainsVarsBlock(t *testing.T) {
	out := generateVlanLegend(map[string]string{"10": "red"})

	for _, want := range []string{"vars:", "d2-legend:", "VLAN 10", "style.stroke: red"} {
		if !strings.Contains(out, want) {
			t.Errorf("legend missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateVlanLegend_MultipleVlansSorted(t *testing.T) {
	out := generateVlanLegend(map[string]string{
		"30": "green",
		"10": "red",
		"20": "blue",
	})

	pos10 := strings.Index(out, "VLAN 10")
	pos20 := strings.Index(out, "VLAN 20")
	pos30 := strings.Index(out, "VLAN 30")

	if pos10 < 0 || pos20 < 0 || pos30 < 0 {
		t.Fatalf("not all VLANs present in legend:\n%s", out)
	}
	if !(pos10 < pos20 && pos20 < pos30) {
		t.Errorf("VLANs not sorted in legend (10@%d, 20@%d, 30@%d)", pos10, pos20, pos30)
	}
}

// ── GenerateD2FocusConnectionsWithVlans (integration) ────────────────────────

func TestGenerateD2FocusConnectionsWithVlans_StrokeAndLegendPresent(t *testing.T) {
	zones, devices, conns, dps, ifaces, ifacePorts := twoSwitchFixture()

	out := GenerateD2FocusConnectionsWithVlans(devices, conns, zones, dps, ifaces, ifacePorts, false, "untagged", "connections")

	if !strings.Contains(out, "style.stroke:") {
		t.Errorf("expected colored connection in output:\n%s", out)
	}
	if !strings.Contains(out, "vars:") || !strings.Contains(out, "d2-legend:") {
		t.Errorf("expected legend in output:\n%s", out)
	}
	if !strings.Contains(out, "VLAN 20") {
		t.Errorf("expected 'VLAN 20' in legend:\n%s", out)
	}
}

func TestGenerateD2FocusConnectionsWithVlans_NoColorWhenNoUntagged(t *testing.T) {
	zones, devices, _, dps, ifaces, ifacePorts := twoSwitchFixture()

	// All-tagged connection: SW-B → SW-A
	conns := []e.Connection{
		{
			FromDevice:    "SW-B",
			FromModelPort: "eth0",
			FromZoneID:    "zone-dc",
			ToDevice:      "SW-A",
			ToModelPort:   "eth0",
			ToZoneID:      "zone-dc",
		},
	}

	out := GenerateD2FocusConnectionsWithVlans(devices, conns, zones, dps, ifaces, ifacePorts, false, "untagged", "connections")

	if strings.Contains(out, "style.stroke:") {
		t.Errorf("expected no colored connections for all-tagged ports:\n%s", out)
	}
	if strings.Contains(out, "vars:") {
		t.Errorf("expected no legend when no VLANs are colored:\n%s", out)
	}
}

// ── GenerateD2FocusPortsWithVlans (integration) ───────────────────────────────

func TestGenerateD2FocusPortsWithVlans_PortColorPresent(t *testing.T) {
	zones, devices, conns, dps, ifaces, ifacePorts := twoSwitchFixture()

	out := GenerateD2FocusPortsWithVlans(devices, conns, zones, dps, ifaces, ifacePorts, false, "untagged", "both")

	// SW-A eth0 has untagged VLAN 20 → port should have style.stroke
	if !strings.Contains(out, "style.stroke:") {
		t.Errorf("expected port coloring in output:\n%s", out)
	}
	if !strings.Contains(out, "d2-legend:") {
		t.Errorf("expected legend in output:\n%s", out)
	}
}

func TestGenerateD2FocusPortsWithVlans_NoPortColorWhenDisabled(t *testing.T) {
	zones, devices, conns, dps, ifaces, ifacePorts := twoSwitchFixture()

	outColored := GenerateD2FocusPortsWithVlans(devices, conns, zones, dps, ifaces, ifacePorts, false, "untagged", "both")
	outPlain := GenerateD2FocusPortsWithVlans(devices, conns, zones, dps, ifaces, ifacePorts, false, "untagged", "connections")

	// colorTarget="both": device port nodes inside device blocks get their own style.stroke
	// e.g.  1-1: "eth0" {\n    style.stroke: red\n  }
	if !strings.Contains(outColored, "\"eth0\" {") {
		t.Errorf("colorTarget=both: expected port block with style in device block:\n%s", outColored)
	}

	// colorTarget="connections": port nodes are plain labels, no nested style block
	if strings.Contains(outPlain, "\"eth0\" {") {
		t.Errorf("colorTarget=connections: port nodes should not have nested style blocks:\n%s", outPlain)
	}

	// Connection coloring and legend are always generated for colorTarget="connections"
	if !strings.Contains(outPlain, "style.stroke:") {
		t.Errorf("colorTarget=connections: connections should still be colored by untagged VLAN:\n%s", outPlain)
	}
	if !strings.Contains(outPlain, "d2-legend:") {
		t.Errorf("colorTarget=connections: legend should still be present from connection coloring:\n%s", outPlain)
	}
}

// ── New tests for vlanScope and colorTarget ───────────────────────────────────

func TestGenerateD2ConnectionStrings_AllScopeMultipleLinesPerLink(t *testing.T) {
	zones, devices, conns, dps, ifaces, ifacePorts := twoSwitchFixture()

	// The fixture has intersection {10, 20} → should emit 2 colored lines for the one link.
	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(conns, deviceMap, zoneFullName)
	assignPortNumbers(deviceMap)

	colorMap := make(map[string]string)
	out := generateD2ConnectionStringsWithVlans(conns, deviceMap, zoneFullName, colorMap, dps, ifaces, ifacePorts, "all")

	// Count occurrences of "style.stroke:" — should be 2 (one per VLAN in intersection)
	count := strings.Count(out, "style.stroke:")
	if count != 2 {
		t.Errorf("vlanScope=all: expected 2 style.stroke lines for intersection {10,20}, got %d:\n%s", count, out)
	}
	if len(colorMap) != 2 {
		t.Errorf("vlanScope=all: expected 2 VLANs in colorMap, got %d: %v", len(colorMap), colorMap)
	}
}

func TestGenerateD2ConnectionStrings_AllScopeIntersectionOnly(t *testing.T) {
	zones, devices, conns, dps, _, ifacePorts := twoSwitchFixture()

	// Override interfaces: SW-A has VLAN 30 (unique), SW-B has VLAN 40 (unique), neither shares a VLAN.
	ifaces := []e.DeviceInterface{
		{
			ID:       "iface-a",
			DeviceID: "dev-a",
			Name:     "a-only",
			VlanConfigs: []e.PortVlanConfig{
				{VlanNumber: "30", Tagged: true},
			},
		},
		{
			ID:       "iface-b",
			DeviceID: "dev-b",
			Name:     "b-only",
			VlanConfigs: []e.PortVlanConfig{
				{VlanNumber: "40", Tagged: true},
			},
		},
	}

	zoneFullName := buildZoneFullNameMap(zones)
	deviceMap := buildDeviceMap(devices, zoneFullName)
	collectPortsFromConnections(conns, deviceMap, zoneFullName)
	assignPortNumbers(deviceMap)

	colorMap := make(map[string]string)
	out := generateD2ConnectionStringsWithVlans(conns, deviceMap, zoneFullName, colorMap, dps, ifaces, ifacePorts, "all")

	// Intersection is empty → plain uncolored line
	if strings.Contains(out, "style.stroke:") {
		t.Errorf("vlanScope=all with empty intersection: expected no style.stroke, got:\n%s", out)
	}
	if len(colorMap) != 0 {
		t.Errorf("vlanScope=all with empty intersection: expected empty colorMap, got %v", colorMap)
	}
}

func TestGenerateD2FocusConnectionsWithVlans_ColorTargetPortsOnly(t *testing.T) {
	zones, devices, conns, dps, ifaces, ifacePorts := twoSwitchFixture()

	out := GenerateD2FocusConnectionsWithVlans(devices, conns, zones, dps, ifaces, ifacePorts, false, "untagged", "ports")

	// Connections should be plain (no style.stroke on connection lines)
	// We check that no "-- " connection line is followed by a style block.
	// The port blocks should have style.stroke (SW-A eth0 is untagged VLAN 20).
	if !strings.Contains(out, "style.stroke:") {
		t.Errorf("colorTarget=ports: expected port coloring (style.stroke) in output:\n%s", out)
	}
	// Port block with nested style: "eth0" {
	if !strings.Contains(out, "\"eth0\" {") {
		t.Errorf("colorTarget=ports: expected colored port node block in output:\n%s", out)
	}
	// Connection lines must NOT carry their own style block — verify no "-- " is followed by " {"
	// by checking the connection section doesn't contain "-- SW" followed by " {\n"
	if !strings.Contains(out, "d2-legend:") {
		t.Errorf("colorTarget=ports: expected legend from port coloring:\n%s", out)
	}
}

func TestGenerateD2FocusConnectionsWithVlans_ColorTargetConnectionsOnly(t *testing.T) {
	zones, devices, conns, dps, ifaces, ifacePorts := twoSwitchFixture()

	out := GenerateD2FocusConnectionsWithVlans(devices, conns, zones, dps, ifaces, ifacePorts, false, "untagged", "connections")

	// Connections should be colored
	if !strings.Contains(out, "style.stroke:") {
		t.Errorf("colorTarget=connections: expected connection coloring in output:\n%s", out)
	}
	// Port nodes must NOT have nested style blocks
	if strings.Contains(out, "\"eth0\" {") {
		t.Errorf("colorTarget=connections: port nodes should not have nested style blocks:\n%s", out)
	}
}

func TestGenerateD2FocusConnectionsWithVlans_VlanScopeAll(t *testing.T) {
	zones, devices, conns, dps, ifaces, ifacePorts := twoSwitchFixture()

	// The fixture intersection is {10, 20} → 2 colored connection lines for the one physical link.
	out := GenerateD2FocusConnectionsWithVlans(devices, conns, zones, dps, ifaces, ifacePorts, false, "all", "connections")

	count := strings.Count(out, "style.stroke:")
	if count < 2 {
		t.Errorf("vlanScope=all: expected at least 2 style.stroke lines, got %d:\n%s", count, out)
	}
	if !strings.Contains(out, "VLAN 10") || !strings.Contains(out, "VLAN 20") {
		t.Errorf("vlanScope=all: expected legend entries for VLAN 10 and VLAN 20:\n%s", out)
	}
}
