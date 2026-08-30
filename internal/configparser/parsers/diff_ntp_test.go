// SPDX-License-Identifier: MIT
package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nsl-graph/internal/configparser"
)

func TestDiffNTP_NoIntended(t *testing.T) {
	diffs := diffNTP(nil, &configparser.ConfigData{NTP: &configparser.ConfigNTPConfig{}})
	assert.Len(t, diffs, 0)
}

func TestDiffNTP_NilObserved(t *testing.T) {
	intended := &configparser.ConfigData{
		NTP: &configparser.ConfigNTPConfig{
			Enabled:  true,
			Timezone: "Europe/Madrid",
			Servers: []configparser.ConfigNTPServer{
				{Address: "0.pool.ntp.org", Enabled: true},
			},
		},
	}
	diffs := diffNTP(intended, nil)
	// nil observed: emit ntp-add for every intended server plus a
	// ntp-set for the scalar fields. Servers and scalar are
	// independent now.
	assert.Len(t, diffs, 2)
	assert.Equal(t, "ntp-add", diffs[0].Kind)
	assert.Equal(t, "ntp-set", diffs[1].Kind)
}

func TestDiffNTP_Identical(t *testing.T) {
	cfg := &configparser.ConfigNTPConfig{
		Enabled:  true,
		Timezone: "Europe/Madrid",
		Servers: []configparser.ConfigNTPServer{
			{Address: "0.pool.ntp.org", Enabled: true},
		},
	}
	diffs := diffNTP(&configparser.ConfigData{NTP: cfg}, &configparser.ConfigData{NTP: cfg})
	assert.Len(t, diffs, 0)
}

func TestDiffNTP_ServerChange(t *testing.T) {
	intended := &configparser.ConfigData{
		NTP: &configparser.ConfigNTPConfig{
			Enabled:  true,
			Timezone: "Europe/Madrid",
			Servers: []configparser.ConfigNTPServer{
				{Address: "1.pool.ntp.org"},
			},
		},
	}
	observed := &configparser.ConfigData{
		NTP: &configparser.ConfigNTPConfig{
			Enabled:  true,
			Timezone: "Europe/Madrid",
			Servers: []configparser.ConfigNTPServer{
				{Address: "0.pool.ntp.org"},
			},
		},
	}
	diffs := diffNTP(intended, observed)
	// ntp-add 1.pool.ntp.org, ntp-del 0.pool.ntp.org; no ntp-set
	// because the scalar fields match.
	// Server change (1.pool.ntp.org replaces 0.pool.ntp.org) AND
	// scalar ntp-set (because server addresses differ between the
	// two ConfigData, ntpEqual returns false).
	assert.Len(t, diffs, 3)
	assert.Equal(t, "ntp-add", diffs[0].Kind)
	assert.Equal(t, "1.pool.ntp.org", diffs[0].New)
	assert.Equal(t, "ntp-del", diffs[1].Kind)
	assert.Equal(t, "0.pool.ntp.org", diffs[1].New)
}
