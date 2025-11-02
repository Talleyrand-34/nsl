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
	"sync"
)

// GlobalPluginManager is a singleton instance that manages the plugin registry
// across the entire application. It provides thread-safe access to the plugin
// system for runtime configuration changes.
var GlobalPluginManager *PluginManager

// PluginManager provides thread-safe access to the plugin registry
// and allows runtime configuration changes without restarting the application.
type PluginManager struct {
	registry *Registry
	mu       sync.RWMutex
}

// NewPluginManager creates a new plugin manager with the given registry
func NewPluginManager(registry *Registry) *PluginManager {
	return &PluginManager{
		registry: registry,
	}
}

// GetRegistry returns the current plugin registry
func (pm *PluginManager) GetRegistry() *Registry {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.registry
}

// SetActiveSorter changes the active sorting plugin at runtime
func (pm *PluginManager) SetActiveSorter(pluginID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.registry.SetActiveSorter(pluginID)
}

// GetActiveSorter returns the currently active sorter plugin
func (pm *PluginManager) GetActiveSorter() ConnectionSorterPlugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.registry.GetActiveSorter()
}

// ListAvailableSorters returns a list of all registered sorter plugin IDs
func (pm *PluginManager) ListAvailableSorters() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.registry.ListAvailableSorters()
}

// GetSorter returns a specific sorter plugin by ID
func (pm *PluginManager) GetSorter(pluginID string) ConnectionSorterPlugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.registry.GetSorter(pluginID)
}

// InitializeGlobalPluginManager initializes the global plugin manager singleton.
// This should be called once during application startup.
func InitializeGlobalPluginManager(registry *Registry) {
	GlobalPluginManager = NewPluginManager(registry)
}
