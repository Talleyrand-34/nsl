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
package plugins

import (
	"fmt"

	e "nsl-graph/internal/repository/entities"
)

// Registry manages the collection of registered plugins and
// determines which plugin is currently active
type Registry struct {
	// Map of plugin ID to plugin instance
	sorters map[string]ConnectionSorterPlugin

	// The currently active sorter plugin
	activeSorter ConnectionSorterPlugin
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		sorters: make(map[string]ConnectionSorterPlugin),
	}
}

// RegisterSorter adds a connection sorter plugin to the registry
func (r *Registry) RegisterSorter(plugin ConnectionSorterPlugin) error {
	if plugin == nil {
		return fmt.Errorf("cannot register nil plugin")
	}

	id := plugin.ID()
	if id == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	if _, exists := r.sorters[id]; exists {
		return fmt.Errorf("plugin with ID '%s' is already registered", id)
	}

	r.sorters[id] = plugin
	return nil
}

// SetActiveSorter selects the active sorting plugin by ID.
// Returns an error if the plugin is not registered.
func (r *Registry) SetActiveSorter(pluginID string) error {
	plugin, exists := r.sorters[pluginID]
	if !exists {
		return fmt.Errorf("plugin with ID '%s' is not registered", pluginID)
	}

	r.activeSorter = plugin
	return nil
}

// GetActiveSorter returns the currently active sorter plugin.
// Returns nil if no plugin is active.
func (r *Registry) GetActiveSorter() ConnectionSorterPlugin {
	return r.activeSorter
}

// ApplySorting applies the active sorter plugin to connections.
// If no sorter is active, returns connections unchanged.
func (r *Registry) ApplySorting(connections []e.Connection) []e.Connection {
	if r.activeSorter == nil {
		// No active sorter, return original order
		return connections
	}

	return r.activeSorter.SortConnections(connections)
}

// ListAvailableSorters returns a list of all registered sorter plugin IDs
func (r *Registry) ListAvailableSorters() []string {
	ids := make([]string, 0, len(r.sorters))
	for id := range r.sorters {
		ids = append(ids, id)
	}
	return ids
}

// GetSorter returns a specific sorter plugin by ID.
// Returns nil if the plugin doesn't exist.
func (r *Registry) GetSorter(pluginID string) ConnectionSorterPlugin {
	return r.sorters[pluginID]
}
