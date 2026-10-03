// SPDX-License-Identifier: AGPL-3.0-or-later
// diff_syslog.go: shared diff logic for syslog configuration.
package parsers

import (
	"nsl-graph/internal/configparser"
)

// diffSyslog returns a single syslog-set change if intended differs from observed.
// nil intended → no diff (don't delete); nil observed → always set.
func diffSyslog(intended, observed *configparser.ConfigData) []configparser.ConfigChange {
	if intended == nil || intended.Syslog == nil {
		return nil
	}
	if observed == nil || observed.Syslog == nil || !syslogEqual(intended.Syslog, observed.Syslog) {
		return []configparser.ConfigChange{{
			Kind:  "syslog-set",
			Path:  "system.syslog",
			Patch: []string{"syslog"},
		}}
	}
	return nil
}

// syslogEqual compares two Syslog configs for equality.
func syslogEqual(a, b *configparser.ConfigSyslogConfig) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Enabled != b.Enabled || a.PreserveFQDN != b.PreserveFQDN {
		return false
	}
	if len(a.Targets) != len(b.Targets) {
		return false
	}
	for i := range a.Targets {
		if !targetEqual(&a.Targets[i], &b.Targets[i]) {
			return false
		}
	}
	return true
}

// targetEqual compares two syslog targets by Address.
func targetEqual(a, b *configparser.ConfigSyslogTarget) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Address == b.Address
}
