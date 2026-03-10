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
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	q "nsl-graph/internal/repository/application"
	"nsl-graph/internal/api/core"
	"nsl-graph/internal/api/devices"
	"nsl-graph/internal/api/connections"
	"nsl-graph/internal/api/vlans"
	"nsl-graph/internal/api/plugins"
)

// Legacy comment - handlers now organized by domain in internal/api/

// RegisterRoutes registers all API routes organized by domain
func RegisterRoutes(r *mux.Router, service q.NetServiceInt) {
	// Register domain-specific routes
	core.RegisterRoutes(r, service)
	devices.RegisterRoutes(r, service)
	connections.RegisterRoutes(r, service)
	vlans.RegisterRoutes(r, service)
	plugins.RegisterRoutes(r)
}

func StartServer(dbPath string, port int) {
	// Open DB connection ONCE
	service, err := serviceConnection(dbPath)
	if err != nil {
		panic(fmt.Errorf("failed to create service: %w", err))
	}

	r := mux.NewRouter()
	RegisterRoutes(r, service)

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
