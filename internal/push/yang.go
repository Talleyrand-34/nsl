// yang.go: YANG validation gate before SSH opens.
//
// A push that violates the nsl-topology model invariants is rejected *before*
// any network I/O. The rule mirror: the same constraint that datastore.Validate
// enforces at the link-layer is enforced here at the port/VLAN layer.
package push

import (
	"fmt"

	"nsl-graph/internal/configparser"
)

// YANGGate runs the model invariants on a ConfigData before Render is called.
type YANGGate struct {
	// future: validator hook that maps ConfigData → canon.Root and runs pyang
}

func NewYANGGate(_ interface{ /* placeholder for a future validator */ }) *YANGGate {
	return &YANGGate{}
}

// Validate returns nil when the config is consistent, or a structured error
// naming the offending field.
func (g *YANGGate) Validate(cfg *configparser.ConfigData) error {
	if cfg == nil {
		return nil
	}
	for _, iface := range cfg.Interfaces {
		// For each (iface, vlan) pair, record the tag bit. A conflict is
		// two entries for the same VLAN where one is tagged and one is not.
		flags := make(map[string]bool)
		hasEntry := make(map[string]bool)
		for _, v := range iface.VLANs {
			if hasEntry[v.ID] && flags[v.ID] != v.Tagged {
				return fmt.Errorf("yang-gate: port %s carries VLAN %s as both tagged and untagged (rule: 1 VLAN per port per tagging)", iface.Name, v.ID)
			}
			flags[v.ID] = v.Tagged
			hasEntry[v.ID] = true
		}
	}
	return nil
}
