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

// DeviceNameSorter sorts connections by device name
type DeviceNameSorter struct{}

// NewDeviceNameSorter creates a new device name sorter plugin
func NewDeviceNameSorter() *DeviceNameSorter {
	return &DeviceNameSorter{}
}

func (d *DeviceNameSorter) ID() string {
	return "device_name"
}

func (d *DeviceNameSorter) Name() string {
	return "Sort by Device Name"
}

func (d *DeviceNameSorter) Description() string {
	return "Sorts connections by: FromDevice (primary), ToDevice (secondary), FromModelPort (tertiary)"
}

func (d *DeviceNameSorter) SortConnections(connections []e.Connection) []e.Connection {
	// Create a copy to avoid modifying the original slice
	sorted := make([]e.Connection, len(connections))
	copy(sorted, connections)

	sort.SliceStable(sorted, func(i, j int) bool {
		// Primary: Sort by FromDevice
		if sorted[i].FromDevice != sorted[j].FromDevice {
			return sorted[i].FromDevice < sorted[j].FromDevice
		}

		// Secondary: Sort by ToDevice
		if sorted[i].ToDevice != sorted[j].ToDevice {
			return sorted[i].ToDevice < sorted[j].ToDevice
		}

		// Tertiary: Sort by FromModelPort
		return sorted[i].FromModelPort < sorted[j].FromModelPort
	})

	return sorted
}
