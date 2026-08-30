// SPDX-License-Identifier: MIT
// diff_ntp.go: additive NTP diff. Emits one change per server delta so
// the renderer can do `add_list` / `del_list` instead of `uci del`
// followed by a full replace — preserving the device's original NTP
// server list across an apply. Scalar fields (Enabled, Timezone,
// LocalClock) still emit a single `ntp-set` change, but the renderer
// for that case must not touch the server list.
package parsers

import (
	"nsl-graph/internal/configparser"
)

// diffNTP returns additive NTP changes: one per server to add or
// remove, plus a single `ntp-set` if any scalar field changed.
// nil intended → no diff (don't delete). nil observed → emit adds
// for all intended servers.
func diffNTP(intended, observed *configparser.ConfigData) []configparser.ConfigChange {
	if intended == nil || intended.NTP == nil {
		return nil
	}
	obsServers := map[string]bool{}
	var obsCfg *configparser.ConfigNTPConfig
	if observed != nil {
		obsCfg = observed.NTP
		if obsCfg != nil {
			for _, s := range obsCfg.Servers {
				if s.Address != "" {
					obsServers[s.Address] = true
				}
			}
		}
	}
	var out []configparser.ConfigChange
	intentServers := map[string]bool{}
	for _, s := range intended.NTP.Servers {
		if s.Address == "" {
			continue
		}
		intentServers[s.Address] = true
		if !obsServers[s.Address] {
			out = append(out, configparser.ConfigChange{
				Kind: "ntp-add",
				Path: "system.ntp.server=" + s.Address,
				New:  s.Address,
			})
		}
	}
	if obsCfg != nil {
		for _, s := range obsCfg.Servers {
			if s.Address == "" {
				continue
			}
			if !intentServers[s.Address] {
				out = append(out, configparser.ConfigChange{
					Kind: "ntp-del",
					Path: "system.ntp.server=" + s.Address,
					New:  s.Address,
				})
			}
		}
	}
	// Scalar field changes produce a single `ntp-set` so the renderer
	// still owns the enable/timezone/local_clock settings. The renderer's
	// ntp-set path must not touch the server list.
	if !ntpEqual(intended.NTP, obsCfg) {
		out = append(out, configparser.ConfigChange{
			Kind: "ntp-set",
			Path: "system.ntp",
		})
	}
	return out
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