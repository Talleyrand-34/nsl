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
package cmd_utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	c "nsl-graph/cmd"
	q "nsl-graph/internal/repository/application"
	"nsl-graph/internal/repository/plugins"
	"nsl-graph/internal/repository/plugins/builtin"

	// infra "nsl-graph/internal/repository/infra/manual_cloverdb/base"

	infra "nsl-graph/internal/repository/infra/cloverdb/base"
	// sqliteInfra "nsl-graph/internal/repository/infra/sqlc_sqlite/base"
)

// openDatabaseConnection establishes and returns a database connection.
func OpenDatabaseConnectionSqlite() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", c.Srcdbpath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}
	return db, nil
}
func OpenDatabaseConnectionClover() {
	os.Mkdir(c.Srcdbpath, 0755)
}

// Return repositoryDB connection using global config
func ServiceConnection() (q.NetServiceInt, error) {
	return GetServiceConnection(c.Srcdbpath)
}

// Return repositoryDB connection with custom database path
func GetServiceConnection(dbPath string) (q.NetServiceInt, error) {
	backend := strings.ToLower(c.Backend)

	switch backend {
	case "cloverdb", "clover":
		// Use CloverDB backend
		os.Mkdir(dbPath, 0755)
		baseRepo, err := infra.NewCloverRepository(dbPath)
		if err != nil {
			log.Fatalf("Error connecting to CloverDB: %v", err)
		}

		// Load plugin configuration
		pluginConfig, err := plugins.LoadConfig(c.PluginConfig)
		if err != nil {
			log.Printf("Warning: Failed to load plugin config from '%s': %v. Using defaults.", c.PluginConfig, err)
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

	case "sqlite", "":
		// SQLite backend not yet fully supported
		// Use SQLite backend (default)
		// 66 -      db, err := OpenDatabaseConnectionSqlite()
		// 67 -      if err != nil {
		// 68 -        log.Fatalf("Error connecting to SQLite database: %v", err)
		// 69 -      }
		// 70 -      repository, err := sqliteInfra.NewSQLiteRepositoryFromDB(db)
		// 71 -      if err != nil {
		// 72 -        log.Fatalf("Error initializing SQLite repository: %v", err)
		// 73 -      }
		// 74 -      service := q.NewNetService(repository)
		// 75 -      return service, nil
		return nil, fmt.Errorf("SQLite backend is not yet fully supported - please use 'cloverdb' backend instead")

	default:
		return nil, fmt.Errorf("unsupported backend type: %s (supported: sqlite, cloverdb)", backend)
	}
}

// PrintPrettyJSON prints the JSON data in a human-readable format.
func PrintPrettyJSON(jsonData string) error {
	var prettyData map[string]interface{}

	// Unmarshal the JSON string into a map for formatting
	err := json.Unmarshal([]byte(jsonData), &prettyData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Marshal the map back into a pretty-printed JSON string
	prettyJSON, err := json.MarshalIndent(prettyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(prettyJSON))
	return nil
}

func PrintStringArrayPrettyJson(arr []string) {
	b, _ := json.MarshalIndent(arr, "", "  ")
	fmt.Println(string(b))
}

// flagproc gets the values of the given flag names from the command and returns them as a slice of strings.
func Flagproc(cmd *cobra.Command, flagNames []string) []string {
	values := make([]string, len(flagNames))
	for i, flag := range flagNames {
		val, err := cmd.Flags().GetString(flag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag '%v': %v\n", flag, err)
			os.Exit(1)
		}
		values[i] = val
	}
	return values
}
