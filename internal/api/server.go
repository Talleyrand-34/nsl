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
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	q "nsl-graph/internal/repository/application"
	infra "nsl-graph/internal/repository/infra/cloverdb/base"

	"nsl-graph/internal/api/connections"
	"nsl-graph/internal/api/core"
	"nsl-graph/internal/api/devices"
	"nsl-graph/internal/api/scanning"
	"nsl-graph/internal/api/vlans"
	"nsl-graph/internal/observ"
)

func StartServer(dbPath string, port int) {
	// Open DB connection ONCE
	service, err := serviceConnection(dbPath)
	if err != nil {
		panic(fmt.Errorf("failed to create service: %w", err))
	}

	r := mux.NewRouter()
	// Observability middleware wraps every route: panic recovery + a structured
	// request log line per call.
	r.Use(observ.Recover, observ.RequestLogger)
	registerAllRoutes(r, service)

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start server in a goroutine so we can listen for signals
	go func() {
		slog.Info("starting HTTP server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	// Wait for SIGINT or SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	slog.Info("shutting down", "signal", sig.String())

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	} else {
		slog.Info("HTTP server gracefully stopped")
	}
}

func serviceConnection(path string) (q.NetServiceInt, error) {
	baseRepo, err := infra.NewCloverRepository(path)
	if err != nil {
		return nil, fmt.Errorf("Error creating repository: %w", err)
	}

	// Create service with base repository
	service := q.NewNetService(baseRepo)
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

	// Network scanning routes
	registerScanningRoutes(r, service)
}

// registerScanningRoutes registers network scanning endpoints
func registerScanningRoutes(r *mux.Router, service q.NetServiceInt) {
	// Network scanning endpoints
	r.HandleFunc("/scan/run", scanning.ScanRunHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/network", scanning.ScanNetworkHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/host", scanning.ScanHostHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/host-ssh", scanning.ScanHostSSHHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/analyze", scanning.AnalyzeDeviceHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/execute", scanning.ExecuteImportPlanHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/import", scanning.ImportDevicesHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/import-file", scanning.ImportScanFileHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/connections", scanning.ScanConnectionsHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/connections/import", scanning.ImportConnectionsHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan/profiles", scanning.ScanProfilesHandler(service)).Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")
	r.HandleFunc("/scan/status", scanning.GetScanStatusHandler()).Methods("GET", "OPTIONS")
	r.HandleFunc("/scan/validate", scanning.ValidateSubnetHandler()).Methods("GET", "OPTIONS")
}
