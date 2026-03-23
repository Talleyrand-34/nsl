/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)

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
	e "nsl-graph/internal/repository/entities"
)

// InsertionOrderSorter maintains the original database insertion order (no sorting)
type InsertionOrderSorter struct{}

// NewInsertionOrderSorter creates a new insertion order sorter plugin
func NewInsertionOrderSorter() *InsertionOrderSorter {
	return &InsertionOrderSorter{}
}

func (i *InsertionOrderSorter) ID() string {
	return "insertion_order"
}

func (i *InsertionOrderSorter) Name() string {
	return "Database Insertion Order (Default)"
}

func (i *InsertionOrderSorter) Description() string {
	return "Maintains the original CloverDB insertion order without any sorting. This is the default behavior."
}

func (i *InsertionOrderSorter) SortConnections(connections []e.Connection) []e.Connection {
	// No sorting - return connections as-is
	return connections
}
