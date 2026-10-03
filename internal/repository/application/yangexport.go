// yangexport.go: exports the specification as RFC 7951 JSON, conforming to the
// YANG modules in internal/yang/modules.
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
	"encoding/json"
	"fmt"
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
	// loadSpec is shared with the diff: an export and an audit must be looking at the
	// same specification, or the audit is auditing something else.
	root, warnings, err := ns.IntendedTree()
	if err != nil {
		return nil, nil, err
	}

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, warnings, fmt.Errorf("encoding RFC 7951 JSON: %w", err)
	}
	return out, warnings, nil
}
