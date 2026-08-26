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
	assert.Len(t, diffs, 1)
	assert.Equal(t, "ntp-set", diffs[0].Kind)
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
	assert.Len(t, diffs, 1)
	assert.Equal(t, "ntp-set", diffs[0].Kind)
}
