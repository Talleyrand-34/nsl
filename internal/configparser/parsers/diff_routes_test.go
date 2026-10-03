// SPDX-License-Identifier: AGPL-3.0-or-later
package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nsl-graph/internal/configparser"
)

func TestDiffRoutes_Add(t *testing.T) {
	intended := &configparser.ConfigData{
		Routes: []configparser.ConfigRoute{
			{Network: "192.168.5.0/24", Gateway: "10.0.0.1", Interface: "wan"},
		},
	}
	diffs := diffRoutes(intended, nil)
	assert.Len(t, diffs, 1)
	assert.Equal(t, "route-add", diffs[0].Kind)
	assert.Equal(t, "192.168.5.0/24", diffs[0].Path)
	assert.Equal(t, "192.168.5.0/24", diffs[0].New)
}

func TestDiffRoutes_Update(t *testing.T) {
	// intended has the route, observed has it too — no diff.
	intended := &configparser.ConfigData{
		Routes: []configparser.ConfigRoute{
			{Network: "192.168.5.0/24", Gateway: "10.0.0.1", Interface: "wan"},
		},
	}
	observed := &configparser.ConfigData{
		Routes: []configparser.ConfigRoute{
			{Network: "192.168.5.0/24", Gateway: "10.0.0.1", Interface: "wan"},
		},
	}
	diffs := diffRoutes(intended, observed)
	assert.Len(t, diffs, 0)
}

func TestDiffRoutes_MixedAddAndDelete(t *testing.T) {
	intended := &configparser.ConfigData{
		Routes: []configparser.ConfigRoute{
			{Network: "10.1.0.0/16", Gateway: "10.0.0.1", Interface: "wan"},
			{Network: "10.2.0.0/16", Gateway: "10.0.0.1", Interface: "wan"},
		},
	}
	observed := &configparser.ConfigData{
		Routes: []configparser.ConfigRoute{
			{Network: "10.1.0.0/16", Gateway: "10.0.0.1", Interface: "wan"},
			{Network: "10.3.0.0/16", Gateway: "10.0.0.1", Interface: "wan"},
		},
	}
	diffs := diffRoutes(intended, observed)
	assert.Len(t, diffs, 2)

	var kinds []string
	for _, d := range diffs {
		kinds = append(kinds, d.Kind)
	}
	assert.Contains(t, kinds, "route-add")
	assert.Contains(t, kinds, "route-del")
}

func TestDiffRoutes_NilIntended(t *testing.T) {
	diffs := diffRoutes(nil, &configparser.ConfigData{})
	assert.Len(t, diffs, 0)
}
