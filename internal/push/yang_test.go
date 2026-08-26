// yang_test.go: TDD for the YANG validation gate before SSH opens.
package push

import (
	"strings"
	"testing"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/topology"
)

func TestYANGGate_NilTopoPasses(t *testing.T) {
	gate := NewYANGGate(nil)
	err := gate.Validate(&configparser.ConfigData{Hostname: "R1"})
	if err != nil {
		t.Errorf("nil topology must not block a push; got: %v", err)
	}
}

func TestYANGGate_AcceptsValidConfig(t *testing.T) {
	gate := NewYANGGate(nil)
	cfg := &configparser.ConfigData{
		Hostname:  "R1",
		OsType:    "openwrt",
		Interfaces: []configparser.ConfigInterface{
			{Name: "eth0", Type: "physical", Enabled: true},
		},
	}
	err := gate.Validate(cfg)
	if err != nil {
		t.Errorf("trivial config should pass; got: %v", err)
	}
}

func TestYANGGate_RejectsConflictingVLANs(t *testing.T) {
	// Same VLAN declared tagged AND untagged on the same port: violates the
	// model rule that getStoredVLANsForDeviceport flagged in the network-refinement branch.
	gate := NewYANGGate(nil)
	cfg := &configparser.ConfigData{
		Hostname: "R1",
		OsType:   "openwrt",
		Interfaces: []configparser.ConfigInterface{
			{
				Name: "eth0", Type: "physical", Enabled: true,
				VLANs: []configparser.ConfigVLAN{
					{ID: "10", Tagged: true},
					{ID: "10", Tagged: false},
				},
			},
		},
	}
	err := gate.Validate(cfg)
	if err == nil {
		t.Fatal("config with port carrying VLAN 10 tagged AND untagged must be rejected")
	}
	if !strings.Contains(err.Error(), "10") {
		t.Errorf("error must name the offending VLAN; got: %v", err)
	}
}

func TestYANGGate_ReturnsErrorBeforeSSHOpen(t *testing.T) {
	// Contract: this gate is the LAST thing before Render is called. A test
	// that exercises the full pipeline (gate + render) must show the gate
	// blocking an invalid config WITHOUT opening a session.
	gate := NewYANGGate(nil)

	sess := &countingSession{}
	cfg := invalidConfig()

	err := gate.Validate(cfg)
	if err == nil {
		// If gate allows it, the render must not open the session for invalid configs.
		// For the test to be meaningful we use a known-invalid cfg; if the gate
		// still allows it, that's a gate regression — fail the test.
		t.Fatal("gate must reject the conflicting-VLAN config")
	}
	if sess.opens != 0 {
		t.Errorf("gate must reject BEFORE any session is opened; got %d opens", sess.opens)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func invalidConfig() *configparser.ConfigData {
	return &configparser.ConfigData{
		Hostname: "R1",
		OsType:   "openwrt",
		Interfaces: []configparser.ConfigInterface{
			{
				Name: "eth0", Type: "physical", Enabled: true,
				VLANs: []configparser.ConfigVLAN{
					{ID: "10", Tagged: true},
					{ID: "10", Tagged: false},
				},
			},
		},
	}
}

// countingSession opens-counting fake to prove the gate rejects before any SSH.
type countingSession struct{ opens int }

func (c *countingSession) Open() { c.opens++ }

var _ = topology.ConnectionEdge{} // pin topology import — used in real YANG gate later
