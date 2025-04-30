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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	q "nsl-graph/internal/repository/application"
)

func rootHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpoints := []string{
			// /allports/all and /allports/device
			"GET    /allports/all",
			"GET    /allports/device?deviceid=<id>",

			// /brands
			"DELETE /brands",
			"GET    /brands",
			"POST   /brands",

			// /connections
			"DELETE /connections",
			"GET    /connections",
			"POST   /connections",

			// /deviceclasses
			"DELETE /deviceclasses",
			"GET    /deviceclasses",
			"POST   /deviceclasses",

			// /deviceports
			"DELETE /deviceports",
			"GET    /deviceports",
			"POST   /deviceports",

			// /devices
			"DELETE /devices",
			"GET    /devices",
			"POST   /devices",

			// /modelports
			"DELETE /modelports",
			"GET    /modelports",
			"POST   /modelports",

			// /models
			"DELETE /models",
			"GET    /models",
			"POST   /models",

			// /proprietaries
			"DELETE /proprietaries",
			"GET    /proprietaries",
			"POST   /proprietaries",

			// /zones
			"DELETE /zones",
			"GET    /zones",
			"POST   /zones",

			// /zonetypes
			"DELETE /zonetypes",
			"GET    /zonetypes",
			"POST   /zonetypes",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(endpoints)
	}
}

// // --- Export All Structs ---
// func exportAllStructsHandler(service q.NetServiceInt) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		data := service.ExportAllStructs()
// 		w.Header().Set("Content-Type", "application/octet-stream")
// 		w.Write(data)
// 	}
// }

// --- Register All Routes ---
func RegisterRoutes(r *mux.Router, service q.NetServiceInt) {
	r.HandleFunc("/", rootHandler()).Methods("GET")
	// Brand
	r.HandleFunc("/brands", addBrandHandler(service)).Methods("POST")
	r.HandleFunc("/brands", getBrandsHandler(service)).Methods("GET")
	// DeviceClass
	r.HandleFunc("/deviceclasses", addDeviceClassHandler(service)).Methods("POST")
	r.HandleFunc("/deviceclasses", getDeviceClassesHandler(service)).Methods("GET")
	// ZoneType
	r.HandleFunc("/zonetypes", addZoneTypeHandler(service)).Methods("POST")
	r.HandleFunc("/zonetypes", getZoneTypesHandler(service)).Methods("GET")
	// Proprietary
	r.HandleFunc("/proprietaries", addProprietaryHandler(service)).Methods("POST")
	r.HandleFunc("/proprietaries", getProprietariesHandler(service)).Methods("GET")
	// Zone
	r.HandleFunc("/zones", addZoneHandler(service)).Methods("POST")
	r.HandleFunc("/zones", getZonesHandler(service)).Methods("GET")
	// Model
	r.HandleFunc("/models", addModelHandler(service)).Methods("POST")
	r.HandleFunc("/models", getModelsHandler(service)).Methods("GET")
	// Device
	r.HandleFunc("/devices", addDeviceHandler(service)).Methods("POST")
	r.HandleFunc("/devices", getDevicesHandler(service)).Methods("GET")
	// ModelPort
	r.HandleFunc("/modelports", addModelPortHandler(service)).Methods("POST")
	r.HandleFunc("/modelports", getModelPortsHandler(service)).Methods("GET")
	// DevicePort
	r.HandleFunc("/deviceports", addDevicePortHandler(service)).Methods("POST")
	r.HandleFunc("/deviceports", getDevicePortsHandler(service)).Methods("GET")
	// Connection
	r.HandleFunc("/connections", addConnectionHandler(service)).Methods("POST")
	r.HandleFunc("/connections", getConnectionsHandler(service)).Methods("GET")
	// All Ports for a Device
	r.HandleFunc("/allports/device", getAllPortsDeviceHandler(service)).Methods("GET")
	// All Ports (All Devices)
	r.HandleFunc("/allports/all", getAllPortsAllHandler(service)).Methods("GET")
	// Export
	// r.HandleFunc("/export", exportAllStructsHandler(service)).Methods("GET")
	// Brand
	r.HandleFunc("/brands", deleteBrandHandler(service)).Methods("DELETE")
	// DeviceClass
	r.HandleFunc("/deviceclasses", deleteDeviceClassHandler(service)).Methods("DELETE")
	// ZoneType
	r.HandleFunc("/zonetypes", deleteZoneTypeHandler(service)).Methods("DELETE")
	// Proprietary
	r.HandleFunc("/proprietaries", deleteProprietaryHandler(service)).Methods("DELETE")
	// Zone
	r.HandleFunc("/zones", deleteZoneHandler(service)).Methods("DELETE")
	// Model
	r.HandleFunc("/models", deleteModelHandler(service)).Methods("DELETE")
	// Device
	r.HandleFunc("/devices", deleteDeviceHandler(service)).Methods("DELETE")
	// ModelPort
	r.HandleFunc("/modelports", deleteModelPortHandler(service)).Methods("DELETE")
	// DevicePort
	r.HandleFunc("/deviceports", deleteDevicePortHandler(service)).Methods("DELETE")
	// Connection
	r.HandleFunc("/connections", deleteConnectionHandler(service)).Methods("DELETE")
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
