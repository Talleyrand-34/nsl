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
			"GET    /brands",
			"POST   /brands",
			"GET    /deviceclasses",
			"POST   /deviceclasses",
			"GET    /zonetypes",
			"POST   /zonetypes",
			"GET    /proprietaries",
			"POST   /proprietaries",
			"GET    /zones",
			"POST   /zones",
			"GET    /models",
			"POST   /models",
			"GET    /devices",
			"POST   /devices",
			"GET    /modelports",
			"POST   /modelports",
			"GET    /deviceports",
			"POST   /deviceports",
			"GET    /connections",
			"POST   /connections",
			"GET    /allports/device?deviceid=<id>",
			"GET    /allports/all",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(endpoints)
	}
}

// --- Brand ---
func addBrandHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Brand string `json:"brand"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.AddBrand(req.Brand); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getBrandsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		brands, err := service.GetBrands()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(brands)
	}
}

// --- DeviceClass ---
func addDeviceClassHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Brand string `json:"brand"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.AddDeviceClass(req.Brand); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getDeviceClassesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		classes, err := service.GetDeviceClasses()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(classes)
	}
}

// --- ZoneType ---
func addZoneTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.AddZoneType(req.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getZoneTypesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		types, err := service.GetZonetypes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(types)
	}
}

// --- Proprietary ---
func addProprietaryHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.AddProprietary(req.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getProprietariesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		props, err := service.GetProperties()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(props)
	}
}

// --- Zone ---
func addZoneHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name        string `json:"name"`
			Father      string `json:"father"`
			FatherID    string `json:"fatherid"`
			Proprietary string `json:"proprietary"`
			ZoneName    string `json:"zonename"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.AddZone(req.Name, req.Father, req.FatherID, req.Proprietary, req.ZoneName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getZonesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zones, err := service.GetZones()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(zones)
	}
}

// --- Model ---
func addModelHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ModelName string `json:"modelName"`
			BrandName string `json:"brandName"`
			ClassName string `json:"className"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.AddModel(req.ModelName, req.BrandName, req.ClassName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getModelsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models, err := service.GetModels()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(models)
	}
}

// --- Device ---
func addDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Label       string `json:"label"`
			Model       string `json:"model"`
			ZoneId      string `json:"zoneId"`
			ZoneName    string `json:"zoneName"`
			Proprietary string `json:"proprietary"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.AddDevice(req.Label, req.Model, req.ZoneId, req.ZoneName, req.Proprietary); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getDevicesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devices, err := service.GetDevices()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(devices)
	}
}

// --- ModelPort ---
func addModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name      string `json:"name"`
			PosX      string `json:"posx"`
			PosY      string `json:"posy"`
			ModelName string `json:"modelName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.AddModelPort(req.Name, req.PosX, req.PosY, req.ModelName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getModelPortsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ports, err := service.GetModelPorts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(ports)
	}
}

// --- DevicePort ---
func addDevicePortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			DeviceID    string `json:"deviceid"`
			ModelPortID string `json:"modelportid"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.AddDevicePort(req.DeviceID, req.ModelPortID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getDevicePortsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ports, err := service.GetDevicePorts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(ports)
	}
}

// --- Connection ---
func addConnectionHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			FromDevice    string `json:"fromDevice"`
			FromModelPort string `json:"fromModelPort"`
			ToDevice      string `json:"toDevice"`
			ToModelPort   string `json:"toModelPort"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.AddConnection(req.FromDevice, req.FromModelPort, req.ToDevice, req.ToModelPort); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getConnectionsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conns, err := service.GetConnections()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(conns)
	}
}

// --- All Ports for Device ---
func getAllPortsDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("deviceid")
		if deviceID == "" {
			http.Error(w, "deviceid is required", http.StatusBadRequest)
			return
		}
		ports, err := service.GetAllPortsDevice(deviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(ports)
	}
}

// --- All Ports (All Devices) ---
func getAllPortsAllHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ports, err := service.GetAllPortsAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(ports)
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
