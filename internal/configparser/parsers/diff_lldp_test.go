// SPDX-License-Identifier: AGPL-3.0-or-later
package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nsl-graph/internal/configparser"
)

func TestDiffLLDP_NoIntended(t *testing.T) {
	diffs := diffLLDP(nil, &configparser.ConfigData{LLDP: &configparser.ConfigLLDPSettings{Enabled: true}})
	assert.Len(t, diffs, 0)
}

func TestDiffLLDP_NilObserved(t *testing.T) {
	intended := &configparser.ConfigData{
		LLDP: &configparser.ConfigLLDPSettings{Enabled: true, SystemName: "router1"},
	}
	diffs := diffLLDP(intended, nil)
	assert.Len(t, diffs, 1)
	assert.Equal(t, "lldp-set", diffs[0].Kind)
}

func TestDiffLLDP_Identical(t *testing.T) {
	cfg := &configparser.ConfigLLDPSettings{Enabled: true, SystemName: "router1"}
	diffs := diffLLDP(&configparser.ConfigData{LLDP: cfg}, &configparser.ConfigData{LLDP: cfg})
	assert.Len(t, diffs, 0)
}

func TestDiffLLDP_SystemNameChange(t *testing.T) {
	diffs := diffLLDP(
		&configparser.ConfigData{LLDP: &configparser.ConfigLLDPSettings{Enabled: true, SystemName: "new-name"}},
		&configparser.ConfigData{LLDP: &configparser.ConfigLLDPSettings{Enabled: true, SystemName: "old-name"}},
	)
	assert.Len(t, diffs, 1)
	assert.Equal(t, "lldp-set", diffs[0].Kind)
}
