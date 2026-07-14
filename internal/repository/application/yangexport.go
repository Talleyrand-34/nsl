// yangexport.go: exports the specification as RFC 7951 JSON, conforming to the
// YANG modules in internal/yang/modules.
package application

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
	"fmt"

	"nsl-graph/internal/yang/mapping"
)

// ExportYANG renders the whole specification as RFC 7951 JSON, valid against
// ietf-network / ietf-network-topology / ietf-l2-topology / nsl-topology /
// nsl-inventory.
//
// The returned warnings report values the domain model holds but the schema will
// not accept -- an out-of-range VLAN, a reference to an owner nobody created. They
// are dropped from the output rather than failing the export, so the document
// always validates and the warnings are a report of what the hand-rolled schema has
// been quietly tolerating. See internal/yang/mapping.
//
// It deliberately does NOT go through ExportAllStructs: that returns entities.All,
// a stale SQL-era projection which omits device interfaces, IPs, MACs and port VLAN
// configs, always reports Policies as empty, and carries int64/-1 foreign keys that
// no longer resolve against CloverDB's string IDs. Composing the resolved getters
// instead costs one more call each and loses nothing.
func (ns *NetService) ExportYANG() ([]byte, []string, error) {
	src := mapping.Source{}

	// Each getter is fatal: a partial export would silently produce a document that
	// validates but describes a network that does not exist.
	for _, load := range []struct {
		what string
		fn   func() error
	}{
		{"brands", func() (err error) { src.Brands, err = ns.GetBrands(); return }},
		{"model types", func() (err error) { src.ModelTypes, err = ns.GetModelTypes(); return }},
		{"os types", func() (err error) { src.OsTypes, err = ns.GetOsTypes(); return }},
		{"owners", func() (err error) { src.Owners, err = ns.GetOwners(); return }},
		{"zone types", func() (err error) { src.ZoneTypes, err = ns.GetZonetypes(); return }},
		{"zones", func() (err error) { src.Zones, err = ns.GetZones(); return }},
		{"models", func() (err error) { src.Models, err = ns.GetModels(); return }},
		{"model ports", func() (err error) { src.ModelPorts, err = ns.GetModelPorts(); return }},
		{"devices", func() (err error) { src.Devices, err = ns.GetDevices(); return }},
		{"device ports", func() (err error) { src.DevicePorts, err = ns.GetDevicePorts(); return }},
		{"connections", func() (err error) { src.Connections, err = ns.GetConnections(); return }},
	} {
		if err := load.fn(); err != nil {
			return nil, nil, fmt.Errorf("loading %s: %w", load.what, err)
		}
	}

	root, warnings := mapping.FromEntities(src)

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, warnings, fmt.Errorf("encoding RFC 7951 JSON: %w", err)
	}
	return out, warnings, nil
}
