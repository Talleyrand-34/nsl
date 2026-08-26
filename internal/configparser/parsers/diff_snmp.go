// SPDX-License-Identifier: MIT
// diff_snmp.go: shared diff logic for SNMP configuration.
package parsers

import (
	"sort"
	"nsl-graph/internal/configparser"
)

func snmpEqual(a, b *configparser.ConfigSNMPConfig) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Enabled != b.Enabled || a.Location != b.Location || a.Contact != b.Contact {
		return false
	}
	if len(a.Communities) != len(b.Communities) {
		return false
	}
	// Sort copies so we don't mutate the originals.
	ac := make([]configparser.ConfigSNMPCommunity, len(a.Communities))
	bc := make([]configparser.ConfigSNMPCommunity, len(b.Communities))
	copy(ac, a.Communities)
	copy(bc, b.Communities)
	sort.Slice(ac, func(i, j int) bool { return ac[i].Name < ac[j].Name })
	sort.Slice(bc, func(i, j int) bool { return bc[i].Name < bc[j].Name })
	for i := range ac {
		if ac[i].Name != bc[i].Name || ac[i].Access != bc[i].Access ||
			ac[i].TrapTarget != bc[i].TrapTarget {
			return false
		}
	}
	return true
}

// diffSNMP returns a snmp-set change if intended differs from observed.
func diffSNMP(intended, observed *configparser.ConfigData) []configparser.ConfigChange {
	if intended == nil || intended.SNMP == nil {
		return nil
	}
	if observed == nil || observed.SNMP == nil || !snmpEqual(intended.SNMP, observed.SNMP) {
		return []configparser.ConfigChange{{
			Kind:  "snmp-set",
			Path:  "system.snmp",
			Patch: []string{"snmp"},
		}}
	}
	return nil
}
