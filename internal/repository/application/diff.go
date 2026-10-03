// diff.go: compare the specification against what a scan actually found.
// SPDX-License-Identifier: AGPL-3.0-or-later
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
	"fmt"

	"nsl-graph/internal/datastore"
	s "nsl-graph/internal/scanner"
	"nsl-graph/internal/yang/canon"
	"nsl-graph/internal/yang/mapping"
)

// IntendedTree renders the specification -- what the network SHOULD be -- as a
// canonical tree.
func (ns *NetService) IntendedTree() (*canon.Root, []string, error) {
	src, err := ns.loadSpec()
	if err != nil {
		return nil, nil, err
	}
	root, warnings := mapping.FromEntities(src)
	return root, warnings, nil
}

// ObservedTree renders scanned devices -- what the network IS -- as a canonical tree,
// resolved against the specification so that the same physical device carries the same
// node-id in both. Without that resolution the two trees would share no identities and
// a diff between them would be meaningless.
func (ns *NetService) ObservedTree(devices []s.SNMPDevice) (*canon.Root, []string, error) {
	src, err := ns.loadSpec()
	if err != nil {
		return nil, nil, err
	}
	root, warnings := mapping.FromScan(devices, src)
	return root, warnings, nil
}

// DiffAgainstScan reports how the network differs from its specification.
//
// This is a config audit: it answers "is the network actually wired and configured the
// way we said it should be?" -- and it is the prerequisite for ever pushing
// configuration, because you cannot safely change a device until you can say what you
// intend versus what is there.
func (ns *NetService) DiffAgainstScan(devices []s.SNMPDevice) ([]datastore.Change, []string, error) {
	intended, w1, err := ns.IntendedTree()
	if err != nil {
		return nil, nil, err
	}
	observed, w2, err := ns.ObservedTree(devices)
	if err != nil {
		return nil, nil, err
	}
	return datastore.Diff(intended, observed), append(w1, w2...), nil
}

// loadSpec reads the whole specification out of the repository.
func (ns *NetService) loadSpec() (mapping.Source, error) {
	src := mapping.Source{}

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
			return src, fmt.Errorf("loading %s: %w", load.what, err)
		}
	}

	return src, nil
}
