// SPDX-License-Identifier: MIT
// diff_banner.go: shared diff logic for banner configuration.
package parsers

import "nsl-graph/internal/configparser"

// diffBanner returns a banner-set change if intended differs from observed.
// nil intended → no diff; nil observed → always set.
func diffBanner(intended, observed *configparser.ConfigData) []configparser.ConfigChange {
	if intended == nil || intended.Banner == nil {
		return nil
	}
	if observed == nil || observed.Banner == nil ||
		intended.Banner.LoginBanner != observed.Banner.LoginBanner ||
		intended.Banner.PostLogin != observed.Banner.PostLogin {
		return []configparser.ConfigChange{{
			Kind:  "banner-set",
			Path:  "system.banner",
			Patch: []string{"banner"},
		}}
	}
	return nil
}
