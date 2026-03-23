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
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PluginConfig represents the structure of the plugins.yaml configuration file
type PluginConfig struct {
	Plugins PluginsSection `yaml:"plugins"`
}

// PluginsSection contains all plugin configurations
type PluginsSection struct {
	ConnectionSorters ConnectionSortersConfig `yaml:"connection_sorters"`
}

// ConnectionSortersConfig defines connection sorting plugin configuration
type ConnectionSortersConfig struct {
	// Active is the ID of the currently active sorter plugin
	Active string `yaml:"active"`

	// Available lists all available sorter plugins with metadata
	Available []PluginMetadata `yaml:"available"`
}

// PluginMetadata provides information about a plugin
type PluginMetadata struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// LoadConfig loads plugin configuration from a YAML file.
// If the file doesn't exist, returns default configuration (insertion_order).
func LoadConfig(path string) (*PluginConfig, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return default configuration
		return getDefaultConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin config file: %w", err)
	}

	// Parse YAML
	var config PluginConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse plugin config YAML: %w", err)
	}

	// Validate configuration
	if config.Plugins.ConnectionSorters.Active == "" {
		config.Plugins.ConnectionSorters.Active = "insertion_order"
	}

	return &config, nil
}

// getDefaultConfig returns the default plugin configuration
// when no config file is found
func getDefaultConfig() *PluginConfig {
	return &PluginConfig{
		Plugins: PluginsSection{
			ConnectionSorters: ConnectionSortersConfig{
				Active: "insertion_order",
				Available: []PluginMetadata{
					{
						ID:          "insertion_order",
						Name:        "Database Insertion Order (Default)",
						Description: "Original CloverDB order, no sorting applied",
					},
				},
			},
		},
	}
}
