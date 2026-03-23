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
package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	q "nsl-graph/internal/repository/application"
	infra "nsl-graph/internal/repository/infra/cloverdb/base"
	"nsl-graph/internal/repository/plugins"
	"nsl-graph/internal/repository/plugins/builtin"

	"nsl-graph/internal/api/core"
	"nsl-graph/internal/api/devices"
	"nsl-graph/internal/api/connections"
	"nsl-graph/internal/api/vlans"
	apiplugins "nsl-graph/internal/api/plugins"
	"nsl-graph/internal/api/scanning"
)

func StartServer(dbPath string, port int) {
	// Open DB connection ONCE
	service, err := serviceConnection(dbPath)
	if err != nil {
		panic(fmt.Errorf("failed to create service: %w", err))
	}

	r := mux.NewRouter()
	registerAllRoutes(r, service)

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start server in a goroutine so we can listen for signals
	go func() {
		fmt.Printf("Starting server on %v\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Wait for SIGINT or SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	fmt.Printf("Received signal %s, shutting down...\n", sig)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("HTTP server shutdown error: %v\n", err)
	} else {
		fmt.Println("HTTP server gracefully stopped.")
	}
}

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

// registerAllRoutes registers all API routes from different packages
func registerAllRoutes(r *mux.Router, service q.NetServiceInt) {
	// Core routes (root, diagram)
	core.RegisterRoutes(r, service)

	// Device-related routes (brands, models, devices, etc.)
	devices.RegisterRoutes(r, service)

	// Connection-related routes
	connections.RegisterRoutes(r, service)

	// VLAN-related routes
	vlans.RegisterRoutes(r, service)

	// Plugin management routes
	apiplugins.RegisterRoutes(r)

	// Network scanning routes
	registerScanningRoutes(r, service)
}

// registerScanningRoutes registers network scanning endpoints
func registerScanningRoutes(r *mux.Router, service q.NetServiceInt) {
	// Network scanning endpoints
	r.HandleFunc("/scan/network", scanning.ScanNetworkHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/host", scanning.ScanHostHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/import", scanning.ImportDevicesHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/status", scanning.GetScanStatusHandler()).Methods("GET", "OPTIONS")
	r.HandleFunc("/scan/validate", scanning.ValidateSubnetHandler()).Methods("GET", "OPTIONS")
}
