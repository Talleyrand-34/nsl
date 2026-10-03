// SPDX-License-Identifier: AGPL-3.0-or-later
// diff_lldp.go: shared diff logic for LLDP configuration.
package parsers

import "nsl-graph/internal/configparser"

// lldpEqual compares two LLDP configs for equality.
func lldpEqual(a, b *configparser.ConfigLLDPSettings) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Enabled == b.Enabled &&
		a.ChassisID == b.ChassisID &&
		a.SystemName == b.SystemName &&
		a.SystemDesc == b.SystemDesc &&
		a.MgmtAddress == b.MgmtAddress &&
		a.InterfacePattern == b.InterfacePattern
}

// diffLLDP returns a single lldp-set change if intended differs from observed.
func diffLLDP(intended, observed *configparser.ConfigData) []configparser.ConfigChange {
	if intended == nil || intended.LLDP == nil {
		return nil
	}
	if observed == nil || observed.LLDP == nil || !lldpEqual(intended.LLDP, observed.LLDP) {
		return []configparser.ConfigChange{{
			Kind:  "lldp-set",
			Path:  "system.lldp",
			Patch: []string{"lldp"},
		}}
	}
	return nil
}
