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
package plugins

import (
	e "nsl-graph/internal/repository/entities"
)

// ConnectionSorterPlugin defines the interface for connection sorting plugins.
// Plugins implement custom sorting logic for connections without modifying
// the underlying database implementation.
type ConnectionSorterPlugin interface {
	// ID returns a unique identifier for the plugin (e.g., "zone_name", "device_name")
	// This ID is used in configuration files to select the active plugin.
	ID() string

	// Name returns a human-readable name for the plugin
	Name() string

	// Description returns a detailed description of the sorting behavior
	Description() string

	// SortConnections applies the sorting logic to a slice of connections.
	// The original slice is not modified; a new sorted slice is returned.
	SortConnections(connections []e.Connection) []e.Connection
}
