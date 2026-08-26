// SPDX-License-Identifier: MIT
// diff_ntp.go: shared diff logic for NTP configuration.
package parsers

import (
	"nsl-graph/internal/configparser"
)

// diffNTP returns a single ntp-set change if intended differs from observed.
// nil intended → no diff (don't delete); nil observed → always set.
func diffNTP(intended, observed *configparser.ConfigData) []configparser.ConfigChange {
	if intended == nil || intended.NTP == nil {
		return nil
	}
	if observed == nil || observed.NTP == nil || !ntpEqual(intended.NTP, observed.NTP) {
		return []configparser.ConfigChange{{
			Kind:  "ntp-set",
			Path:  "system.ntp",
			Patch: []string{"ntp"},
		}}
	}
	return nil
}

// ntpEqual compares two NTP configs for equality.
func ntpEqual(a, b *configparser.ConfigNTPConfig) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Enabled != b.Enabled || a.LocalClock != b.LocalClock || a.Timezone != b.Timezone {
		return false
	}
	if len(a.Servers) != len(b.Servers) {
		return false
	}
	for i := range a.Servers {
		if a.Servers[i].Address != b.Servers[i].Address ||
			a.Servers[i].Prefer != b.Servers[i].Prefer ||
			a.Servers[i].IBurst != b.Servers[i].IBurst ||
			a.Servers[i].Enabled != b.Servers[i].Enabled {
			return false
		}
	}
	return true
}
