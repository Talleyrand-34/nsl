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
package api

import (
	"fmt"
	"log"

	q "nsl-graph/internal/repository/application"
	infra "nsl-graph/internal/repository/infra/cloverdb/base"
	"nsl-graph/internal/repository/plugins"
	"nsl-graph/internal/repository/plugins/builtin"
)

func serviceConnection(path string) (q.NetServiceInt, error) {
	baseRepo, err := infra.NewCloverRepository(path)
	if err != nil {
		return nil, fmt.Errorf("Error creating repository: %w", err)
	}

	// Load plugin configuration
	pluginConfig, err := plugins.LoadConfig("plugins.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load plugin config: %v. Using defaults.", err)
		pluginConfig, _ = plugins.LoadConfig("") // Get default config
	}

	// Create and configure plugin registry
	registry := plugins.NewRegistry()

	// Register all built-in plugins
	registry.RegisterSorter(builtin.NewInsertionOrderSorter())
	registry.RegisterSorter(builtin.NewZoneNameSorter())
	registry.RegisterSorter(builtin.NewDeviceNameSorter())
	registry.RegisterSorter(builtin.NewReverseIDSorter())

	// Set active sorter from configuration
	if err := registry.SetActiveSorter(pluginConfig.Plugins.ConnectionSorters.Active); err != nil {
		log.Printf("Warning: Failed to set active sorter '%s': %v. Using insertion_order.",
			pluginConfig.Plugins.ConnectionSorters.Active, err)
		registry.SetActiveSorter("insertion_order")
	}

	// Initialize global plugin manager for runtime configuration
	plugins.InitializeGlobalPluginManager(registry)

	// Wrap repository with plugin decorator
	pluginRepo := plugins.NewPluginAwareRepository(baseRepo, registry)

	// Create service with plugin-aware repository
	service := q.NewNetService(pluginRepo)
	return service, nil
}
