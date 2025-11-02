/*
Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)

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
package builtin

import (
	"sort"

	e "nsl-graph/internal/repository/entities"
)

// ZoneNameSorter sorts connections by zone name
type ZoneNameSorter struct{}

// NewZoneNameSorter creates a new zone name sorter plugin
func NewZoneNameSorter() *ZoneNameSorter {
	return &ZoneNameSorter{}
}

func (z *ZoneNameSorter) ID() string {
	return "zone_name"
}

func (z *ZoneNameSorter) Name() string {
	return "Sort by Zone Name"
}

func (z *ZoneNameSorter) Description() string {
	return "Sorts connections by: FromZoneName (primary), ToZoneName (secondary), FromDevice (tertiary)"
}

func (z *ZoneNameSorter) SortConnections(connections []e.Connection) []e.Connection {
	// Create a copy to avoid modifying the original slice
	sorted := make([]e.Connection, len(connections))
	copy(sorted, connections)

	sort.SliceStable(sorted, func(i, j int) bool {
		// Primary: Sort by FromZoneName
		if sorted[i].FromZoneName != sorted[j].FromZoneName {
			return sorted[i].FromZoneName < sorted[j].FromZoneName
		}

		// Secondary: Sort by ToZoneName
		if sorted[i].ToZoneName != sorted[j].ToZoneName {
			return sorted[i].ToZoneName < sorted[j].ToZoneName
		}

		// Tertiary: Sort by FromDevice
		return sorted[i].FromDevice < sorted[j].FromDevice
	})

	return sorted
}
