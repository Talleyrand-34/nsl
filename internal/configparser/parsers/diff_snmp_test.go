// SPDX-License-Identifier: AGPL-3.0-or-later
package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nsl-graph/internal/configparser"
)

func TestDiffSNMP_NoIntended(t *testing.T) {
	diffs := diffSNMP(nil, &configparser.ConfigData{SNMP: &configparser.ConfigSNMPConfig{Enabled: true}})
	assert.Len(t, diffs, 0)
}

func TestDiffSNMP_NilObserved(t *testing.T) {
	intended := &configparser.ConfigData{
		SNMP: &configparser.ConfigSNMPConfig{
			Enabled:     true,
			Location:    "Rack 4",
			Communities: []configparser.ConfigSNMPCommunity{{Name: "public", Access: "ro"}},
		},
	}
	diffs := diffSNMP(intended, nil)
	assert.Len(t, diffs, 1)
	assert.Equal(t, "snmp-set", diffs[0].Kind)
}

func TestDiffSNMP_Identical(t *testing.T) {
	cfg := &configparser.ConfigSNMPConfig{
		Enabled:     true,
		Location:    "Rack 4",
		Communities: []configparser.ConfigSNMPCommunity{{Name: "public", Access: "ro"}},
	}
	diffs := diffSNMP(&configparser.ConfigData{SNMP: cfg}, &configparser.ConfigData{SNMP: cfg})
	assert.Len(t, diffs, 0)
}

func TestDiffSNMP_CommunityChange(t *testing.T) {
	diffs := diffSNMP(
		&configparser.ConfigData{SNMP: &configparser.ConfigSNMPConfig{Enabled: true, Communities: []configparser.ConfigSNMPCommunity{{Name: "private", Access: "rw"}}}},
		&configparser.ConfigData{SNMP: &configparser.ConfigSNMPConfig{Enabled: true, Communities: []configparser.ConfigSNMPCommunity{{Name: "public", Access: "ro"}}}},
	)
	assert.Len(t, diffs, 1)
	assert.Equal(t, "snmp-set", diffs[0].Kind)
}
