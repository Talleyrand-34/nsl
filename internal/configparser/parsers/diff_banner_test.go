// SPDX-License-Identifier: MIT
package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nsl-graph/internal/configparser"
)

func TestDiffBanner_NoIntended(t *testing.T) {
	diffs := diffBanner(nil, &configparser.ConfigData{Banner: &configparser.ConfigBanner{LoginBanner: "hi"}})
	assert.Len(t, diffs, 0)
}

func TestDiffBanner_NilObserved(t *testing.T) {
	intended := &configparser.ConfigData{
		Banner: &configparser.ConfigBanner{LoginBanner: "Welcome", PostLogin: "Goodbye"},
	}
	diffs := diffBanner(intended, nil)
	assert.Len(t, diffs, 1)
	assert.Equal(t, "banner-set", diffs[0].Kind)
}

func TestDiffBanner_Identical(t *testing.T) {
	b := &configparser.ConfigBanner{LoginBanner: "Welcome", PostLogin: "Goodbye"}
	diffs := diffBanner(&configparser.ConfigData{Banner: b}, &configparser.ConfigData{Banner: b})
	assert.Len(t, diffs, 0)
}

func TestDiffBanner_LoginBannerChange(t *testing.T) {
	diffs := diffBanner(
		&configparser.ConfigData{Banner: &configparser.ConfigBanner{LoginBanner: "New"}},
		&configparser.ConfigData{Banner: &configparser.ConfigBanner{LoginBanner: "Old"}},
	)
	assert.Len(t, diffs, 1)
	assert.Equal(t, "banner-set", diffs[0].Kind)
}
