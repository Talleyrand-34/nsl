// SPDX-License-Identifier: AGPL-3.0-or-later
package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nsl-graph/internal/configparser"
)

func TestDiffSyslog_NilIntended(t *testing.T) {
	diffs := diffSyslog(nil, &configparser.ConfigData{
		Syslog: &configparser.ConfigSyslogConfig{Enabled: true},
	})
	assert.Len(t, diffs, 0)
}

func TestDiffSyslog_NilObserved(t *testing.T) {
	intended := &configparser.ConfigData{
		Syslog: &configparser.ConfigSyslogConfig{
			Enabled: true,
			Targets: []configparser.ConfigSyslogTarget{
				{Address: "192.0.2.1", Port: 514, Protocol: "udp"},
			},
		},
	}
	diffs := diffSyslog(intended, nil)
	assert.Len(t, diffs, 1)
	assert.Equal(t, "syslog-set", diffs[0].Kind)
}

func TestDiffSyslog_Identical(t *testing.T) {
	cfg := &configparser.ConfigSyslogConfig{
		Enabled: true,
		Targets: []configparser.ConfigSyslogTarget{
			{Address: "192.0.2.1", Port: 514, Protocol: "udp"},
		},
	}
	diffs := diffSyslog(&configparser.ConfigData{Syslog: cfg}, &configparser.ConfigData{Syslog: cfg})
	assert.Len(t, diffs, 0)
}

func TestDiffSyslog_TargetChanged(t *testing.T) {
	intended := &configparser.ConfigData{
		Syslog: &configparser.ConfigSyslogConfig{
			Enabled: true,
			Targets: []configparser.ConfigSyslogTarget{
				{Address: "192.0.2.2", Port: 514, Protocol: "udp"},
			},
		},
	}
	observed := &configparser.ConfigData{
		Syslog: &configparser.ConfigSyslogConfig{
			Enabled: true,
			Targets: []configparser.ConfigSyslogTarget{
				{Address: "192.0.2.1", Port: 514, Protocol: "udp"},
			},
		},
	}
	diffs := diffSyslog(intended, observed)
	assert.Len(t, diffs, 1)
	assert.Equal(t, "syslog-set", diffs[0].Kind)
}
