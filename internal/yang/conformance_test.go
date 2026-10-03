// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

// Package yang_test asserts that what the mapper actually emits conforms to the
// standard YANG models -- not merely that the modules parse.
//
// This is the project's central claim ("the entity model IS RFC 8345's
// node/termination-point/link"), and here it is a `go test` assertion rather than a
// sentence in a document. It shells out to yanglint (libyang), the reference
// validator, because nothing in Go can check a `must` statement or a leafref for us.
//
// It skips when yanglint is absent so the suite still passes on a machine without it;
// `make yang-validate` is the gate that must not be skipped in CI.
package yang_test

/*
  Copyright © 2026 Talleyrand-34 (t34@t34.dev)

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU Affero General Public License as published
  by the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  GNU Affero General Public License for more details.

  You should have received a copy of the GNU Affero General Public License
  along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e "nsl-graph/internal/repository/entities"
	"nsl-graph/internal/yang/mapping"
)

const modulesDir = "modules"

// modules must be listed after the modules they augment, or libyang treats the
// augmented module as merely imported and silently SKIPS the `when` checks that gate
// our augments -- validating less than it appears to.
var modules = []string{
	"ietf-network.yang",
	"ietf-network-topology.yang",
	"ietf-l2-topology.yang",
	"nsl-topology.yang",
	"nsl-inventory.yang",
}

// validate runs yanglint over a JSON instance document and returns its combined
// output plus whether it was accepted.
func validate(t *testing.T, docPath string) (string, bool) {
	t.Helper()

	bin, err := exec.LookPath("yanglint")
	if err != nil {
		t.Skip("yanglint not on PATH (Debian: sudo apt install libyang3-tools); " +
			"`make yang-validate` is the gate that must not be skipped")
	}

	args := []string{"-t", "config", "-p", modulesDir}
	for _, m := range modules {
		args = append(args, filepath.Join(modulesDir, m))
	}
	args = append(args, docPath)

	out, err := exec.Command(bin, args...).CombinedOutput()
	return string(out), err == nil
}

// collapse folds all runs of whitespace into single spaces, so an assertion can match
// text that YANG description formatting has line-wrapped.
func collapse(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// writeDoc marshals v as RFC 7951 JSON into a temp file and returns its path.
func writeDoc(t *testing.T, v any) string {
	t.Helper()

	buf, err := json.MarshalIndent(v, "", "  ")
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "doc.json")
	require.NoError(t, os.WriteFile(path, buf, 0o600))
	return path
}

// spec is a representative specification: two devices in DIFFERENT zones, linked.
//
// The cross-zone link is the point. It is exactly what a zone-per-network mapping
// could not express -- RFC 8345 requires both endpoints of a link to be in the same
// network ("Must be in the same topology") -- which is why zones are an inventory
// catalogue and the topology is a single network.
func spec() mapping.Source {
	return mapping.Source{
		Brands:     []e.Brand{{ID: "b1", Name: "Siemens"}},
		ModelTypes: []e.ModelType{{ID: "mt1", Name: "Switch"}},
		OsTypes:    []e.OsType{{ID: "os1", Name: "openwrt"}},
		Owners:     []e.Owner{{ID: "o1", Name: "IT"}},
		ZoneTypes:  []e.ZoneType{{ID: "zt1", Name: "Security"}},
		Zones: []e.Zone{
			{ID: "z-hq", Name: "HQ", LocationType: "Security", Owner: "IT"},
			{ID: "z-dmz", Name: "DMZ", FatherID: "z-hq", LocationType: "Security", Owner: "IT"},
		},
		Models: []e.ModelDevice{
			{ID: "m1", Model: "XC206", Brand: "Siemens", ModelType: "Switch", OsType: "openwrt"},
		},
		ModelPorts: []e.ModelPort{
			{ID: "mp1", Name: "P1", Model: "XC206", Brand: "Siemens"},
			{ID: "mp2", Name: "radio0", Model: "XC206", Brand: "Siemens", PortType: "wifi", Band: "5GHz"},
		},
		Devices: []e.Device{
			{ID: "d1", Label: "sw1", Model: "XC206", Brand: "Siemens", ZoneID: "z-dmz", Owner: "IT", Ips: []string{"10.0.2.20"}},
			{ID: "d2", Label: "fw1", Model: "XC206", Brand: "Siemens", ZoneID: "z-hq", Owner: "IT", Ips: []string{"10.0.2.1"}},
		},
		DevicePorts: []e.DevicePort{
			{
				ID: "dp1", DeviceID: "d1", ModelID: "mp1", DevLabel: "sw1", PortName: "P1",
				MacAddress: "00:1B:1B:00:00:01",
				VlanConfigs: []e.PortVlanConfig{
					{VlanNumber: "10", Tagged: false},
					{VlanNumber: "20", Tagged: true},
				},
			},
			{
				ID: "dp2", DeviceID: "d2", ModelID: "mp1", DevLabel: "fw1", PortName: "P1",
				MacAddress:  "00:1B:1B:00:00:02",
				VlanConfigs: []e.PortVlanConfig{{VlanNumber: "20", Tagged: true}},
			},
		},
		Connections: []e.Connection{
			{
				ID: "c1", FromDevice: "sw1", FromModelPort: "P1", ToDevice: "fw1", ToModelPort: "P1",
				DiscoveredVia: []string{"ssh-lldp@sw1:P1", "snmp-lldp@fw1:P1"},
			},
		},
	}
}

// TestExportConformsToRFC8345 is the claim, as a test.
func TestExportConformsToRFC8345(t *testing.T) {
	root, warnings := mapping.FromEntities(spec())
	require.Empty(t, warnings, "the fixture must be clean; warnings mean the fixture is wrong, not the schema")

	out, ok := validate(t, writeDoc(t, root))
	assert.True(t, ok, "the exported document must validate against RFC 8345 / RFC 8944 / IEEE 802.1Q:\n%s", out)
}

// TestSchemaRejectsWeakUnreviewedLink is the negative, and it is the one that matters.
//
// nsl-topology carries a `must` -- a link whose only evidence is the forwarding
// database may not be committed unreviewed, because the MAC may be behind an
// intermediate device. If this test ever starts PASSING validation, the `must` has
// silently stopped working and the guarantee is gone.
func TestSchemaRejectsWeakUnreviewedLink(t *testing.T) {
	root, _ := mapping.FromEntities(spec())

	link := &root.Networks.Network[0].Links[0]
	link.Confidence = "weak"
	link.Reviewed = false

	out, ok := validate(t, writeDoc(t, root))

	assert.False(t, ok, "a weak, unreviewed link MUST be rejected by the schema; it was accepted")
	// yanglint reproduces the error-message with the description's own line breaks, so
	// compare on collapsed whitespace rather than the literal string.
	assert.Contains(t, collapse(out), "must be reviewed by an operator",
		"rejection should cite nsl-topology's own error-message")
}

// A weak link that HAS been reviewed is legitimate: an operator looked at the
// forwarding-database evidence and accepted it. The `must` must not block that.
func TestSchemaAcceptsWeakReviewedLink(t *testing.T) {
	root, _ := mapping.FromEntities(spec())

	link := &root.Networks.Network[0].Links[0]
	link.Confidence = "weak"
	link.Reviewed = true

	out, ok := validate(t, writeDoc(t, root))
	assert.True(t, ok, "a REVIEWED weak link is legitimate and must validate:\n%s", out)
}

// The typed identifiers are the other half of what the schema buys. These are values
// entities.* holds happily as strings today.
func TestSchemaRejectsOutOfRangeVlan(t *testing.T) {
	root, _ := mapping.FromEntities(spec())

	// Bypass the mapper, which would have dropped this with a warning: the point here
	// is that the SCHEMA refuses it too, so a hand-edited or third-party document
	// cannot smuggle it in.
	tp := &root.Networks.Network[0].Nodes[0].TerminationPoints[0]
	tp.VlanMembership[0].VlanID = 4999

	out, ok := validate(t, writeDoc(t, root))

	assert.False(t, ok, "VLAN 4999 is outside the 12-bit 802.1Q range and must be rejected")
	assert.Contains(t, out, "range", "rejection should cite the range constraint:\n%s", out)
}

// TestSchemaRejectsDanglingInventoryReference proves the referential integrity claim
// where it actually holds: the nsl-* leafrefs.
//
// entities.Device.Model is a model NAME held as a bare string with nothing checking
// it, which is why renaming a model silently orphans its devices. As a leafref it
// cannot dangle.
func TestSchemaRejectsDanglingInventoryReference(t *testing.T) {
	root, _ := mapping.FromEntities(spec())

	// Bypass the mapper, which would have dropped this with a warning. The point is
	// that the SCHEMA refuses it too, so a hand-edited or third-party document cannot
	// smuggle it past.
	root.Networks.Network[0].Nodes[0].Model = "model-does-not-exist"

	out, ok := validate(t, writeDoc(t, root))

	assert.False(t, ok, "a device referencing a nonexistent model must be rejected")
	assert.Contains(t, out, "leafref", "rejection should cite the leafref:\n%s", out)
}

// TestSchemaAllowsDanglingLinkEndpoint_ByDesign pins a limit of the standard that is
// easy to over-claim, and that I did over-claim before writing this test.
//
// RFC 8345 sets `require-instance false` on ALL FOUR link-endpoint leafrefs
// (source-node, source-tp, dest-node, dest-tp -- ietf-network-topology.yang:161-198).
// That is deliberate: a topology may legitimately reference nodes it does not itself
// hold, e.g. one that lives in an underlay layer. So the schema does NOT catch a link
// pointing at a device that does not exist.
//
// The guard is therefore the mapper's, not the schema's: FromEntities resolves both
// endpoints and drops the connection with a warning if either fails
// (TestConnectionWithUnresolvableEndpointIsDropped). This test exists so that
// distinction stays explicit -- and so nobody "fixes" a failing validation by weakening
// the mapper's check, believing the schema is backing them up. It is not.
func TestSchemaAllowsDanglingLinkEndpoint_ByDesign(t *testing.T) {
	root, _ := mapping.FromEntities(spec())

	root.Networks.Network[0].Links[0].Source.SourceNode = "urn:nsl:device:does-not-exist"

	_, ok := validate(t, writeDoc(t, root))

	assert.True(t, ok,
		"RFC 8345 sets require-instance false on link endpoints, so a dangling endpoint "+
			"validates. If this ever starts FAILING, the standard modules changed and the "+
			"mapper's endpoint check could be relaxed.")
}
