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
			"PUT    /brands",

			// /connections
			"DELETE /connections",
			"GET    /connections",
			"POST   /connections",
			"PUT    /connections",

			// /connectiontypes
			"DELETE /connectiontypes",
			"GET    /connectiontypes",
			"POST   /connectiontypes",
			"PUT    /connectiontypes",

			// /deviceclasses
			"DELETE /deviceclasses",
			"GET    /deviceclasses",
			"POST   /deviceclasses",
			"PUT    /deviceclasses",

			// /deviceports
			"DELETE /deviceports",
			"GET    /deviceports",
			"POST   /deviceports",

			// /devices
			"DELETE /devices",
			"GET    /devices",
			"POST   /devices",
			"PUT    /devices",

			// /modelports
			"DELETE /modelports",
			"GET    /modelports",
			"POST   /modelports",
			"POST   /modelports/bulk",
			"PUT    /modelports",

			// /models
			"DELETE /models",
			"GET    /models",
			"POST   /models",
			"PUT    /models",

			// /proprietaries
			"DELETE /proprietaries",
			"GET    /proprietaries",
			"POST   /proprietaries",
			"PUT    /proprietaries",

			// /zones
			"DELETE /zones",
			"GET    /zones",
			"POST   /zones",
			"PUT    /zones",

			// /zonetypes
			"DELETE /zonetypes",
			"GET    /zonetypes",
			"POST   /zonetypes",
			"PUT    /zonetypes",

			// /vlans
			"DELETE /vlans",
			"GET    /vlans",
			"POST   /vlans",
			"PUT    /vlans",

			// /diagram
			"GET    /diagram",
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
	r.HandleFunc("/modelports/bulk", addBulkModelPortHandler(service)).Methods("POST")
	r.HandleFunc("/modelports", addModelPortHandler(service)).Methods("POST")
	r.HandleFunc("/modelports", getModelPortsHandler(service)).Methods("GET")
	// DevicePort
	r.HandleFunc("/deviceports", addDevicePortHandler(service)).Methods("POST")
	r.HandleFunc("/deviceports", getDevicePortsHandler(service)).Methods("GET")
	// Connection
	r.HandleFunc("/connections", addConnectionHandler(service)).Methods("POST")
	r.HandleFunc("/connections", getConnectionsHandler(service)).Methods("GET")
	// ConnectionType
	r.HandleFunc("/connectiontypes", addConnectionTypeHandler(service)).Methods("POST")
	r.HandleFunc("/connectiontypes", getConnectionsTypeHandler(service)).Methods("GET")
	// VLAN
	r.HandleFunc("/vlans", addVlanHandler(service)).Methods("POST")
	r.HandleFunc("/vlans", getVlansHandler(service)).Methods("GET")
	// All Ports for a Device
	r.HandleFunc("/allports/device", getAllPortsDeviceHandler(service)).Methods("GET")
	// All Ports (All Devices)
	r.HandleFunc("/allports/all", getAllPortsAllHandler(service)).Methods("GET")
	r.HandleFunc("/diagram", getDiagram(service)).Methods("GET")
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
	r.HandleFunc("/connections", updateConnectionHandler(service)).Methods("PUT")
	// ConnectionType
	r.HandleFunc("/connectiontypes", deleteConnectionTypeHandler(service)).Methods("DELETE")
	// VLAN
	r.HandleFunc("/vlans", deleteVlanHandler(service)).Methods("DELETE")

	// ADD ALL MISSING UPDATE ROUTES
	// Brand
	r.HandleFunc("/brands", updateBrandHandler(service)).Methods("PUT")
	// DeviceClass
	r.HandleFunc("/deviceclasses", updateDeviceClassHandler(service)).Methods("PUT")
	// ZoneType
	r.HandleFunc("/zonetypes", updateZoneTypeHandler(service)).Methods("PUT")
	// Proprietary
	r.HandleFunc("/proprietaries", updateProprietaryHandler(service)).Methods("PUT")
	// Zone
	r.HandleFunc("/zones", updateZoneHandler(service)).Methods("PUT")
	// Model
	r.HandleFunc("/models", updateModelHandler(service)).Methods("PUT")
	// Device
	r.HandleFunc("/devices", updateDeviceHandler(service)).Methods("PUT")
	// ModelPort
	r.HandleFunc("/modelports", updateModelPortHandler(service)).Methods("PUT")
	// ConnectionType
	r.HandleFunc("/connectiontypes", updateConnectionTypeHandler(service)).Methods("PUT")
	// VLAN
	r.HandleFunc("/vlans", updateVlanHandler(service)).Methods("PUT")
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
